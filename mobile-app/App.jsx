import React, { useState, useEffect, useCallback } from 'react'
import {
  ActivityIndicator, Alert, Image, KeyboardAvoidingView,
  Platform, SafeAreaView, ScrollView, StyleSheet,
  Text, TextInput, TouchableOpacity, View,
} from 'react-native'
import * as ImagePicker from 'expo-image-picker'
import AsyncStorage from '@react-native-async-storage/async-storage'
import { StatusBar } from 'expo-status-bar'

// ── Configuración ─────────────────────────────────────────────────────────────
const DEFAULT_BACKEND = process.env.EXPO_PUBLIC_BACKEND_URL || 'http://192.168.100.31:8080'
const ANTHROPIC_KEY   = process.env.EXPO_PUBLIC_ANTHROPIC_API_KEY || ''

// ── Prompt OCR (idéntico al web scanner) ──────────────────────────────────────
const OCR_PROMPT = `Analiza esta acta electoral boliviana. Extrae los datos y detecta problemas físicos.

REGLAS DE LECTURA:
- Votos en casillas de 3 dígitos: |0|9|5| = 95
- Código de mesa: número largo en esquina superior izquierda
- El ÁREA CENTRAL es el recuadro grande del conteo de votos por candidato

DETECCIÓN DE OBSERVACIONES — responde true/false para cada campo:
- anulada: el documento tiene escrito "ANULADA", "NULA" o sello de anulación visible
- area_central_danada: el recuadro central está mojado, arrugado, roto o manchado de forma que impide leer los votos
- tiene_letras_en_numeros: hay letras escritas en casillas que deben ser numéricas
- tiene_numeros_dudosos: hay un número sobreescrito o corregido con tachón que no permite distinguir con certeza qué número es
- tiene_tachaduras_no_aclaradas: hay tachaduras sin aclaración en la sección de Observaciones del acta
- solo_bordes_danados: SOLO los bordes o esquinas están mojados/arrugados pero el área central está completamente legible

IMPORTANTE:
- Si solo está mojada una esquina o borde → solo_bordes_danados=true, area_central_danada=false
- Si están mojados/dañados los votos → area_central_danada=true
- Si dice explícitamente "ANULADA" → anulada=true

Devuelve ÚNICAMENTE este JSON (null para campos no legibles):
{
  "codigo_mesa": número largo de mesa,
  "numero_mesa": número corto de mesa,
  "p1": votos Daenerys Targaryen,
  "p2": votos Sansa Stark,
  "p3": votos Robert Baratheon,
  "p4": votos Tyrion Lannister,
  "partido_1": nombre del partido P1,
  "partido_2": nombre del partido P2,
  "partido_3": nombre del partido P3,
  "partido_4": nombre del partido P4,
  "validos": total votos válidos,
  "blancos": votos blancos,
  "nulos": votos nulos,
  "habilitados": electores habilitados,
  "papeletas": papeletas en ánfora,
  "anulada": true/false,
  "observaciones": texto de observaciones del acta o "",
  "area_central_danada": true/false,
  "tiene_letras_en_numeros": true/false,
  "tiene_numeros_dudosos": true/false,
  "tiene_tachaduras_no_aclaradas": true/false,
  "solo_bordes_danados": true/false
}`

// ── Lógica de validación (idéntica al web scanner) ────────────────────────────
function detectarObservaciones(result) {
  const obs = []
  if (result.area_central_danada)
    obs.push({ tipo: 'AREA_CENTRAL_DAÑADA', texto: 'El área central del recuadro de conteo está dañada o ilegible. Causal de observación según Ley 026 Art. 190.' })
  if (result.tiene_letras_en_numeros)
    obs.push({ tipo: 'LETRAS_EN_CASILLAS_NUMÉRICAS', texto: 'Se detectaron letras en casillas numéricas. Guía OEP: números claros en cada casilla.' })
  if (result.tiene_numeros_dudosos)
    obs.push({ tipo: 'NÚMERO_SOBREESCRITO_ILEGIBLE', texto: 'Número corregido con tachón — no puede distinguirse. Revisión humana requerida.' })
  if (result.tiene_tachaduras_no_aclaradas)
    obs.push({ tipo: 'TACHADURA_SIN_ACLARACIÓN', texto: 'Tachaduras sin aclaración en Observaciones. Causal de nulidad según Ley 026.' })
  return obs
}

function validate(d) {
  const alerts = []
  if (d.anulada) alerts.push({ t: 'err', m: 'Acta marcada como ANULADA — no puede ingresar al cómputo.' })
  const pv = [d.p1, d.p2, d.p3, d.p4].filter(v => v != null)
  const sum = pv.reduce((a, b) => a + b, 0)
  if (d.validos != null && pv.length > 0 && sum !== d.validos)
    alerts.push({ t: 'err', m: `Inconsistencia: Σ candidatos (${sum}) ≠ válidos (${d.validos})` })
  if (d.validos != null && d.blancos != null && d.nulos != null && d.papeletas != null) {
    const exp = d.validos + d.blancos + d.nulos
    if (exp !== d.papeletas) alerts.push({ t: 'err', m: `Papeletas (${d.papeletas}) ≠ válidos+blancos+nulos (${exp})` })
  }
  if (d.habilitados != null && d.papeletas != null && d.papeletas > d.habilitados)
    alerts.push({ t: 'warn', m: `Papeletas (${d.papeletas}) > habilitados (${d.habilitados})` })
  if (pv.length === 0) alerts.push({ t: 'err', m: 'No se pudieron extraer votos por candidato' })
  if (!d.anulada && alerts.filter(a => a.t === 'err').length === 0 && pv.length > 0)
    alerts.push({ t: 'ok', m: 'Datos aritméticos consistentes ✓' })
  return alerts
}

// ── Claude Vision OCR ─────────────────────────────────────────────────────────
async function callClaudeOCR(base64, mediaType = 'image/jpeg') {
  const resp = await fetch('https://api.anthropic.com/v1/messages', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'x-api-key': ANTHROPIC_KEY,
      'anthropic-version': '2023-06-01',
      'anthropic-dangerous-direct-browser-access': 'true',
    },
    body: JSON.stringify({
      model: 'claude-haiku-4-5-20251001',
      max_tokens: 1024,
      messages: [{ role: 'user', content: [
        { type: 'image', source: { type: 'base64', media_type: mediaType, data: base64 } },
        { type: 'text', text: OCR_PROMPT },
      ]}],
    }),
  })
  const raw = await resp.json()
  if (!resp.ok) throw new Error(raw.error?.message || `Anthropic error ${resp.status}`)
  const text = raw.content?.[0]?.text || ''
  const match = text.match(/\{[\s\S]*\}/)
  if (!match) throw new Error('Claude no devolvió JSON válido: ' + text.substring(0, 200))
  return JSON.parse(match[0])
}

// ── App ───────────────────────────────────────────────────────────────────────
export default function App() {
  const [screen, setScreen] = useState('loading')
  const [user, setUser] = useState(null)
  const [token, setToken] = useState(null)
  const [backendUrl, setBackendUrl] = useState(DEFAULT_BACKEND)

  useEffect(() => {
    Promise.all([
      AsyncStorage.getItem('user'),
      AsyncStorage.getItem('token'),
      AsyncStorage.getItem('backendUrl'),
    ]).then(([u, t, b]) => {
      if (b) setBackendUrl(b)
      if (u && t) { setUser(JSON.parse(u)); setToken(t); setScreen('scanner') }
      else setScreen('login')
    })
  }, [])

  const handleLogin = useCallback(async (u, t) => {
    setUser(u); setToken(t)
    await AsyncStorage.setItem('user', JSON.stringify(u))
    await AsyncStorage.setItem('token', t)
    setScreen('scanner')
  }, [])

  const handleLogout = useCallback(async () => {
    await AsyncStorage.multiRemove(['user', 'token'])
    setUser(null); setToken(null); setScreen('login')
  }, [])

  const handleSaveBackend = useCallback(async (url) => {
    setBackendUrl(url)
    await AsyncStorage.setItem('backendUrl', url)
  }, [])

  if (screen === 'loading') return (
    <View style={[ss.flex1, ss.center, { backgroundColor: '#1a1a2e' }]}>
      <ActivityIndicator color="#fff" size="large" />
    </View>
  )
  if (screen === 'login') return (
    <LoginScreen backendUrl={backendUrl} onSaveBackend={handleSaveBackend} onLogin={handleLogin} />
  )
  if (screen === 'scanner') return (
    <ScannerScreen backendUrl={backendUrl} token={token} user={user} onLogout={handleLogout} />
  )
  return null
}

// ── LoginScreen ───────────────────────────────────────────────────────────────
function LoginScreen({ backendUrl, onSaveBackend, onLogin }) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [url, setUrl] = useState(backendUrl)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function handleLogin() {
    if (!username.trim() || !password.trim()) { setError('Ingresa usuario y contraseña'); return }
    setLoading(true); setError('')
    try {
      await onSaveBackend(url.trim())
      const resp = await fetch(`${url.trim()}/api/v1/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: username.trim(), password }),
      })
      const data = await resp.json()
      if (!resp.ok) throw new Error(data.error || 'Credenciales incorrectas')
      onLogin(data.user || { usuario: username.trim() }, data.token)
    } catch (e) {
      setError(e.message || 'Error de conexión. Verifica la IP del servidor y la red WiFi.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <SafeAreaView style={[ss.flex1, { backgroundColor: '#1a1a2e' }]}>
      <StatusBar style="light" />
      <KeyboardAvoidingView style={ss.flex1} behavior={Platform.OS === 'ios' ? 'padding' : 'height'}>
        <ScrollView contentContainerStyle={[ss.center, { padding: 24, flexGrow: 1 }]}>
          <Text style={s.loginTitle}>Scanner RRV Bolivia</Text>
          <Text style={s.loginSub}>Claude Vision IA + Normativa OEP</Text>

          <View style={s.loginCard}>
            <Text style={s.inputLabel}>URL del servidor (misma red WiFi)</Text>
            <TextInput
              style={s.input} value={url} onChangeText={setUrl}
              placeholder="http://192.168.X.X:8080" placeholderTextColor="#aaa"
              autoCapitalize="none" keyboardType="url"
            />
            <Text style={s.inputLabel}>Usuario</Text>
            <TextInput
              style={s.input} value={username} onChangeText={setUsername}
              placeholder="admin" placeholderTextColor="#aaa"
              autoCapitalize="none"
            />
            <Text style={s.inputLabel}>Contraseña</Text>
            <TextInput
              style={s.input} value={password} onChangeText={setPassword}
              placeholder="••••••••" placeholderTextColor="#aaa"
              secureTextEntry
            />
            {error ? <Text style={s.errorText}>{error}</Text> : null}
            <TouchableOpacity style={[s.btn, loading && ss.disabled]} onPress={handleLogin} disabled={loading}>
              {loading ? <ActivityIndicator color="#fff" /> : <Text style={s.btnText}>Ingresar</Text>}
            </TouchableOpacity>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  )
}

// ── ScannerScreen ─────────────────────────────────────────────────────────────
function ScannerScreen({ backendUrl, token, user, onLogout }) {
  const [phase, setPhase] = useState('idle') // idle | processing | result | done
  const [imageUri, setImageUri] = useState(null)
  const [imageBase64, setImageBase64] = useState(null)
  const [progress, setProgress] = useState({ pct: 0, text: '' })
  const [fields, setFields] = useState(null)
  const [alerts, setAlerts] = useState([])
  const [observaciones, setObservaciones] = useState([])
  const [soloBordes, setSoloBordes] = useState(false)
  const [duplicado, setDuplicado] = useState(null)
  const [decision, setDecision] = useState(null)
  const [showRejectInput, setShowRejectInput] = useState(false)
  const [rejectReason, setRejectReason] = useState('')
  const [saving, setSaving] = useState(false)

  function resetAll() {
    setPhase('idle'); setImageUri(null); setImageBase64(null)
    setFields(null); setAlerts([]); setObservaciones([]); setSoloBordes(false)
    setDuplicado(null); setDecision(null); setShowRejectInput(false); setRejectReason('')
  }

  async function pickImage(useCamera) {
    const perm = useCamera
      ? await ImagePicker.requestCameraPermissionsAsync()
      : await ImagePicker.requestMediaLibraryPermissionsAsync()
    if (!perm.granted) { Alert.alert('Permiso denegado', 'Necesitas otorgar permiso para continuar.'); return }

    const result = useCamera
      ? await ImagePicker.launchCameraAsync({ mediaTypes: 'images', quality: 0.75, base64: true })
      : await ImagePicker.launchImageLibraryAsync({ mediaTypes: 'images', quality: 0.75, base64: true })

    if (!result.canceled && result.assets?.[0]) {
      const asset = result.assets[0]
      setImageUri(asset.uri)
      setImageBase64(asset.base64)
      resetAll()
      setPhase('idle')
    }
  }

  async function handleProcess() {
    if (!imageBase64) return
    setPhase('processing')
    setProgress({ pct: 10, text: 'Analizando con Claude Vision...' })
    try {
      const result = await callClaudeOCR(imageBase64)
      if (!result) throw new Error('No se obtuvo respuesta del OCR')

      setProgress({ pct: 60, text: 'Verificando duplicados...' })
      setDuplicado(null)
      if (result.codigo_mesa) {
        try {
          const dup = await fetch(`${backendUrl}/api/v1/rrv/scan/check?mesa=${result.codigo_mesa}`)
          const dupData = await dup.json()
          if (dupData.existe) setDuplicado(dupData)
        } catch (e) { /* si el backend no responde, continuar */ }
      }

      setProgress({ pct: 80, text: 'Detectando observaciones OEP...' })
      const obs = detectarObservaciones(result)
      setObservaciones(obs)
      setSoloBordes(!!result.solo_bordes_danados && !result.area_central_danada)

      setProgress({ pct: 95, text: 'Validando datos...' })
      setFields(result)
      setAlerts(validate(result))
      setPhase('result')
    } catch (e) {
      setPhase('idle')
      Alert.alert('Error OCR', e.message || 'No se pudo procesar el acta')
    }
  }

  async function handleDecision(type) {
    if (type === 'rechazada' && !showRejectInput) { setShowRejectInput(true); return }
    if (type === 'rechazada' && !rejectReason.trim()) { Alert.alert('', 'Ingresa la razón del rechazo'); return }
    if (type === 'aceptada' && duplicado) { Alert.alert('Acta duplicada', 'Ya existe un acta para esta mesa. Marca Observada o Rechazada.'); return }

    setSaving(true)
    try {
      const headers = { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` }
      const base = (f, def) => f != null ? Number(f) : def

      if (type === 'aceptada') {
        const resp = await fetch(`${backendUrl}/api/v1/rrv/scan`, {
          method: 'POST', headers,
          body: JSON.stringify({
            codigo_mesa: base(fields.codigo_mesa, 0),
            numero_mesa: base(fields.numero_mesa, 0),
            p1: base(fields.p1, 0), p2: base(fields.p2, 0),
            p3: base(fields.p3, 0), p4: base(fields.p4, 0),
            partido_1: fields.partido_1 || null, partido_2: fields.partido_2 || null,
            partido_3: fields.partido_3 || null, partido_4: fields.partido_4 || null,
            validos: base(fields.validos, 0), blancos: base(fields.blancos, 0),
            nulos: base(fields.nulos, 0), habilitados: base(fields.habilitados, 0),
            papeletas: base(fields.papeletas, 0),
            fuente_scan: 'mobile_expo',
          }),
        })
        if (!resp.ok) {
          const err = await resp.json().catch(() => ({}))
          throw new Error(err.error || `Error ${resp.status}`)
        }
      } else if (type === 'observada' || type === 'anulada') {
        const obsPayload = type === 'anulada'
          ? [{ tipo: 'ACTA_ANULADA', descripcion: 'El acta tiene marcación explícita de ANULADA.' }]
          : observaciones.map(o => ({ tipo: o.tipo, descripcion: o.texto }))
        await fetch(`${backendUrl}/api/v1/rrv/scan/observada`, {
          method: 'POST', headers,
          body: JSON.stringify({
            codigo_mesa: base(fields.codigo_mesa, 0),
            fuente_scan: 'mobile_expo',
            anulada: type === 'anulada',
            razon_principal: obsPayload[0]?.descripcion || 'Acta observada',
            observaciones_detectadas: obsPayload,
          }),
        }).catch(() => {})
      } else {
        await fetch(`${backendUrl}/api/v1/rrv/scan/rechazada`, {
          method: 'POST', headers,
          body: JSON.stringify({
            codigo_mesa: base(fields.codigo_mesa, 0),
            p1: base(fields.p1, 0), p2: base(fields.p2, 0),
            p3: base(fields.p3, 0), p4: base(fields.p4, 0),
            validos: base(fields.validos, 0), blancos: base(fields.blancos, 0),
            nulos: base(fields.nulos, 0),
            fuente_scan: 'mobile_expo',
            razon: rejectReason.trim(),
            observaciones_detectadas: observaciones.map(o => ({ tipo: o.tipo, descripcion: o.texto })),
          }),
        }).catch(() => {})
      }
      setDecision(type === 'anulada' ? 'anulada' : type)
      setPhase('done')
    } catch (e) {
      Alert.alert('Error', e.message || 'No se pudo guardar el acta')
    } finally {
      setSaving(false)
    }
  }

  function updateField(key, val) {
    setFields(prev => ({ ...prev, [key]: val === '' ? null : isNaN(Number(val)) ? val : Number(val) }))
  }

  return (
    <SafeAreaView style={[ss.flex1, { backgroundColor: '#f7f8fa' }]}>
      <StatusBar style="dark" />
      {/* Header */}
      <View style={s.header}>
        <Text style={s.headerTitle}>Scanner RRV</Text>
        <Text style={s.headerUser}>{user?.usuario || user?.nombre || ''}</Text>
        <TouchableOpacity onPress={onLogout}><Text style={s.logoutBtn}>Salir</Text></TouchableOpacity>
      </View>

      <ScrollView style={ss.flex1} contentContainerStyle={{ padding: 16, paddingBottom: 40 }} keyboardShouldPersistTaps="handled">

        {/* Captura de imagen */}
        <View style={s.card}>
          <Text style={s.cardTitle}>Acta Electoral</Text>
          <View style={s.btnRow}>
            <TouchableOpacity style={[s.captureBtn, { flex: 1 }]} onPress={() => pickImage(true)}>
              <Text style={s.captureBtnText}>📷 Cámara</Text>
            </TouchableOpacity>
            <View style={{ width: 10 }} />
            <TouchableOpacity style={[s.captureBtn, { flex: 1, backgroundColor: '#6c5ce7' }]} onPress={() => pickImage(false)}>
              <Text style={s.captureBtnText}>🖼 Galería</Text>
            </TouchableOpacity>
          </View>

          {imageUri && (
            <Image source={{ uri: imageUri }} style={s.preview} resizeMode="contain" />
          )}

          {imageUri && phase === 'idle' && (
            <TouchableOpacity style={[s.btn, { marginTop: 12 }]} onPress={handleProcess}>
              <Text style={s.btnText}>🔍 Transcribir con OCR</Text>
            </TouchableOpacity>
          )}

          {phase === 'processing' && (
            <View style={{ marginTop: 12 }}>
              <View style={s.progressBg}>
                <View style={[s.progressFill, { width: `${progress.pct}%` }]} />
              </View>
              <Text style={s.progressText}>{progress.text}</Text>
            </View>
          )}
        </View>

        {/* Resultados */}
        {phase === 'result' && fields && (
          <View style={s.card}>
            <Text style={s.cardTitle}>Resultado de transcripción</Text>

            {/* Duplicado */}
            {duplicado && (
              <View style={[s.alertBox, { borderColor: '#eab308', backgroundColor: '#fef9c3' }]}>
                <Text style={[s.alertTitle, { color: '#713f12' }]}>⚠️ Acta duplicada detectada</Text>
                <Text style={[s.alertText, { color: '#854d0e' }]}>Ya existe un acta para mesa {duplicado.mesa}. No puede aceptarse.</Text>
              </View>
            )}

            {/* Anulada */}
            {fields.anulada && (
              <View style={[s.alertBox, { borderColor: '#b91c1c', backgroundColor: '#fff0f0' }]}>
                <Text style={[s.alertTitle, { color: '#b91c1c' }]}>🚫 Acta marcada como ANULADA</Text>
                <Text style={[s.alertText, { color: '#7f1d1d' }]}>No puede ingresar al cómputo RRV ni al oficial.</Text>
              </View>
            )}

            {/* Solo bordes */}
            {soloBordes && !fields.anulada && (
              <View style={[s.alertBox, { borderColor: '#0ea5e9', backgroundColor: '#f0f9ff' }]}>
                <Text style={[s.alertText, { color: '#0369a1' }]}>ℹ️ Solo bordes/esquinas dañados — área central legible. El acta puede procesarse.</Text>
              </View>
            )}

            {/* Observaciones OEP */}
            {observaciones.map((obs, i) => (
              <View key={i} style={[s.alertBox, { borderColor: '#f97316', backgroundColor: '#fff7ed' }]}>
                <Text style={[s.alertTitle, { color: '#c2410c' }]}>{obs.tipo}</Text>
                <Text style={[s.alertText, { color: '#7c2d12' }]}>{obs.texto}</Text>
              </View>
            ))}

            {/* Alertas aritméticas */}
            {alerts.map((a, i) => (
              <View key={i} style={[s.alertBox, {
                borderColor: a.t === 'ok' ? '#22c55e' : a.t === 'warn' ? '#eab308' : '#ef4444',
                backgroundColor: a.t === 'ok' ? '#f0fff4' : a.t === 'warn' ? '#fffbeb' : '#fff0f0',
              }]}>
                <Text style={[s.alertText, { color: a.t === 'ok' ? '#16a34a' : a.t === 'warn' ? '#92400e' : '#c0392b' }]}>
                  {a.t === 'ok' ? '✅' : a.t === 'warn' ? '⚠️' : '❌'} {a.m}
                </Text>
              </View>
            ))}

            {/* Campos editables */}
            <Text style={[s.sectionLabel, { marginTop: 12 }]}>Datos extraídos (editables)</Text>
            <View style={s.fieldsGrid}>
              {[
                ['Código Mesa', 'codigo_mesa'], ['Nro. Mesa', 'numero_mesa'],
                [fields.partido_1 || 'Partido 1', 'p1'], [fields.partido_2 || 'Partido 2', 'p2'],
                [fields.partido_3 || 'Partido 3', 'p3'], [fields.partido_4 || 'Partido 4', 'p4'],
                ['Votos Válidos', 'validos'], ['Blancos', 'blancos'],
                ['Nulos', 'nulos'], ['Habilitados', 'habilitados'], ['Papeletas', 'papeletas'],
              ].map(([label, key]) => (
                <View key={key} style={s.fieldItem}>
                  <Text style={s.fieldLabel}>{label}</Text>
                  <TextInput
                    style={s.fieldInput}
                    value={fields[key] != null ? String(fields[key]) : ''}
                    onChangeText={v => updateField(key, v)}
                    keyboardType="numeric"
                  />
                </View>
              ))}
            </View>

            {/* Observaciones del acta */}
            {fields.observaciones ? (
              <View style={[s.alertBox, { borderColor: '#e2e8f0', backgroundColor: '#f8f9fa', marginTop: 8 }]}>
                <Text style={[s.alertText, { color: '#555' }]}><Text style={{ fontWeight: '700' }}>Observaciones del acta: </Text>{fields.observaciones}</Text>
              </View>
            ) : null}

            {/* Razón de rechazo */}
            {showRejectInput && (
              <TextInput
                style={[s.fieldInput, { marginTop: 10, height: 64, textAlignVertical: 'top' }]}
                value={rejectReason} onChangeText={setRejectReason}
                placeholder="Motivo del rechazo..."
                multiline
              />
            )}

            {/* Botones de decisión */}
            <View style={[s.btnRow, { marginTop: 14 }]}>
              {fields.anulada ? (
                <TouchableOpacity style={[s.decisionBtn, { backgroundColor: '#b91c1c' }, saving && ss.disabled]} disabled={saving} onPress={() => handleDecision('anulada')}>
                  <Text style={s.decisionBtnText}>{saving ? '...' : '🚫 Confirmar Anulada'}</Text>
                </TouchableOpacity>
              ) : (
                <>
                  <TouchableOpacity
                    style={[s.decisionBtn, { backgroundColor: '#22c55e' }, (saving || !!duplicado) && ss.disabled]}
                    disabled={saving || !!duplicado} onPress={() => handleDecision('aceptada')}
                  >
                    <Text style={s.decisionBtnText}>{saving ? '...' : '✅ Aceptar'}</Text>
                  </TouchableOpacity>
                  {observaciones.length > 0 && (
                    <TouchableOpacity style={[s.decisionBtn, { backgroundColor: '#f97316' }, saving && ss.disabled]} disabled={saving} onPress={() => handleDecision('observada')}>
                      <Text style={s.decisionBtnText}>📋 Observada</Text>
                    </TouchableOpacity>
                  )}
                  <TouchableOpacity style={[s.decisionBtn, { backgroundColor: '#ef4444' }, saving && ss.disabled]} disabled={saving} onPress={() => handleDecision('rechazada')}>
                    <Text style={s.decisionBtnText}>{saving ? '...' : showRejectInput ? '❌ Confirmar' : '❌ Rechazar'}</Text>
                  </TouchableOpacity>
                </>
              )}
            </View>
          </View>
        )}

        {/* Estado final */}
        {phase === 'done' && decision && (
          <View style={[s.card, {
            borderWidth: 2,
            borderColor: decision === 'aceptada' ? '#22c55e' : decision === 'observada' ? '#f97316' : decision === 'anulada' ? '#b91c1c' : '#ef4444',
            alignItems: 'center',
          }]}>
            <Text style={{ fontSize: 48 }}>
              {decision === 'aceptada' ? '✅' : decision === 'observada' ? '📋' : decision === 'anulada' ? '🚫' : '❌'}
            </Text>
            <Text style={[s.cardTitle, { color: decision === 'aceptada' ? '#16a34a' : decision === 'observada' ? '#c2410c' : '#b91c1c' }]}>
              {decision === 'aceptada' ? 'Acta Aceptada — Registrada en RRV'
                : decision === 'observada' ? 'Acta Observada — No computa'
                : decision === 'anulada' ? 'Acta Anulada'
                : 'Acta Rechazada'}
            </Text>
            {decision === 'aceptada' && fields && (
              <Text style={[s.alertText, { color: '#555', textAlign: 'center', marginTop: 4 }]}>
                {fields.partido_1||'P1'}:{fields.p1} · {fields.partido_2||'P2'}:{fields.p2} · {fields.partido_3||'P3'}:{fields.p3} · {fields.partido_4||'P4'}:{fields.p4}
              </Text>
            )}
            {decision === 'rechazada' && rejectReason ? (
              <Text style={[s.alertText, { color: '#c0392b', marginTop: 4 }]}>Razón: {rejectReason}</Text>
            ) : null}
            <TouchableOpacity style={[s.btn, { marginTop: 16, backgroundColor: '#6c5ce7' }]} onPress={resetAll}>
              <Text style={s.btnText}>Nueva transcripción</Text>
            </TouchableOpacity>
          </View>
        )}
      </ScrollView>
    </SafeAreaView>
  )
}

// ── Estilos ───────────────────────────────────────────────────────────────────
const ss = StyleSheet.create({
  flex1: { flex: 1 },
  center: { justifyContent: 'center', alignItems: 'center' },
  disabled: { opacity: 0.5 },
})

const s = StyleSheet.create({
  // Login
  loginTitle: { fontSize: 26, fontWeight: '800', color: '#fff', marginBottom: 4 },
  loginSub: { fontSize: 14, color: '#94a3b8', marginBottom: 32 },
  loginCard: { backgroundColor: '#fff', borderRadius: 12, padding: 20, width: '100%', maxWidth: 400, shadowColor: '#000', shadowOpacity: 0.1, shadowRadius: 10, elevation: 4 },
  inputLabel: { fontSize: 12, fontWeight: '700', color: '#888', textTransform: 'uppercase', letterSpacing: 0.5, marginBottom: 4, marginTop: 12 },
  input: { borderWidth: 1.5, borderColor: '#ddd', borderRadius: 8, padding: 10, fontSize: 14, color: '#1a1a2e', backgroundColor: '#fafafa' },
  errorText: { color: '#c0392b', fontSize: 13, marginTop: 8 },
  // Shared
  btn: { backgroundColor: '#1a1a2e', borderRadius: 8, padding: 14, alignItems: 'center', marginTop: 12 },
  btnText: { color: '#fff', fontWeight: '700', fontSize: 15 },
  // Header
  header: { flexDirection: 'row', alignItems: 'center', padding: 16, paddingBottom: 10, backgroundColor: '#fff', borderBottomWidth: 1, borderColor: '#f0f0f0' },
  headerTitle: { fontSize: 16, fontWeight: '800', color: '#1a1a2e', flex: 1 },
  headerUser: { fontSize: 12, color: '#888', marginRight: 12 },
  logoutBtn: { fontSize: 13, color: '#6c5ce7', fontWeight: '700' },
  // Card
  card: { backgroundColor: '#fff', borderRadius: 12, padding: 16, marginBottom: 14, shadowColor: '#000', shadowOpacity: 0.05, shadowRadius: 6, elevation: 2 },
  cardTitle: { fontSize: 16, fontWeight: '800', color: '#1a1a2e', marginBottom: 12 },
  sectionLabel: { fontSize: 11, fontWeight: '700', color: '#888', textTransform: 'uppercase', letterSpacing: 0.5, marginBottom: 8 },
  // Capture
  captureBtn: { backgroundColor: '#1a1a2e', borderRadius: 8, padding: 12, alignItems: 'center' },
  captureBtnText: { color: '#fff', fontWeight: '700', fontSize: 14 },
  btnRow: { flexDirection: 'row', gap: 8 },
  preview: { width: '100%', height: 220, borderRadius: 8, marginTop: 12, backgroundColor: '#f0f0f0' },
  // Progress
  progressBg: { height: 8, borderRadius: 4, backgroundColor: '#f0f0f0', overflow: 'hidden', marginBottom: 4 },
  progressFill: { height: '100%', backgroundColor: '#6c5ce7', borderRadius: 4 },
  progressText: { fontSize: 12, color: '#888' },
  // Alerts
  alertBox: { borderWidth: 1.5, borderRadius: 8, padding: 10, marginBottom: 8 },
  alertTitle: { fontSize: 12, fontWeight: '800', marginBottom: 2 },
  alertText: { fontSize: 12, lineHeight: 18 },
  // Fields grid
  fieldsGrid: { flexDirection: 'row', flexWrap: 'wrap', gap: 8 },
  fieldItem: { width: '47%' },
  fieldLabel: { fontSize: 10, fontWeight: '700', color: '#888', textTransform: 'uppercase', letterSpacing: 0.3, marginBottom: 2 },
  fieldInput: { borderWidth: 1.5, borderColor: '#ddd', borderRadius: 6, padding: 8, fontSize: 14, fontWeight: '700', color: '#1a1a2e', backgroundColor: '#fafafa' },
  // Decision buttons
  decisionBtn: { flex: 1, borderRadius: 8, padding: 12, alignItems: 'center' },
  decisionBtnText: { color: '#fff', fontWeight: '700', fontSize: 13 },
})

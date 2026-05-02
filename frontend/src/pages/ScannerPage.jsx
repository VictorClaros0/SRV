import React, { useState } from 'react'
import * as pdfjsLib from 'pdfjs-dist'
import FileUploader from '../components/scanner/FileUploader.jsx'
import CameraCapture from '../components/scanner/CameraCapture.jsx'

pdfjsLib.GlobalWorkerOptions.workerSrc = `https://cdn.jsdelivr.net/npm/pdfjs-dist@${pdfjsLib.version}/build/pdf.worker.min.mjs`

const FIREBASE_RTDB = 'https://lawdocs-75aef-default-rtdb.firebaseio.com'
const CLOUD = { name: 'dk5u8dljb', key: '779231566857641', secret: 'GFzuC8LmqeqAVZQfnOFBa5yrKzA' }

// ── PDF → JPEG comprimido ────────────────────────────────────────────────────
async function pdfToImage(file) {
  const buf = await file.arrayBuffer()
  const pdf = await pdfjsLib.getDocument({ data: buf }).promise
  const page = await pdf.getPage(1)
  const vp = page.getViewport({ scale: 2.0 })
  const canvas = document.createElement('canvas')
  canvas.width = vp.width; canvas.height = vp.height
  await page.render({ canvasContext: canvas.getContext('2d'), viewport: vp }).promise
  return new Promise(r => canvas.toBlob(r, 'image/jpeg', 0.92))
}

async function sha1(m) {
  const b = await crypto.subtle.digest('SHA-1', new TextEncoder().encode(m))
  return Array.from(new Uint8Array(b)).map(x => x.toString(16).padStart(2,'0')).join('')
}

async function uploadCloudinary(file) {
  const ts = Math.floor(Date.now()/1000)
  const sig = await sha1('folder=rrv_actas&timestamp='+ts+CLOUD.secret)
  const fd = new FormData()
  fd.append('file',file); fd.append('api_key',CLOUD.key)
  fd.append('timestamp',ts); fd.append('signature',sig); fd.append('folder','rrv_actas')
  const r = await fetch(`https://api.cloudinary.com/v1_1/${CLOUD.name}/image/upload`,{method:'POST',body:fd})
  if (!r.ok) throw new Error('Error subiendo imagen')
  return (await r.json()).secure_url
}

async function blobToBase64(blob) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result.split(',')[1])
    reader.onerror = reject
    reader.readAsDataURL(blob)
  })
}

// ── Prompt OCR con reglas OEP Bolivia ────────────────────────────────────────
const OCR_PROMPT = `Analiza esta acta electoral boliviana. Extrae los datos y detecta problemas físicos.

REGLAS DE LECTURA:
- Votos en casillas de 3 dígitos: |0|9|5| = 95
- Código de mesa: número largo en esquina superior izquierda
- El ÁREA CENTRAL es el recuadro grande del conteo de votos por candidato

DETECCIÓN DE OBSERVACIONES — responde true/false para cada campo:

- anulada: el documento tiene escrito "ANULADA", "NULA" o sello de anulación visible
- area_central_danada: el recuadro central de conteo está mojado, arrugado, roto o manchado de forma que impide leer los votos; O la hoja está tan mojada/dañada que el centro del cuadrado de votos no se puede leer
- tiene_letras_en_numeros: hay letras (a,b,c,d...) escritas en casillas que deben ser numéricas
- tiene_numeros_dudosos: hay un número sobreescrito, corregido encima con tachón o borrón que no permite distinguir con certeza qué número es (ej: un 7 corregido a 1 pero no se entiende)
- tiene_tachaduras_no_aclaradas: hay tachaduras, borrones o correcciones sobre datos sin aclaración en la sección de Observaciones del acta
- solo_bordes_danados: SOLO los bordes o esquinas están mojados/arrugados pero el área central de votos está completamente legible (este caso NO debe marcar area_central_danada como true)

IMPORTANTE:
- Si solo está mojada una esquina o borde → solo_bordes_danados=true, area_central_danada=false
- Si está mojada/dañada la parte donde están los votos → area_central_danada=true
- Si la hoja dice explícitamente "ANULADA" → anulada=true

Devuelve ÚNICAMENTE este JSON (null para campos no legibles):
{
  "codigo_mesa": número largo de mesa,
  "numero_mesa": número corto de mesa,
  "p1": votos Daenerys Targaryen,
  "p2": votos Sansa Stark,
  "p3": votos Robert Baratheon,
  "p4": votos Tyrion Lannister,
  "partido_1": nombre del partido que corresponde a P1,
  "partido_2": nombre del partido que corresponde a P2,
  "partido_3": nombre del partido que corresponde a P3,
  "partido_4": nombre del partido que corresponde a P4,
  "validos": total votos válidos,
  "blancos": votos blancos,
  "nulos": votos nulos,
  "habilitados": electores habilitados,
  "papeletas": papeletas en ánfora,
  "no_utilizadas": papeletas no utilizadas,
  "apertura_h": hora apertura,
  "apertura_m": minutos apertura,
  "cierre_h": hora cierre,
  "cierre_m": minutos cierre,
  "anulada": true/false,
  "observaciones": texto de observaciones del acta o "",
  "area_central_danada": true/false,
  "tiene_letras_en_numeros": true/false,
  "tiene_numeros_dudosos": true/false,
  "tiene_tachaduras_no_aclaradas": true/false,
  "solo_bordes_danados": true/false
}`

async function callClaudeOCR(imageBlob) {
  const apiKey = import.meta.env.VITE_ANTHROPIC_API_KEY
  if (!apiKey) throw new Error('Falta VITE_ANTHROPIC_API_KEY en frontend/.env')
  const b64 = await blobToBase64(imageBlob)
  const resp = await fetch('https://api.anthropic.com/v1/messages', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'x-api-key': apiKey,
      'anthropic-version': '2023-06-01',
      'anthropic-dangerous-direct-browser-access': 'true',
    },
    body: JSON.stringify({
      model: 'claude-haiku-4-5-20251001',
      max_tokens: 1024,
      messages: [{ role: 'user', content: [
        { type: 'image', source: { type: 'base64', media_type: 'image/jpeg', data: b64 } },
        { type: 'text', text: OCR_PROMPT },
      ]}],
    }),
  })
  const raw = await resp.json()
  if (!resp.ok) throw new Error(raw.error?.message || `Anthropic API error ${resp.status}`)
  const text = raw.content?.[0]?.text || ''
  const match = text.match(/\{[\s\S]*\}/)
  if (!match) throw new Error('Claude no devolvió JSON válido: ' + text.substring(0, 200))
  return JSON.parse(match[0])
}

// ── Clasificar observaciones según normativa OEP / Ley 026 ──────────────────
// Regla: solo_bordes_danados NO genera observación (el acta se puede procesar).
function detectarObservaciones(result) {
  const obs = []
  if (result.area_central_danada)
    obs.push({
      tipo: 'AREA_CENTRAL_DAÑADA',
      texto: 'El área central del recuadro de conteo de votos está dañada, mojada o ilegible. Los datos de votación no son verificables con certeza. Causal de observación según Ley 026 Art. 190: los datos esenciales deben ser legibles para su cómputo.',
    })
  if (result.tiene_letras_en_numeros)
    obs.push({
      tipo: 'LETRAS_EN_CASILLAS_NUMÉRICAS',
      texto: 'Se detectaron letras escritas en casillas que deben contener únicamente números. El acta está mal llenada. La Guía OEP establece que los resultados deben llenarse con números claros, colocando un número en cada casilla.',
    })
  if (result.tiene_numeros_dudosos)
    obs.push({
      tipo: 'NÚMERO_SOBREESCRITO_ILEGIBLE',
      texto: 'Existe al menos un número que fue corregido con tachón o sobreescrito encima y no puede distinguirse con certeza cuál es el valor real. Dato dudoso o inconsistente: no debe entrar al cómputo sin revisión humana.',
    })
  if (result.tiene_tachaduras_no_aclaradas)
    obs.push({
      tipo: 'TACHADURA_SIN_ACLARACIÓN',
      texto: 'Hay borrones, tachaduras o correcciones sobre datos del acta que no están señaladas ni aclaradas en la sección de Observaciones. Causal de nulidad según Ley 026: alteración de datos sin señalamiento en Observaciones invalida el acta.',
    })
  // solo_bordes_danados NO genera observación — el acta se procesa normalmente.
  return obs
}

// ── Validación aritmética ────────────────────────────────────────────────────
function validate(d) {
  const alerts = []
  if (d.anulada)
    alerts.push({ t:'err', m:'Acta marcada como ANULADA — no puede ingresar al cómputo.' })
  const pv = [d.p1,d.p2,d.p3,d.p4].filter(v=>v!=null)
  const sum = pv.reduce((a,b)=>a+b,0)
  if (d.validos!=null && pv.length>0 && sum!==d.validos)
    alerts.push({t:'err', m:`Inconsistencia aritmética: Σ candidatos (${sum}) ≠ válidos (${d.validos})`})
  if (d.validos!=null && d.blancos!=null && d.nulos!=null && d.papeletas!=null) {
    const exp = d.validos + d.blancos + d.nulos
    if (exp!==d.papeletas) alerts.push({t:'err', m:`Papeletas ánfora (${d.papeletas}) ≠ válidos+blancos+nulos (${exp})`})
  }
  if (d.habilitados!=null && d.papeletas!=null && d.papeletas > d.habilitados)
    alerts.push({t:'warn', m:`Papeletas (${d.papeletas}) > habilitados (${d.habilitados})`})
  if (pv.length===0) alerts.push({t:'err', m:'No se pudieron extraer votos por candidato'})
  if (!d.anulada && alerts.filter(a=>a.t==='err').length===0 && pv.length>0)
    alerts.push({t:'ok', m:'Datos aritméticos consistentes ✓'})
  return alerts
}

// ── Guardar acta OBSERVADA: backend log + Firebase ───────────────────────────
async function saveObservadaLog(fields, observaciones, imgUrl, archivo) {
  const now = new Date()
  const id = now.getTime().toString()
  const logPayload = {
    codigo_mesa: fields.codigo_mesa || null,
    numero_mesa: fields.numero_mesa || null,
    p1: fields.p1 || null,
    p2: fields.p2 || null,
    p3: fields.p3 || null,
    p4: fields.p4 || null,
    partido_1: fields.partido_1 || null,
    partido_2: fields.partido_2 || null,
    partido_3: fields.partido_3 || null,
    partido_4: fields.partido_4 || null,
    validos: fields.validos || null,
    blancos: fields.blancos || null,
    nulos: fields.nulos || null,
    habilitados: fields.habilitados || null,
    papeletas: fields.papeletas || null,
    fuente_scan: 'web_scanner',
    anulada: fields.anulada || false,
    razon_principal: observaciones[0]?.texto || 'Acta observada',
    imagen_url: imgUrl || null,
    observaciones_detectadas: observaciones.map(o => ({ tipo: o.tipo, descripcion: o.texto })),
  }
  // Escribir en log del backend (aparece en data/logs/observadas.log del proyecto)
  try {
    await fetch('/api/v1/rrv/scan/observada', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(logPayload),
    })
  } catch (e) { console.warn('Backend log:', e) }
  // También guardar en Firebase para historial
  const log = {
    id, timestamp: now.toISOString(), hora: now.toLocaleTimeString('es-BO'),
    fuente: 'web_scanner', estado: 'observada',
    codigo_acta: fields.codigo_mesa || null,
    imagen_url: imgUrl || null, archivo: archivo || null,
    observaciones_detectadas: observaciones.map(o => ({ tipo: o.tipo, descripcion: o.texto })),
    razon_principal: observaciones[0]?.texto || 'Acta observada',
    datos: {
      p1: fields.p1??null, p2: fields.p2??null, p3: fields.p3??null, p4: fields.p4??null,
      votos_validos: fields.validos??null, votos_blancos: fields.blancos??null,
      votos_nulos: fields.nulos??null, habilitados: fields.habilitados??null,
      papeletas_anfora: fields.papeletas??null,
    },
    observaciones_acta: fields.observaciones || '',
    anulada: fields.anulada || false,
  }
  await fetch(`${FIREBASE_RTDB}/logs_observadas/${id}.json`, {
    method: 'PUT', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(log),
  })
  return id
}

// ── Guardar acta ACEPTADA en RRV (MongoDB conteo rápido) ─────────────────────
async function saveActaRRV(fields, imgUrl) {
  const resp = await fetch('/api/v1/rrv/scan', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      codigo_mesa: Number(fields.codigo_mesa) || 0,
      numero_mesa: Number(fields.numero_mesa) || 0,
      p1: Number(fields.p1) || 0,
      p2: Number(fields.p2) || 0,
      p3: Number(fields.p3) || 0,
      p4: Number(fields.p4) || 0,
      partido_1: fields.partido_1 || null,
      partido_2: fields.partido_2 || null,
      partido_3: fields.partido_3 || null,
      partido_4: fields.partido_4 || null,
      validos: Number(fields.validos) || 0,
      blancos: Number(fields.blancos) || 0,
      nulos: Number(fields.nulos) || 0,
      habilitados: Number(fields.habilitados) || 0,
      papeletas: Number(fields.papeletas) || 0,
      imagen_url: imgUrl || null,
      fuente_scan: 'web_scanner',
    }),
  })
  if (!resp.ok) {
    const err = await resp.json().catch(() => ({}))
    throw new Error(err.error || 'Error guardando en RRV')
  }
  return resp.json()
}

// ── Estilos ──────────────────────────────────────────────────────────────────
const st = {
  card: { background:'#fff', borderRadius:10, padding:'24px 28px', boxShadow:'0 2px 8px rgba(0,0,0,.06)', marginBottom:16 },
  title: { fontSize:22, fontWeight:800, color:'#1a1a2e', marginBottom:4 },
  subtitle: { fontSize:13, color:'#888', marginBottom:20 },
  tabs: { display:'flex', gap:0, marginBottom:20, borderBottom:'2px solid #f0f0f0' },
  tab: (a) => ({ padding:'9px 20px', cursor:'pointer', fontSize:13, fontWeight:600, border:'none', background:'none', color:a?'#1a1a2e':'#888', borderBottom:a?'2px solid #1a1a2e':'2px solid transparent', marginBottom:-2 }),
  btn: (l,d) => ({ padding:'12px 28px', fontSize:15, fontWeight:700, border:'none', borderRadius:8, cursor:d?'not-allowed':'pointer', background:d?'#e0e0e0':l?'#9ca3af':'#1a1a2e', color:d?'#aaa':'#fff', marginTop:20, width:'100%' }),
  progressBar: { height:8, borderRadius:4, background:'#f0f0f0', overflow:'hidden', marginBottom:6, marginTop:12 },
  progressFill: (p) => ({ height:'100%', width:p+'%', background:'linear-gradient(90deg,#6c5ce7,#a29bfe)', borderRadius:4, transition:'width .3s' }),
  progressText: { fontSize:12, color:'#888' },
  grid: { display:'grid', gridTemplateColumns:'repeat(auto-fill, minmax(140px, 1fr))', gap:'10px 12px', marginTop:16 },
  fieldLabel: { fontSize:11, color:'#888', fontWeight:700, textTransform:'uppercase', letterSpacing:.3, marginBottom:4 },
  fieldInput: { padding:'8px 12px', border:'1.5px solid #ddd', borderRadius:6, fontSize:14, fontWeight:700, width:'100%', outline:'none' },
  alertOk: { background:'#f0fff4', border:'1px solid #bbf7d0', borderRadius:6, padding:'8px 12px', fontSize:13, color:'#16a34a', marginBottom:6 },
  alertWarn: { background:'#fffbeb', border:'1px solid #fde68a', borderRadius:6, padding:'8px 12px', fontSize:13, color:'#92400e', marginBottom:6 },
  alertErr: { background:'#fff0f0', border:'1px solid #fac0c0', borderRadius:6, padding:'8px 12px', fontSize:13, color:'#c0392b', marginBottom:6 },
  obsBox: { background:'#fff7ed', border:'2px solid #f97316', borderRadius:8, padding:'12px 16px', marginBottom:8 },
  obsTitle: { fontSize:13, fontWeight:800, color:'#c2410c', marginBottom:4 },
  obsText: { fontSize:12, color:'#7c2d12', lineHeight:1.6 },
  obsBadge: { display:'inline-block', background:'#fed7aa', color:'#9a3412', borderRadius:4, padding:'2px 8px', fontSize:11, fontWeight:700, marginRight:6, marginBottom:4 },
  anuladaBox: { background:'#fff0f0', border:'2px solid #ef4444', borderRadius:8, padding:'14px 18px', marginBottom:12 },
  anuladaTitle: { fontSize:14, fontWeight:800, color:'#b91c1c', marginBottom:4 },
  anuladaText: { fontSize:12, color:'#7f1d1d', lineHeight:1.5 },
  btnAccept: { padding:'12px 24px', fontSize:15, fontWeight:700, border:'none', borderRadius:8, cursor:'pointer', background:'#22c55e', color:'#fff', flex:1 },
  btnReject: { padding:'12px 24px', fontSize:15, fontWeight:700, border:'none', borderRadius:8, cursor:'pointer', background:'#ef4444', color:'#fff', flex:1 },
  btnObservada: { padding:'12px 24px', fontSize:15, fontWeight:700, border:'none', borderRadius:8, cursor:'pointer', background:'#f97316', color:'#fff', flex:1 },
  btnRow: { display:'flex', gap:12, marginTop:16 },
  textarea: { width:'100%', padding:10, border:'1.5px solid #fac0c0', borderRadius:6, fontSize:13, marginTop:8, minHeight:60, fontFamily:'inherit' },
  finalOk: { marginTop:16, padding:20, borderRadius:10, background:'#f0fff4', border:'2px solid #22c55e', textAlign:'center' },
  finalRej: { marginTop:16, padding:20, borderRadius:10, background:'#fff0f0', border:'2px solid #ef4444', textAlign:'center' },
  finalObs: { marginTop:16, padding:20, borderRadius:10, background:'#fff7ed', border:'2px solid #f97316', textAlign:'center' },
  finalAnulada: { marginTop:16, padding:20, borderRadius:10, background:'#fff0f0', border:'2px solid #b91c1c', textAlign:'center' },
  error: { marginTop:12, background:'#fff0f0', border:'1px solid #fac0c0', borderRadius:6, padding:'10px 14px', color:'#c0392b', fontSize:13 },
  badge: { display:'inline-block', background:'#ede9fe', color:'#6c5ce7', borderRadius:6, padding:'2px 10px', fontSize:12, fontWeight:700, marginLeft:8 },
  infoBordes: { background:'#f0f9ff', border:'1px solid #bae6fd', borderRadius:6, padding:'8px 12px', fontSize:12, color:'#0369a1', marginBottom:8 },
}

export default function ScannerPage() {
  const [tab, setTab] = useState('archivo')
  const [file, setFile] = useState(null)
  const [cameraFile, setCameraFile] = useState(null)
  const [loading, setLoading] = useState(false)
  const [progress, setProgress] = useState({ pct:0, text:'' })
  const [error, setError] = useState('')
  const [fields, setFields] = useState(null)
  const [alerts, setAlerts] = useState([])
  const [observaciones, setObservaciones] = useState([])
  const [soloBordes, setSoloBordes] = useState(false)
  const [decision, setDecision] = useState(null)
  const [showRejectInput, setShowRejectInput] = useState(false)
  const [rejectReason, setRejectReason] = useState('')
  const [imgUrl, setImgUrl] = useState(null)
  const [saving, setSaving] = useState(false)
  const [duplicado, setDuplicado] = useState(null)

  const activeFile = tab === 'archivo' ? file : cameraFile
  const canProcess = !!activeFile && !loading

  function reset() {
    setFile(null); setCameraFile(null); setError(''); setFields(null)
    setAlerts([]); setObservaciones([]); setSoloBordes(false); setDecision(null)
    setShowRejectInput(false); setRejectReason(''); setImgUrl(null); setDuplicado(null)
  }

  function updateField(key, val) {
    setFields(prev => ({ ...prev, [key]: val === '' ? null : Number(val) || val }))
  }

  async function handleProcess() {
    if (!activeFile) return
    setLoading(true); setError(''); setFields(null); setAlerts([]); setObservaciones([])
    setSoloBordes(false); setDecision(null); setShowRejectInput(false); setRejectReason(''); setImgUrl(null)
    try {
      let imageBlob = activeFile
      const isPdf = activeFile.name?.toLowerCase().endsWith('.pdf') || activeFile.type === 'application/pdf'
      if (isPdf) {
        setProgress({ pct:10, text:'Convirtiendo PDF a imagen...' })
        imageBlob = await pdfToImage(activeFile)
      }
      setProgress({ pct:30, text:'Cargando...' })
      const result = await callClaudeOCR(imageBlob)
      if (!result) return

      setProgress({ pct:65, text:'Verificando duplicados...' })
      setDuplicado(null)
      if (result.codigo_mesa) {
        try {
          const dup = await fetch(`/api/v1/rrv/scan/check?mesa=${result.codigo_mesa}`)
          const dupData = await dup.json()
          if (dupData.existe || dupData.duplicada) {
            setDuplicado(dupData)
          }
        } catch (e) { console.warn('Check duplicado:', e) }
      }

      setProgress({ pct:75, text:'Detectando observaciones OEP...' })
      const obs = detectarObservaciones(result)
      setObservaciones(obs)
      setSoloBordes(!!result.solo_bordes_danados && !result.area_central_danada)

      setProgress({ pct:85, text:'Validando datos...' })
      setFields(result)
      setAlerts(validate(result))

      setProgress({ pct:93, text:'Subiendo imagen...' })
      try {
        const url = await uploadCloudinary(activeFile)
        setImgUrl(url)
      } catch (e) { console.warn('Cloudinary:', e) }

      setProgress({ pct:100, text:'Listo ✓' })
    } catch (e) {
      setError(e.message || 'Error procesando el acta')
    } finally {
      setLoading(false)
    }
  }

  async function handleDecision(type) {
    if (type === 'rechazada' && !showRejectInput) { setShowRejectInput(true); return }
    if (type === 'rechazada' && !rejectReason.trim()) { setError('Ingresa la razón del rechazo'); return }
    setSaving(true); setError('')
    try {
      if (type === 'observada') {
        await saveObservadaLog(fields, observaciones, imgUrl, activeFile?.name)
        setDecision('observada')
      } else if (type === 'anulada') {
        // Acta marcada como ANULADA: solo log, no entra al cómputo
        await saveObservadaLog(
          fields,
          [{ tipo: 'ACTA_ANULADA', texto: 'El acta tiene marcación explícita de ANULADA. No puede ingresar al cómputo.' }],
          imgUrl, activeFile?.name,
        )
        setDecision('anulada')
      } else if (type === 'aceptada') {
        if (duplicado) {
          throw new Error('Acta duplicada detectada. No puede aceptarse. Marque Observada o Rechazada.')
        }
        // Guardar en RRV (MongoDB conteo rápido)
        await saveActaRRV(fields, imgUrl)
        // Firebase: historial y datos completos para el dashboard
        const scanId = Date.now().toString()
        await fetch(`${FIREBASE_RTDB}/mobile_scans/${scanId}.json`, {
          method: 'PUT', headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            id: scanId, timestamp: new Date().toISOString(),
            hora: new Date().toLocaleTimeString('es-BO'),
            codigo_acta: fields?.codigo_mesa || null,
            imagen_url: imgUrl, estado: 'aceptada', fuente: 'web_scanner',
            p1: Number(fields?.p1)||0, p2: Number(fields?.p2)||0,
            p3: Number(fields?.p3)||0, p4: Number(fields?.p4)||0,
            partido_1: fields?.partido_1||null, partido_2: fields?.partido_2||null,
            partido_3: fields?.partido_3||null, partido_4: fields?.partido_4||null,
            votos_validos: Number(fields?.validos)||0,
            votos_blancos: Number(fields?.blancos)||0,
            votos_nulos: Number(fields?.nulos)||0,
            papeletas: Number(fields?.papeletas)||0,
            datos: fields,
          }),
        })
        // Índice por mesa para detección de duplicados desde app móvil
        if (fields?.codigo_mesa) {
          await fetch(`${FIREBASE_RTDB}/mesa_index/${fields.codigo_mesa}.json`, {
            method: 'PUT', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ scan_id: scanId, timestamp: new Date().toISOString(), fuente: 'web_scanner', estado: 'aceptada' }),
          }).catch(() => {})
        }
        setDecision('aceptada')
      } else {
        // rechazada — log en backend + Firebase
        try {
          await fetch('/api/v1/rrv/scan/rechazada', {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              codigo_mesa: Number(fields?.codigo_mesa) || 0,
              numero_mesa: Number(fields?.numero_mesa) || 0,
              p1: Number(fields?.p1) || 0, p2: Number(fields?.p2) || 0,
              p3: Number(fields?.p3) || 0, p4: Number(fields?.p4) || 0,
              validos: Number(fields?.validos) || 0,
              blancos: Number(fields?.blancos) || 0,
              nulos: Number(fields?.nulos) || 0,
              habilitados: Number(fields?.habilitados) || 0,
              papeletas: Number(fields?.papeletas) || 0,
              imagen_url: imgUrl || null,
              fuente_scan: 'web_scanner',
              razon: rejectReason.trim(),
              observaciones_detectadas: observaciones.map(o => ({ tipo: o.tipo, descripcion: o.texto })),
            }),
          })
        } catch (e) { console.warn('Backend rechazada log:', e) }
        const scanId = Date.now().toString()
        await fetch(`${FIREBASE_RTDB}/mobile_scans/${scanId}.json`, {
          method: 'PUT', headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            id: scanId, timestamp: new Date().toISOString(),
            hora: new Date().toLocaleTimeString('es-BO'),
            codigo_acta: fields?.codigo_mesa || null,
            imagen_url: imgUrl, estado: 'rechazada',
            razon_rechazo: rejectReason.trim(), fuente: 'web_scanner',
          }),
        })
        setDecision('rechazada')
      }
    } catch (e) {
      setError('Error guardando: ' + e.message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <div style={{ marginBottom:24 }}>
        <div style={st.title}>
          Scanner / Transcriptor de Actas
          <span style={st.badge}>Claude Vision IA</span>
        </div>
        <div style={st.subtitle}>Adjunta o captura el acta. Claude Vision extrae los datos y detecta observaciones según normativa OEP / Ley 026.</div>
      </div>

      <div style={st.card}>
        <div style={st.tabs}>
          <button style={st.tab(tab==='archivo')} onClick={()=>{setTab('archivo');setError('')}}>📎 Adjuntar archivo</button>
          <button style={st.tab(tab==='camara')} onClick={()=>{setTab('camara');setError('')}}>📷 Usar cámara</button>
        </div>
        {tab==='archivo' && <FileUploader file={file} onFile={setFile} onError={msg=>{if(msg)setError(msg)}} />}
        {tab==='camara' && <CameraCapture onCapture={setCameraFile} />}
        <button style={st.btn(loading,!canProcess)} disabled={!canProcess} onClick={handleProcess}>
          {loading ? 'Cargando...' : '🔍 Transcribir con OCR'}
        </button>
        {loading && (
          <div>
            <div style={st.progressBar}><div style={st.progressFill(progress.pct)} /></div>
            <div style={st.progressText}>{progress.text}</div>
          </div>
        )}
        {error && !loading && <div style={st.error}>{error}</div>}
      </div>

      {fields && !decision && (
        <div style={st.card}>
          <div style={{ fontSize:15, fontWeight:700, color:'#1a1a2e', marginBottom:12 }}>
            📊 Resultado de transcripción
            {observaciones.length > 0 && (
              <span style={{ marginLeft:10, background:'#fed7aa', color:'#9a3412', borderRadius:6, padding:'2px 10px', fontSize:12, fontWeight:700 }}>
                ⚠️ {observaciones.length} observación{observaciones.length>1?'es':''} detectada{observaciones.length>1?'s':''}
              </span>
            )}
            {fields.anulada && (
              <span style={{ marginLeft:10, background:'#fecaca', color:'#b91c1c', borderRadius:6, padding:'2px 10px', fontSize:12, fontWeight:700 }}>
                🚫 ANULADA
              </span>
            )}
          </div>

          {/* Acta duplicada — advertencia */}
          {duplicado && (
            <div style={{ background:'#fef9c3', border:'2px solid #eab308', borderRadius:8, padding:'12px 16px', marginBottom:10 }}>
              <div style={{ fontSize:13, fontWeight:800, color:'#713f12', marginBottom:4 }}>⚠️ Acta duplicada detectada</div>
              <div style={{ fontSize:12, color:'#854d0e' }}>
                Ya existe un acta para la mesa <strong>{duplicado.mesa}</strong> registrada anteriormente desde <strong>{duplicado.fuente || 'scanner'}</strong>.
                Esta acta no puede aceptarse nuevamente en RRV; por favor marque Observada o Rechazada.
              </div>
            </div>
          )}

          {/* Acta anulada — bloqueo total */}
          {fields.anulada && (
            <div style={st.anuladaBox}>
              <div style={st.anuladaTitle}>🚫 Acta marcada como ANULADA</div>
              <div style={st.anuladaText}>
                El acta tiene marcación explícita de nulidad. No puede ingresar al cómputo RRV ni al oficial.
                Se registrará en el log de observadas para trazabilidad.
              </div>
            </div>
          )}

          {/* Info solo bordes dañados — no bloquea */}
          {soloBordes && !fields.anulada && (
            <div style={st.infoBordes}>
              ℹ️ <strong>Solo bordes/esquinas dañados:</strong> El área central de votos es legible.
              El acta puede procesarse normalmente según normativa OEP.
            </div>
          )}

          {/* Observaciones OEP — solo advertencias, no bloquean */}
          {observaciones.map((obs, i) => (
            <div key={i} style={st.obsBox}>
              <div style={st.obsTitle}>
                <span style={st.obsBadge}>{obs.tipo}</span>
                Advertencia — Se recomienda revisión manual
              </div>
              <div style={st.obsText}>{obs.texto}</div>
            </div>
          ))}

          {/* Alertas aritméticas */}
          {alerts.filter(a => !fields.anulada || a.t === 'err').map((a, i) => (
            <div key={i} style={a.t==='ok'?st.alertOk:a.t==='warn'?st.alertWarn:st.alertErr}>
              {a.t==='ok'?'✅':a.t==='warn'?'⚠️':'❌'} {a.m}
            </div>
          ))}

          <div style={st.grid}>
            {[
              ['Código Mesa','codigo_mesa'],['Nro. Mesa','numero_mesa'],
              [fields.partido_1 || 'Partido 1','p1'],[fields.partido_2 || 'Partido 2','p2'],[fields.partido_3 || 'Partido 3','p3'],[fields.partido_4 || 'Partido 4','p4'],
              ['Votos Válidos','validos'],['Votos Blancos','blancos'],['Votos Nulos','nulos'],
              ['Habilitados','habilitados'],['Papeletas Ánfora','papeletas'],['No Utilizadas','no_utilizadas'],
            ].map(([label, key]) => (
              <div key={key}>
                <div style={st.fieldLabel}>{label}</div>
                <input style={st.fieldInput} value={fields[key]??''} onChange={e=>updateField(key,e.target.value)} />
              </div>
            ))}
          </div>

          {fields.observaciones && (
            <div style={{ marginTop:12, fontSize:13, color:'#555', background:'#f8f9fa', borderRadius:6, padding:'8px 12px' }}>
              <strong>Observaciones del acta:</strong> {fields.observaciones}
            </div>
          )}

          {showRejectInput && (
            <div>
              <div style={{ ...st.fieldLabel, marginTop:14 }}>Razón de rechazo</div>
              <textarea style={st.textarea} value={rejectReason} onChange={e=>setRejectReason(e.target.value)}
                placeholder="Motivo por el que no se toma en cuenta esta acta..." />
            </div>
          )}

          <div style={st.btnRow}>
            {/* Acta anulada: solo botón de confirmar anulación */}
            {fields.anulada ? (
              <button style={{...st.btnReject, opacity:saving?0.5:1}} disabled={saving} onClick={()=>handleDecision('anulada')}>
                🚫 Confirmar Anulación (no entra al cómputo)
              </button>
            ) : (
              <>
                <button style={{...st.btnAccept, opacity:saving?0.5:1}} disabled={saving || !!duplicado} onClick={()=>handleDecision('aceptada')}>
                  ✅ Aceptar — Enviar a RRV
                </button>
                {observaciones.length > 0 && (
                  <button style={{...st.btnObservada, opacity:saving?0.5:1}} disabled={saving} onClick={()=>handleDecision('observada')}>
                    📋 Marcar Observada (no computa)
                  </button>
                )}
                <button style={{...st.btnReject, opacity:saving?0.5:1}} disabled={saving} onClick={()=>handleDecision('rechazada')}>
                  ❌ {showRejectInput ? 'Confirmar Rechazo' : 'Rechazar Acta'}
                </button>
              </>
            )}
          </div>
        </div>
      )}

      {decision === 'aceptada' && (
        <div style={st.finalOk}>
          <div style={{ fontSize:48 }}>✅</div>
          <div style={{ fontSize:20, fontWeight:800, color:'#16a34a', marginBottom:4 }}>Acta Aceptada — Registrada en RRV (Conteo Rápido)</div>
          <div style={{ fontSize:13, color:'#555' }}>
            {fields?.partido_1||'Partido 1'}:{fields?.p1} · {fields?.partido_2||'Partido 2'}:{fields?.p2} · {fields?.partido_3||'Partido 3'}:{fields?.p3} · {fields?.partido_4||'Partido 4'}:{fields?.p4} ·
            Válidos:{fields?.validos} · Nulos:{fields?.nulos} · Blancos:{fields?.blancos}
          </div>
          {imgUrl && <div style={{ marginTop:8 }}><a href={imgUrl} target="_blank" rel="noreferrer" style={{ color:'#6c5ce7', fontSize:13 }}>📎 Ver imagen</a></div>}
          <button onClick={reset} style={{ marginTop:12, background:'#f0f0f0', border:'none', borderRadius:6, padding:'8px 18px', cursor:'pointer', fontSize:13 }}>
            Nueva transcripción
          </button>
        </div>
      )}
      {decision === 'observada' && (
        <div style={st.finalObs}>
          <div style={{ fontSize:48 }}>📋</div>
          <div style={{ fontSize:20, fontWeight:800, color:'#c2410c', marginBottom:4 }}>Acta Registrada como Observada</div>
          <div style={{ fontSize:13, color:'#7c2d12', marginBottom:8 }}>
            No entra al cómputo. Guardada en <code>data/logs/observadas.log</code> del proyecto y en Firebase para revisión humana.
          </div>
          {observaciones.map((o,i) => (
            <div key={i} style={{ fontSize:12, color:'#9a3412', marginBottom:4 }}>• {o.texto}</div>
          ))}
          <button onClick={reset} style={{ marginTop:12, background:'#f0f0f0', border:'none', borderRadius:6, padding:'8px 18px', cursor:'pointer', fontSize:13 }}>
            Nueva transcripción
          </button>
        </div>
      )}
      {decision === 'anulada' && (
        <div style={st.finalAnulada}>
          <div style={{ fontSize:48 }}>🚫</div>
          <div style={{ fontSize:20, fontWeight:800, color:'#b91c1c', marginBottom:4 }}>Acta Anulada — No entra al cómputo</div>
          <div style={{ fontSize:13, color:'#7f1d1d', marginBottom:8 }}>
            Registrada en log de observadas. El TED o autoridad electoral debe revisar el material físico.
          </div>
          <button onClick={reset} style={{ marginTop:12, background:'#f0f0f0', border:'none', borderRadius:6, padding:'8px 18px', cursor:'pointer', fontSize:13 }}>
            Nueva transcripción
          </button>
        </div>
      )}
      {decision === 'rechazada' && (
        <div style={st.finalRej}>
          <div style={{ fontSize:48 }}>❌</div>
          <div style={{ fontSize:20, fontWeight:800, color:'#c0392b', marginBottom:4 }}>Acta Rechazada</div>
          <div style={{ fontSize:13, color:'#c0392b', background:'rgba(239,68,68,.08)', padding:10, borderRadius:6, marginTop:8, textAlign:'left' }}>
            <strong>Razón:</strong> {rejectReason}
          </div>
          <button onClick={reset} style={{ marginTop:12, background:'#f0f0f0', border:'none', borderRadius:6, padding:'8px 18px', cursor:'pointer', fontSize:13 }}>
            Nueva transcripción
          </button>
        </div>
      )}
    </div>
  )
}

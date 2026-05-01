import React, { useState } from 'react'
import { api } from '../api.js'
import FileUploader from '../components/scanner/FileUploader.jsx'
import CameraCapture from '../components/scanner/CameraCapture.jsx'
import TranscriptionResult from '../components/scanner/TranscriptionResult.jsx'

const s = {
  header: { marginBottom: 24 },
  title: { fontSize: 22, fontWeight: 800, color: '#1a1a2e', marginBottom: 4 },
  subtitle: { fontSize: 13, color: '#888' },
  card: {
    background: '#fff', borderRadius: 10, padding: '24px 28px',
    boxShadow: '0 2px 8px rgba(0,0,0,.06)', marginBottom: 16,
  },
  sectionTitle: {
    fontSize: 13, fontWeight: 700, color: '#1a1a2e',
    marginBottom: 14, textTransform: 'uppercase', letterSpacing: .4,
  },
  tabs: { display: 'flex', gap: 0, marginBottom: 20, borderBottom: '2px solid #f0f0f0' },
  tab: (active) => ({
    padding: '9px 20px', cursor: 'pointer', fontSize: 13, fontWeight: 600,
    border: 'none', background: 'none', color: active ? '#1a1a2e' : '#888',
    borderBottom: active ? '2px solid #1a1a2e' : '2px solid transparent',
    marginBottom: -2, transition: 'color .15s',
  }),
  optRow: { display: 'flex', alignItems: 'flex-end', gap: 12, marginTop: 14, flexWrap: 'wrap' },
  optLabel: { fontSize: 11, fontWeight: 700, color: '#888', textTransform: 'uppercase', letterSpacing: .4, marginBottom: 4 },
  optInput: {
    padding: '7px 12px', border: '1.5px solid #ddd', borderRadius: 6, fontSize: 13,
    outline: 'none', width: 160,
  },
  btnTranscribir: (loading, disabled) => ({
    padding: '12px 28px', fontSize: 15, fontWeight: 700, border: 'none',
    borderRadius: 8, cursor: disabled ? 'not-allowed' : 'pointer',
    background: disabled ? '#e0e0e0' : loading ? '#9ca3af' : '#1a1a2e',
    color: disabled ? '#aaa' : '#fff',
    marginTop: 20, width: '100%',
    transition: 'background .2s',
  }),
  statusRow: {
    display: 'flex', alignItems: 'center', gap: 8, marginTop: 12,
    fontSize: 13, color: '#555',
  },
  spinner: {
    width: 16, height: 16, border: '2px solid #e0e0e0',
    borderTopColor: '#1a1a2e', borderRadius: '50%',
    animation: 'spin .7s linear infinite',
    flexShrink: 0,
  },
  error: {
    marginTop: 12, background: '#fff0f0', border: '1px solid #fac0c0',
    borderRadius: 6, padding: '10px 14px', color: '#c0392b', fontSize: 13,
  },
  resetBtn: {
    marginTop: 14, background: '#f0f0f0', border: 'none', borderRadius: 6,
    padding: '8px 18px', cursor: 'pointer', fontSize: 13, color: '#444',
  },
  infoBox: {
    background: '#f0f7ff', border: '1px solid #bfdbfe', borderRadius: 6,
    padding: '10px 14px', fontSize: 12, color: '#1e40af', marginBottom: 12,
  },
}

export default function ScannerPage() {
  const [tab, setTab] = useState('archivo') // 'archivo' | 'camara'
  const [file, setFile] = useState(null)
  const [cameraFile, setCameraFile] = useState(null)
  const [codigoActa, setCodigoActa] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [result, setResult] = useState(null)

  const activeFile = tab === 'archivo' ? file : cameraFile
  const canTranscribir = !!activeFile && !loading

  function reset() {
    setFile(null)
    setCameraFile(null)
    setError('')
    setResult(null)
    setCodigoActa('')
  }

  async function handleTranscribir() {
    if (!activeFile) {
      setError('Debes adjuntar un archivo o capturar una foto antes de transcribir.')
      return
    }
    setLoading(true)
    setError('')
    setResult(null)

    try {
      const opts = {}
      if (codigoActa.trim()) opts.codigoActa = codigoActa.trim()

      const res = await api.scanner.transcribir(activeFile, opts)
      setResult(res)
    } catch (e) {
      const msg = e.message || 'Error desconocido.'
      if (msg.toLowerCase().includes('401') || msg.toLowerCase().includes('sesión')) {
        setError('Tu sesión ha expirado. Por favor inicia sesión nuevamente.')
      } else if (msg.toLowerCase().includes('fetch') || msg.toLowerCase().includes('network')) {
        setError('No se pudo conectar con el backend. Verifica que el servidor esté corriendo en localhost:8080.')
      } else {
        setError(msg)
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>

      <div style={s.header}>
        <div style={s.title}>Scanner / Transcriptor de Actas</div>
        <div style={s.subtitle}>
          Captura o adjunta una imagen del acta y envíala al sistema para transcripción automática.
        </div>
      </div>

      <div style={s.card}>
        <div style={s.infoBox}>
          El backend simula la transcripción usando datos del archivo Transcripciones.csv.
          Si provees un código de acta, el sistema buscará esa acta específica y actualizará sus datos.
          Si no provees código, se usará una acta aleatoria como demostración.
        </div>

        {/* Tabs */}
        <div style={s.tabs}>
          <button style={s.tab(tab === 'archivo')} onClick={() => { setTab('archivo'); setError('') }}>
            📎 Adjuntar archivo
          </button>
          <button style={s.tab(tab === 'camara')} onClick={() => { setTab('camara'); setError('') }}>
            📷 Usar cámara
          </button>
        </div>

        {/* Tab: Archivo */}
        {tab === 'archivo' && (
          <FileUploader
            file={file}
            onFile={setFile}
            onError={msg => { if (msg) setError(msg) }}
          />
        )}

        {/* Tab: Cámara */}
        {tab === 'camara' && (
          <CameraCapture onCapture={setCameraFile} />
        )}

        {/* Código de acta opcional */}
        <div style={s.optRow}>
          <div>
            <div style={s.optLabel}>Código de acta (opcional)</div>
            <input
              style={s.optInput}
              type="number"
              placeholder="Ej. 12345"
              value={codigoActa}
              onChange={e => setCodigoActa(e.target.value)}
            />
          </div>
        </div>

        {/* Botón principal */}
        <button
          style={s.btnTranscribir(loading, !canTranscribir)}
          disabled={!canTranscribir}
          onClick={handleTranscribir}
        >
          {loading ? 'Transcribiendo...' : 'Transcribir acta'}
        </button>

        {/* Estado de carga */}
        {loading && (
          <div style={s.statusRow}>
            <div style={s.spinner} />
            Enviando archivo al backend y procesando transcripción...
          </div>
        )}

        {/* Error */}
        {error && !loading && (
          <div style={s.error}>{error}</div>
        )}
      </div>

      {/* Resultado */}
      {result && (
        <div style={s.card}>
          <TranscriptionResult result={result} />
          <button style={s.resetBtn} onClick={reset}>
            Nueva transcripción
          </button>
        </div>
      )}
    </div>
  )
}

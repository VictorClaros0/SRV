import React, { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api.js'

const ESTADO_COLOR = {
  impresa:    '#d4800a',
  transcrita: '#27ae60',
  observada:  '#c0392b',
}
const ESTADO_LABEL = {
  impresa:    'Impresa',
  transcrita: 'Transcrita',
  observada:  'Observada',
}

const s = {
  header: { display: 'flex', alignItems: 'center', marginBottom: 16, gap: 12, flexWrap: 'wrap' },
  title: { fontSize: 22, fontWeight: 700, color: '#1a1a2e', flex: 1 },
  btnPrimary: {
    background: '#1a1a2e', color: '#fff', border: 'none', borderRadius: 8,
    padding: '9px 18px', cursor: 'pointer', fontSize: 14, fontWeight: 600,
  },
  btnN8n: {
    background: '#f0fff4', color: '#27ae60', border: '1.5px solid #27ae60',
    borderRadius: 8, padding: '9px 18px', cursor: 'pointer', fontSize: 14, fontWeight: 600,
  },
  btnFallback: {
    background: '#fff7ed', color: '#c05621', border: '1.5px solid #c05621',
    borderRadius: 8, padding: '9px 18px', cursor: 'pointer', fontSize: 14, fontWeight: 600,
  },
  btnDisabled: {
    background: '#f0f0f0', color: '#aaa', border: 'none', borderRadius: 8,
    padding: '9px 18px', cursor: 'not-allowed', fontSize: 14, fontWeight: 600,
  },
  search: {
    padding: '8px 14px', border: '1.5px solid #ddd', borderRadius: 8,
    fontSize: 14, outline: 'none', width: 220,
  },
  contadores: {
    display: 'flex', gap: 16, marginBottom: 12, flexWrap: 'wrap', alignItems: 'center',
  },
  chip: (color) => ({
    padding: '4px 14px', borderRadius: 20, fontSize: 13, fontWeight: 600,
    background: color + '20', color, border: '1px solid ' + color + '40',
  }),
  progressWrap: {
    marginBottom: 14, background: '#f0f0f0', borderRadius: 8, height: 10, overflow: 'hidden',
  },
  progressBar: (pct) => ({
    height: '100%', borderRadius: 8, width: pct + '%',
    background: 'linear-gradient(90deg, #27ae60, #1a1a2e)',
    transition: 'width .4s ease',
  }),
  progressLabel: { fontSize: 12, color: '#666', marginBottom: 4 },
  table: {
    width: '100%', borderCollapse: 'collapse', background: '#fff',
    borderRadius: 10, overflow: 'hidden', boxShadow: '0 2px 8px rgba(0,0,0,.06)',
  },
  th: {
    background: '#f7f8fa', padding: '12px 16px', textAlign: 'left',
    fontSize: 12, fontWeight: 700, color: '#666', textTransform: 'uppercase', letterSpacing: .5,
  },
  td: { padding: '12px 16px', fontSize: 14, borderTop: '1px solid #f0f0f0' },
  badge: (estado) => ({
    display: 'inline-block', padding: '2px 10px', borderRadius: 20,
    fontSize: 12, fontWeight: 600,
    background: (ESTADO_COLOR[estado] || '#888') + '20',
    color: ESTADO_COLOR[estado] || '#888',
  }),
  pagination: { display: 'flex', gap: 8, marginTop: 16, alignItems: 'center', justifyContent: 'flex-end' },
  pageBtn: (active) => ({
    padding: '5px 12px', borderRadius: 6, border: '1px solid #ddd',
    background: active ? '#1a1a2e' : '#fff', color: active ? '#fff' : '#444',
    cursor: 'pointer', fontSize: 13,
  }),
  empty: { textAlign: 'center', padding: 40, color: '#888' },
  info: { fontSize: 13, color: '#888' },
  alert: (ok) => ({
    marginBottom: 12, padding: '8px 14px', borderRadius: 6,
    background: ok ? '#f0fff4' : '#fff0f0',
    color: ok ? '#27ae60' : '#c0392b', fontSize: 13,
  }),
}

const PAGE_SIZE = 20

export default function Actas() {
  const [actas, setActas] = useState([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [msg, setMsg] = useState({ texto: '', ok: true })
  const [simulando, setSimulando] = useState(false)
  const [n8nFallbackDisponible, setN8nFallbackDisponible] = useState(false)
  const pollRef = useRef(null)
  const nav = useNavigate()

  useEffect(() => {
    loadActas()
    return () => clearInterval(pollRef.current)
  }, [])

  async function loadActas() {
    try {
      const data = await api.actas.list()
      setActas(data || [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  // Contadores de estado
  const conteo = actas.reduce((acc, a) => {
    acc[a.estado] = (acc[a.estado] || 0) + 1
    return acc
  }, {})
  const total = actas.length
  const impresa = conteo.impresa || 0
  const transcrita = conteo.transcrita || 0
  const observada = conteo.observada || 0
  const progresoPct = total > 0 ? Math.round(((transcrita + observada) / total) * 100) : 0

  // Lanza transcripción a través del backend (que llama a n8n).
  // opts.fallback=true: backend procesa el CSV directamente si n8n no responde.
  async function handleSimular(opts = {}) {
    setMsg({ texto: '', ok: true })
    setN8nFallbackDisponible(false)
    setSimulando(true)
    try {
      const snapshot = {
        transcrita: actas.filter(a => a.estado === 'transcrita').length,
        observada: actas.filter(a => a.estado === 'observada').length,
      }

      // Llama al backend, que orquesta n8n (nunca a n8n directamente)
      const res = await api.transcripcion.simular(opts)

      // El backend devuelve HTTP 200 siempre pero success:false cuando n8n falla
      if (!res?.success) {
        const detalle = [res?.message, res?.hint].filter(Boolean).join('. ')
        setMsg({ texto: detalle || 'n8n no disponible.', ok: false })
        setN8nFallbackDisponible(true) // mostrar botón de fallback
        setSimulando(false)
        return
      }

      // Fallback síncrono completado en el propio backend
      if (res.estado === 'COMPLETADO') {
        setMsg({
          texto: `Completado (${res.source === 'backend_fallback' ? 'fallback directo' : 'n8n'}): ${res.actualizadas || 0} transcritas, ${res.observadas || 0} observadas.`,
          ok: true,
        })
        loadActas()
        setSimulando(false)
        return
      }

      // n8n aceptó el webhook — polling hasta que las actas cambien
      setMsg({ texto: 'Flujo iniciado en n8n. Las actas se actualizarán automáticamente...', ok: true })
      let ticks = 0
      const MAX_TICKS = 60
      pollRef.current = setInterval(async () => {
        ticks++
        const data = await api.actas.list().catch(() => null)
        if (!data) return
        setActas(data)
        const impresa = data.filter(a => a.estado === 'impresa').length
        const transcrita = data.filter(a => a.estado === 'transcrita').length
        const observada = data.filter(a => a.estado === 'observada').length
        const estabilizado = impresa === 0 && (transcrita !== snapshot.transcrita || observada !== snapshot.observada)
        if (estabilizado || ticks >= MAX_TICKS) {
          clearInterval(pollRef.current)
          setSimulando(false)
          setMsg({ texto: `Transcripción completada. ${transcrita} transcritas, ${observada} observadas.`, ok: true })
        }
      }, 2000)
    } catch (e) {
      setMsg({ texto: 'Error: ' + e.message, ok: false })
      setN8nFallbackDisponible(true)
      setSimulando(false)
    }
  }

  const filtered = actas.filter(a =>
    !search ||
    String(a.codigoActa).includes(search) ||
    String(a.codigoRecinto).includes(search)
  )
  const pages = Math.ceil(filtered.length / PAGE_SIZE)
  const slice = filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)

  return (
    <div>
      <div style={s.header}>
        <div style={s.title}>Actas</div>
        <input
          style={s.search}
          placeholder="Buscar por código..."
          value={search}
          onChange={e => { setSearch(e.target.value); setPage(1) }}
        />
        <button style={s.btnPrimary} onClick={() => nav('/actas/nueva')}>+ Nueva acta</button>
        <button
          style={simulando ? s.btnDisabled : s.btnN8n}
          onClick={() => handleSimular()}
          disabled={simulando}
          title="Dispara el workflow de n8n que transcribe todas las actas impresas"
        >
          {simulando ? 'Simulando...' : 'Simular transcripción con n8n'}
        </button>
        {n8nFallbackDisponible && (
          <button
            style={simulando ? s.btnDisabled : s.btnFallback}
            onClick={() => handleSimular({ fallback: true })}
            disabled={simulando}
            title="Procesa Transcripciones.csv desde el backend sin llamar a n8n"
          >
            Simular sin n8n (fallback)
          </button>
        )}
      </div>

      {/* Contadores de estado */}
      <div style={s.contadores}>
        <span style={s.chip(ESTADO_COLOR.transcrita)}>Transcritas: {transcrita}</span>
        <span style={s.chip(ESTADO_COLOR.observada)}>Observadas: {observada}</span>
        <span style={{ fontSize: 13, color: '#888' }}>Total: {total}</span>
      </div>

      {/* Barra de progreso */}
      {(simulando || progresoPct > 0) && (
        <>
          <div style={s.progressLabel}>{progresoPct}% procesado</div>
          <div style={s.progressWrap}>
            <div style={s.progressBar(progresoPct)} />
          </div>
        </>
      )}

      {/* Mensaje de estado */}
      {msg.texto && (
        <div style={s.alert(msg.ok)}>{msg.texto}</div>
      )}

      {loading ? (
        <div style={s.empty}>Cargando...</div>
      ) : (
        <>
          <table style={s.table}>
            <thead>
              <tr>
                <th style={s.th}>Código Acta</th>
                <th style={s.th}>Cód. Recinto</th>
                <th style={s.th}>Mesa</th>
                <th style={s.th}>Estado</th>
                <th style={s.th}>P1</th>
                <th style={s.th}>P2</th>
                <th style={s.th}>P3</th>
                <th style={s.th}>P4</th>
                <th style={s.th}>Válidos</th>
                <th style={s.th}>Nulos</th>
                <th style={s.th}>Blancos</th>
                <th style={s.th}>Observaciones</th>
              </tr>
            </thead>
            <tbody>
              {slice.length === 0 && (
                <tr><td colSpan={12} style={s.empty}>Sin resultados</td></tr>
              )}
              {slice.map(a => (
                <tr key={a.id}>
                  <td style={s.td}>{a.codigoActa}</td>
                  <td style={s.td}>{a.codigoRecinto}</td>
                  <td style={s.td}>{a.nroMesa}</td>
                  <td style={s.td}>
                    <span style={s.badge(a.estado)}>{ESTADO_LABEL[a.estado] || a.estado}</span>
                  </td>
                  <td style={s.td}>{a.p1}</td>
                  <td style={s.td}>{a.p2}</td>
                  <td style={s.td}>{a.p3}</td>
                  <td style={s.td}>{a.p4}</td>
                  <td style={s.td}>{a.votosValidos}</td>
                  <td style={s.td}>{a.votosNulos}</td>
                  <td style={s.td}>{a.votosBlanco}</td>
                  <td style={s.td}>{a.observaciones || '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>

          <div style={s.pagination}>
            <span style={s.info}>{filtered.length} actas — página {page} de {pages || 1}</span>
            <button style={s.pageBtn(false)} disabled={page === 1} onClick={() => setPage(p => p - 1)}>‹</button>
            {Array.from({ length: Math.min(pages, 7) }, (_, i) => i + 1).map(p => (
              <button key={p} style={s.pageBtn(p === page)} onClick={() => setPage(p)}>{p}</button>
            ))}
            <button style={s.pageBtn(false)} disabled={page >= pages} onClick={() => setPage(p => p + 1)}>›</button>
          </div>
        </>
      )}
    </div>
  )
}

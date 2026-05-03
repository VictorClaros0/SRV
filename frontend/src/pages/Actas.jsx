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

const SIM_FIELDS = [
  { key: 'codigoActa',        label: 'Código Acta'    },
  { key: 'codigoRecinto',     label: 'Cód. Recinto'   },
  { key: 'nroMesa',           label: 'Nro Mesa'       },
  { key: 'papeletasAnfora',   label: 'Pap. Ánfora'    },
  { key: 'papeletasNoUsadas', label: 'Pap. No Usadas' },
  { key: 'p1',                label: 'P1'             },
  { key: 'p2',                label: 'P2'             },
  { key: 'p3',                label: 'P3'             },
  { key: 'p4',                label: 'P4'             },
  { key: 'votosValidos',      label: 'Válidos'        },
  { key: 'votosNulos',        label: 'Nulos'          },
  { key: 'votosBlanco',       label: 'Blancos'        },
  { key: 'aperturaHora',      label: 'Ap. Hora'       },
  { key: 'aperturaMinutos',   label: 'Ap. Min'        },
  { key: 'cierreHora',        label: 'Ci. Hora'       },
  { key: 'cierreMinutos',     label: 'Ci. Min'        },
  { key: 'observaciones',     label: 'Observaciones'  },
]

const ANIM_SAMPLE = [
  { codigoActa: 1010200001001, codigoRecinto: 1010200001, nroMesa: 1,  papeletasAnfora: 788, papeletasNoUsadas:  89, p1: 140, p2:  39, p3: 124, p4: 345, votosValidos: 648, votosNulos: 64, votosBlanco: 76, aperturaHora: 8, aperturaMinutos:  0, cierreHora: 16, cierreMinutos: 30, observaciones: '' },
  { codigoActa: 1020300002015, codigoRecinto: 1020300002, nroMesa: 15, papeletasAnfora: 652, papeletasNoUsadas: 124, p1: 287, p2:  91, p3:  55, p4: 189, votosValidos: 622, votosNulos: 18, votosBlanco: 12, aperturaHora: 8, aperturaMinutos:  0, cierreHora: 17, cierreMinutos:  0, observaciones: '' },
  { codigoActa: 2040100003032, codigoRecinto: 2040100003, nroMesa: 32, papeletasAnfora: 512, papeletasNoUsadas: 201, p1:  68, p2: 143, p3: 201, p4:  76, votosValidos: 488, votosNulos: 15, votosBlanco:  9, aperturaHora: 8, aperturaMinutos: 30, cierreHora: 16, cierreMinutos: 45, observaciones: 'Tachadura en campo P3' },
  { codigoActa: 3050200004007, codigoRecinto: 3050200004, nroMesa: 7,  papeletasAnfora: 891, papeletasNoUsadas:  45, p1: 412, p2: 187, p3:  93, p4: 154, votosValidos: 846, votosNulos: 32, votosBlanco: 13, aperturaHora: 8, aperturaMinutos:  0, cierreHora: 17, cierreMinutos: 30, observaciones: '' },
  { codigoActa: 4060300005020, codigoRecinto: 4060300005, nroMesa: 20, papeletasAnfora: 445, papeletasNoUsadas: 178, p1:  56, p2: 312, p3:  48, p4:  21, votosValidos: 437, votosNulos:  5, votosBlanco:  3, aperturaHora: 8, aperturaMinutos:  0, cierreHora: 16, cierreMinutos:  0, observaciones: '' },
]

const sSim = {
  panel: {
    background: '#fff',
    border: '1.5px solid #27ae6040',
    borderRadius: 10,
    padding: '14px 18px',
    marginBottom: 14,
    boxShadow: '0 2px 8px rgba(39,174,96,.08)',
  },
  header: { display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12 },
  dot: { width: 8, height: 8, borderRadius: '50%', background: '#27ae60', display: 'inline-block', transition: 'opacity .3s' },
  title: { fontSize: 13, fontWeight: 700, color: '#1a1a2e' },
  counter: { fontSize: 12, color: '#888', fontWeight: 400 },
  grid: { display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(128px, 1fr))', gap: '8px 10px' },
  field: (active, done) => ({
    borderRadius: 6,
    padding: '6px 10px',
    border: active ? '1.5px solid #3498db' : done ? '1.5px solid #d5f5e3' : '1.5px solid #f0f0f0',
    background: active ? '#ebf5fb' : done ? '#f9fffe' : '#fafafa',
    transition: 'border-color .12s, background .12s',
  }),
  fieldLabel: { fontSize: 10, fontWeight: 700, color: '#999', textTransform: 'uppercase', letterSpacing: .3, marginBottom: 3 },
  fieldValue: (active, done) => ({
    fontSize: 13, fontWeight: 600, fontFamily: 'monospace',
    color: active ? '#1a6fa8' : done ? '#27ae60' : '#ccc',
    minHeight: 18,
    wordBreak: 'break-all',
  }),
}

function SimulacionPanel({ onComplete }) {
  const timerRef   = useRef(null)
  const stoppedRef = useRef(false)
  const stateRef   = useRef({ actaIdx: 0, fieldIdx: 0, charIdx: 0 })
  const [cursor, setCursor] = useState(true)
  const [dotOn,  setDotOn]  = useState(true)
  const [ui, setUi] = useState({ displayed: {}, activeField: 0, actaIdx: 0 })

  useEffect(() => {
    stoppedRef.current = false
    timerRef.current = setTimeout(tick, 20)
    const blinkId = setInterval(() => setCursor(v => !v), 530)
    const dotId   = setInterval(() => setDotOn(v => !v), 700)
    return () => {
      stoppedRef.current = true
      clearTimeout(timerRef.current)
      clearInterval(blinkId)
      clearInterval(dotId)
    }
  }, [])

  function tick() {
    if (stoppedRef.current) return
    const s = stateRef.current
    const acta  = ANIM_SAMPLE[s.actaIdx % ANIM_SAMPLE.length]
    const field = SIM_FIELDS[s.fieldIdx]
    const fullVal = String(acta[field.key] ?? '')

    if (s.charIdx < fullVal.length) {
      s.charIdx++
      const partial = fullVal.slice(0, s.charIdx)
      setUi(prev => ({ ...prev, displayed: { ...prev.displayed, [field.key]: partial } }))
      timerRef.current = setTimeout(tick, 20)
    } else if (s.fieldIdx < SIM_FIELDS.length - 1) {
      s.fieldIdx++
      s.charIdx = 0
      setUi(prev => ({ ...prev, activeField: s.fieldIdx }))
      timerRef.current = setTimeout(tick, 160)
    } else {
      onComplete(ANIM_SAMPLE[s.actaIdx % ANIM_SAMPLE.length])
      s.actaIdx++
      s.fieldIdx = 0
      s.charIdx  = 0
      setUi({ displayed: {}, activeField: 0, actaIdx: s.actaIdx })
      timerRef.current = setTimeout(tick, 350)
    }
  }

  const acta = ANIM_SAMPLE[ui.actaIdx % ANIM_SAMPLE.length]

  return (
    <div style={sSim.panel}>
      <div style={sSim.header}>
        <span style={{ ...sSim.dot, opacity: dotOn ? 1 : 0.2 }} />
        <span style={sSim.title}>Transcribiendo acta #{acta.codigoActa}</span>
        <span style={sSim.counter}>&nbsp;— operador ingresando datos</span>
      </div>
      <div style={sSim.grid}>
        {SIM_FIELDS.map((f, i) => {
          const isActive = i === ui.activeField
          const isDone   = i < ui.activeField
          const val      = ui.displayed[f.key] ?? ''
          return (
            <div key={f.key} style={sSim.field(isActive, isDone)}>
              <div style={sSim.fieldLabel}>{f.label}</div>
              <div style={sSim.fieldValue(isActive, isDone)}>
                {val || (!isActive && <span style={{ color: '#ddd' }}>—</span>)}
                {isActive && <span style={{ opacity: cursor ? 1 : 0, color: '#3498db', fontWeight: 300 }}>|</span>}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

export default function Actas() {
  const [actas, setActas] = useState([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [msg, setMsg] = useState({ texto: '', ok: true })
  const [simulando, setSimulando] = useState(false)
  const completedRef = useRef(new Map())
  const nav = useNavigate()

  function handleActaCompleted(sampleData) {
    setActas(prev => {
      const target = prev.find(a => a.estado === 'impresa' && !completedRef.current.has(a.id))
      if (!target) return prev
      const override = {
        ...target,
        estado: sampleData.observaciones ? 'observada' : 'transcrita',
        p1: sampleData.p1, p2: sampleData.p2, p3: sampleData.p3, p4: sampleData.p4,
        votosValidos: sampleData.votosValidos,
        votosNulos: sampleData.votosNulos,
        votosBlanco: sampleData.votosBlanco,
        observaciones: sampleData.observaciones,
      }
      completedRef.current.set(target.id, override)
      return prev.map(a => a.id === target.id ? override : a)
    })
  }

  useEffect(() => {
    loadActas()
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

  async function handleSimular() {
    setMsg({ texto: '', ok: true })
    completedRef.current.clear()
    setSimulando(true)
    setMsg({ texto: '▶ Transcribiendo actas...', ok: true })
    try {
      const result = await api.procesarTranscripciones()
      const data = await api.actas.list()
      setActas(data || [])
      setMsg({ texto: `✓ Completado: ${result.actualizadas} transcritas, ${result.observadas} observadas.`, ok: true })
    } catch (e) {
      setMsg({ texto: 'Error: ' + e.message, ok: false })
    } finally {
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
          onClick={handleSimular}
          disabled={simulando}
          title="Procesa todas las actas impresas y las transcribe a la base de datos"
        >
          {simulando ? '⏳ Transcribiendo...' : '▶ Simular transcripción'}
        </button>
      </div>

      {/* Contadores de estado */}
      <div style={s.contadores}>
        <span style={s.chip(ESTADO_COLOR.transcrita)}>Transcritas: {transcrita}</span>
        <span style={s.chip(ESTADO_COLOR.observada)}>Observadas: {observada}</span>
        <span style={{ fontSize: 13, color: '#888' }}>Total: {total}</span>
      </div>

      {/* Panel de transcripción animada */}
      {simulando && <SimulacionPanel onComplete={handleActaCompleted} />}

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

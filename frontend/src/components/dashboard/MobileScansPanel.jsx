import React, { useEffect, useState, useCallback } from 'react'
import {
  Chart as ChartJS, ArcElement, Tooltip, Legend,
  CategoryScale, LinearScale, BarElement,
} from 'chart.js'
import { Doughnut, Bar } from 'react-chartjs-2'
import { api } from '../../api.js'

ChartJS.register(ArcElement, Tooltip, Legend, CategoryScale, LinearScale, BarElement)

const ESTADO_COLORS = {
  validada:           '#22c55e',
  pendiente_revision: '#f97316',
  rechazada:          '#ef4444',
  observada:          '#a855f7',
}

const s = {
  card: { background: '#fff', borderRadius: 10, padding: '20px 24px', boxShadow: '0 2px 8px rgba(0,0,0,.06)' },
  header: { display: 'flex', alignItems: 'center', gap: 10, marginBottom: 4 },
  title: { fontSize: 14, fontWeight: 700, color: '#1a1a2e', flex: 1 },
  refreshBtn: { background: '#f0f0f0', border: 'none', borderRadius: 5, padding: '4px 10px', cursor: 'pointer', fontSize: 12 },
  note: { fontSize: 12, color: '#888', marginBottom: 14 },
  grid: { display: 'flex', gap: 16, alignItems: 'flex-start', flexWrap: 'wrap' },
  chartWrap: { width: 160, flexShrink: 0 },
  list: { flex: 1, display: 'flex', flexDirection: 'column', gap: 8, minWidth: 200 },
  item: { display: 'flex', gap: 10, alignItems: 'center', background: '#f8fafc', borderRadius: 8, padding: '8px 10px' },
  thumb: { width: 52, height: 52, objectFit: 'cover', borderRadius: 6, border: '1px solid #e2e8f0', flexShrink: 0 },
  thumbPlaceholder: { width: 52, height: 52, borderRadius: 6, background: '#e2e8f0', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 20, flexShrink: 0 },
  info: { flex: 1, minWidth: 0 },
  mesa: { fontSize: 13, fontWeight: 700, color: '#1a1a2e', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' },
  meta: { fontSize: 11, color: '#64748b', marginTop: 2 },
  badge: (estado) => ({
    fontSize: 10, fontWeight: 700, color: '#fff',
    background: ESTADO_COLORS[estado] || '#94a3b8',
    borderRadius: 4, padding: '2px 6px', display: 'inline-block', marginTop: 2,
  }),
  empty: { background: '#f8fafc', border: '1px solid #e2e8f0', borderRadius: 6, padding: '12px 16px', color: '#64748b', fontSize: 13, textAlign: 'center' },
  error: { color: '#ef4444', fontSize: 12, padding: '8px 0' },
  statsRow: { display: 'flex', gap: 10, flexWrap: 'wrap', marginBottom: 14 },
  statChip: (color) => ({ background: color + '18', border: `1px solid ${color}40`, borderRadius: 6, padding: '4px 10px', fontSize: 12, fontWeight: 700, color }),
  sectionLabel: { fontSize: 11, fontWeight: 700, color: '#64748b', textTransform: 'uppercase', letterSpacing: 0.4, marginTop: 16, marginBottom: 8 },
}

const DOUGHNUT_OPTS = {
  responsive: true,
  plugins: { legend: { display: false }, tooltip: { callbacks: { label: (ctx) => ` ${ctx.label}: ${ctx.parsed}` } } },
}

const BAR_OPTS = {
  responsive: true,
  plugins: { legend: { display: false }, tooltip: { callbacks: { label: (ctx) => ` ${ctx.parsed.y} votos` } } },
  scales: { y: { beginAtZero: true, ticks: { precision: 0 } } },
}

export default function MobileScansPanel() {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [lastUpdate, setLastUpdate] = useState(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const res = await api.dashboard.getMobileScans()
      setData(res)
      setLastUpdate(new Date().toLocaleTimeString('es-BO'))
    } catch (e) {
      setError('No se pudo cargar: ' + e.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const scans = data?.scans || []
  const total = data?.total || 0

  // Conteos por estado
  const counts = { validada: 0, pendiente_revision: 0, rechazada: 0, observada: 0 }
  scans.forEach(s => { if (s.estado in counts) counts[s.estado]++ })

  const doughnutData = {
    labels: ['Validadas', 'Pend. revisión', 'Rechazadas', 'Observadas'],
    datasets: [{ data: [counts.validada, counts.pendiente_revision, counts.rechazada, counts.observada], backgroundColor: [ESTADO_COLORS.validada, ESTADO_COLORS.pendiente_revision, ESTADO_COLORS.rechazada, ESTADO_COLORS.observada], borderWidth: 1 }],
  }

  // Votos totales por candidato
  const vt = data?.votos_totales || {}
  const candIds = Object.keys(vt).sort()
  const barData = candIds.length > 0 ? {
    labels: candIds.map(id => vt[id].nombre || id),
    datasets: [{
      data: candIds.map(id => vt[id].votos),
      backgroundColor: ['#3b82f6', '#f97316', '#22c55e', '#a855f7'],
      borderRadius: 4,
    }],
  } : null

  const totalVotos = candIds.reduce((a, id) => a + (vt[id]?.votos || 0), 0)

  return (
    <div style={s.card}>
      <div style={s.header}>
        <div style={s.title}>Escaneos desde App Móvil (Expo)</div>
        <button style={s.refreshBtn} onClick={load}>↺</button>
      </div>
      <div style={s.note}>
        {lastUpdate ? `Actualizado: ${lastUpdate} · ` : ''}{total} acta{total !== 1 ? 's' : ''} escaneada{total !== 1 ? 's' : ''}
      </div>

      {error && <div style={s.error}>{error}</div>}

      {/* Stats chips */}
      <div style={s.statsRow}>
        {Object.entries(counts).map(([k, v]) => (
          <span key={k} style={s.statChip(ESTADO_COLORS[k] || '#94a3b8')}>
            {k.replace('_', ' ').charAt(0).toUpperCase() + k.replace('_', ' ').slice(1)}: {v}
          </span>
        ))}
      </div>

      {loading ? (
        <div style={s.empty}>Cargando desde el servidor...</div>
      ) : total === 0 ? (
        <div style={s.empty}>
          📱 No hay escaneos registrados todavía.<br />
          <span style={{ fontSize: 11, marginTop: 4, display: 'block' }}>
            Abre la app Expo en tu celular y escanea un acta.
          </span>
        </div>
      ) : (
        <>
          {/* Lista + dona de estados */}
          <div style={s.grid}>
            <div style={s.chartWrap}>
              <Doughnut data={doughnutData} options={DOUGHNUT_OPTS} />
            </div>
            <div style={s.list}>
              {scans.slice(0, 6).map((scan, i) => {
                const hora = scan.fecha_recepcion
                  ? new Date(scan.fecha_recepcion).toLocaleTimeString('es-BO')
                  : '—'
                return (
                  <div key={scan.acta_id || i} style={s.item}>
                    {scan.imagen_url ? (
                      <a href={scan.imagen_url} target="_blank" rel="noopener noreferrer">
                        <img src={scan.imagen_url} alt="acta" style={s.thumb} />
                      </a>
                    ) : (
                      <div style={s.thumbPlaceholder}>📄</div>
                    )}
                    <div style={s.info}>
                      <div style={s.mesa}>Mesa {scan.mesa || 'S/N'}</div>
                      <div style={{ ...s.meta }}>{hora} · {scan.fuente_scan || '—'}</div>
                      <span style={s.badge(scan.estado)}>{(scan.estado || '').replace('_', ' ').toUpperCase()}</span>
                    </div>
                  </div>
                )
              })}
              {total > 6 && (
                <div style={{ fontSize: 12, color: '#888', textAlign: 'center' }}>+{total - 6} más en MongoDB</div>
              )}
            </div>
          </div>

          {/* Votos por candidato */}
          {barData && totalVotos > 0 && (
            <>
              <div style={s.sectionLabel}>
                Votos por candidato — {totalVotos.toLocaleString()} votos válidos escaneados
              </div>
              <Bar data={barData} options={BAR_OPTS} />
            </>
          )}
        </>
      )}
    </div>
  )
}

import React, { useState, useEffect } from 'react'
import { api } from '../../api.js'
import LoadingState from '../LoadingState.jsx'
import ErrorState from '../ErrorState.jsx'
import EmptyState from '../EmptyState.jsx'

const s = {
  card: {
    background: '#fff', borderRadius: 10, padding: '20px 24px',
    boxShadow: '0 2px 8px rgba(0,0,0,.06)',
  },
  header: { display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 },
  title: { fontSize: 14, fontWeight: 700, color: '#1a1a2e', flex: 1 },
  select: {
    padding: '6px 10px', border: '1.5px solid #ddd', borderRadius: 6,
    fontSize: 13, outline: 'none', background: '#fff',
  },
  table: { width: '100%', borderCollapse: 'collapse' },
  th: {
    background: '#f7f8fa', padding: '8px 12px', textAlign: 'left',
    fontSize: 11, fontWeight: 700, color: '#888', textTransform: 'uppercase',
  },
  td: { padding: '8px 12px', fontSize: 13, borderTop: '1px solid #f0f0f0' },
  badge: (color) => ({
    display: 'inline-block', padding: '1px 8px', borderRadius: 10,
    fontSize: 11, fontWeight: 600, background: color + '20', color,
  }),
}

export default function GeograficoPanel() {
  const [groupBy, setGroupBy] = useState('departamento')
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  async function load(gb) {
    setLoading(true)
    setError('')
    try {
      const res = await api.dashboard.getGeografico(gb)
      setData(res)
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load(groupBy) }, [groupBy])

  return (
    <div style={s.card}>
      <div style={s.header}>
        <div style={s.title}>Distribución geográfica</div>
        <select style={s.select} value={groupBy} onChange={e => setGroupBy(e.target.value)}>
          <option value="departamento">Por departamento</option>
          <option value="municipio">Por municipio</option>
          <option value="recinto">Por recinto</option>
        </select>
      </div>

      {loading && <LoadingState />}
      {!loading && error && <ErrorState message={error} onRetry={() => load(groupBy)} />}
      {!loading && !error && (!data?.items?.length) && <EmptyState message="Sin datos geográficos." />}
      {!loading && !error && data?.items?.length > 0 && (
        <div style={{ overflowX: 'auto' }}>
          <table style={s.table}>
            <thead>
              <tr>
                <th style={s.th}>{groupBy.charAt(0).toUpperCase() + groupBy.slice(1)}</th>
                <th style={s.th}>Total</th>
                <th style={s.th}>Consistentes</th>
                <th style={s.th}>Inconsistentes</th>
                <th style={s.th}>Solo RRV</th>
                <th style={s.th}>Solo Oficial</th>
                <th style={s.th}>Dif. Votos</th>
              </tr>
            </thead>
            <tbody>
              {data.items.map((item, i) => (
                <tr key={i}>
                  <td style={{ ...s.td, fontWeight: 600 }}>{item.nombre}</td>
                  <td style={s.td}>{item.total_actas}</td>
                  <td style={s.td}>
                    <span style={s.badge('#22c55e')}>{item.consistentes}</span>
                  </td>
                  <td style={s.td}>
                    <span style={s.badge('#ef4444')}>{item.inconsistentes}</span>
                  </td>
                  <td style={s.td}>
                    <span style={s.badge('#f97316')}>{item.solo_rrv}</span>
                  </td>
                  <td style={s.td}>
                    <span style={s.badge('#94a3b8')}>{item.solo_oficial}</span>
                  </td>
                  <td style={s.td}>{item.diferencia_total_votos}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

import React from 'react'
import EmptyState from '../EmptyState.jsx'

const s = {
  card: {
    background: '#fff', borderRadius: 10, padding: '20px 24px',
    boxShadow: '0 2px 8px rgba(0,0,0,.06)',
  },
  title: { fontSize: 14, fontWeight: 700, color: '#1a1a2e', marginBottom: 14 },
  summary: { display: 'flex', alignItems: 'baseline', gap: 8, marginBottom: 12 },
  total: { fontSize: 28, fontWeight: 800, color: '#ef4444' },
  muted: { fontSize: 12, color: '#777' },
  table: { width: '100%', borderCollapse: 'collapse', fontSize: 13 },
  th: { background: '#f7f8fa', padding: '7px 10px', textAlign: 'left', fontSize: 11, color: '#666' },
  td: { padding: '7px 10px', borderTop: '1px solid #f0f0f0', verticalAlign: 'top' },
}

export default function AuditoriaOficialPanel({ data }) {
  const actas = data?.actas || []

  return (
    <div style={s.card}>
      <div style={s.title}>Auditoria oficial</div>
      <div style={s.summary}>
        <span style={s.total}>{data?.total ?? actas.length}</span>
        <span style={s.muted}>actas observadas</span>
      </div>

      {actas.length === 0 && <EmptyState message="Sin actas observadas." />}
      {actas.length > 0 && (
        <table style={s.table}>
          <thead>
            <tr>
              <th style={s.th}>Acta</th>
              <th style={s.th}>Estado</th>
              <th style={s.th}>Observaciones</th>
            </tr>
          </thead>
          <tbody>
            {actas.slice(0, 5).map((acta) => (
              <tr key={acta.id || acta.codigoActa}>
                <td style={s.td}>{acta.codigoActa}</td>
                <td style={s.td}>{acta.estado}</td>
                <td style={s.td}>{acta.observaciones || '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}

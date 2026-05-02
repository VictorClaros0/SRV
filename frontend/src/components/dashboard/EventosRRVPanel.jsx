import React from 'react'
import EmptyState from '../EmptyState.jsx'

const s = {
  card: {
    background: '#fff', borderRadius: 10, padding: '20px 24px',
    boxShadow: '0 2px 8px rgba(0,0,0,.06)',
  },
  title: { fontSize: 14, fontWeight: 700, color: '#1a1a2e', marginBottom: 14 },
  meta: { fontSize: 12, color: '#777', marginBottom: 12 },
  table: { width: '100%', borderCollapse: 'collapse', fontSize: 13 },
  th: { background: '#f7f8fa', padding: '7px 10px', textAlign: 'left', fontSize: 11, color: '#666' },
  td: { padding: '7px 10px', borderTop: '1px solid #f0f0f0', verticalAlign: 'top' },
}

export default function EventosRRVPanel({ data }) {
  const eventos = data?.eventos || []

  return (
    <div style={s.card}>
      <div style={s.title}>Eventos RRV</div>
      <div style={s.meta}>
        {data?.total_eventos ?? eventos.length} eventos en {data?.coleccion || 'rrv_eventos'}
      </div>

      {eventos.length === 0 && <EmptyState message="Sin eventos RRV." />}
      {eventos.length > 0 && (
        <table style={s.table}>
          <thead>
            <tr>
              <th style={s.th}>Tipo</th>
              <th style={s.th}>Acta</th>
              <th style={s.th}>Fecha</th>
            </tr>
          </thead>
          <tbody>
            {eventos.slice(0, 6).map((evento, i) => (
              <tr key={`${evento.acta_id || 'evento'}-${i}`}>
                <td style={s.td}>{evento.tipo_evento}</td>
                <td style={s.td}>{evento.acta_id || '-'}</td>
                <td style={s.td}>
                  {evento.fecha_evento ? new Date(evento.fecha_evento).toLocaleString() : '-'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}

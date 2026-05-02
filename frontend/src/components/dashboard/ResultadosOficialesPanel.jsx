import React from 'react'
import EmptyState from '../EmptyState.jsx'

const s = {
  card: {
    background: '#fff', borderRadius: 10, padding: '20px 24px',
    boxShadow: '0 2px 8px rgba(0,0,0,.06)',
  },
  title: { fontSize: 14, fontWeight: 700, color: '#1a1a2e', marginBottom: 14 },
  stats: { display: 'grid', gridTemplateColumns: 'repeat(3, minmax(0, 1fr))', gap: 10, marginBottom: 14 },
  stat: { background: '#f7f8fa', borderRadius: 8, padding: '10px 12px' },
  label: { fontSize: 11, color: '#777', textTransform: 'uppercase', fontWeight: 700 },
  value: { fontSize: 18, color: '#1a1a2e', fontWeight: 800, marginTop: 4 },
  table: { width: '100%', borderCollapse: 'collapse', fontSize: 13 },
  th: { background: '#f7f8fa', padding: '7px 10px', textAlign: 'left', fontSize: 11, color: '#666' },
  td: { padding: '7px 10px', borderTop: '1px solid #f0f0f0' },
}

export default function ResultadosOficialesPanel({ data }) {
  if (!data) return <div style={s.card}><EmptyState message="Sin resultados oficiales." /></div>

  const resumen = data.resumen || {}
  const comparativa = data.comparativa || []
  const departamentos = (data.porDepartamento || []).slice(0, 5)

  return (
    <div style={s.card}>
      <div style={s.title}>Resultados oficiales</div>
      <div style={s.stats}>
        <div style={s.stat}>
          <div style={s.label}>Total actas</div>
          <div style={s.value}>{resumen.totalActas ?? 0}</div>
        </div>
        <div style={s.stat}>
          <div style={s.label}>Transcritas</div>
          <div style={s.value}>{resumen.actasTranscritas ?? 0}</div>
        </div>
        <div style={s.stat}>
          <div style={s.label}>Observadas</div>
          <div style={s.value}>{resumen.actasObservadas ?? 0}</div>
        </div>
      </div>

      {comparativa.length > 0 && (
        <table style={s.table}>
          <thead>
            <tr>
              <th style={s.th}>Candidato</th>
              <th style={s.th}>Votos</th>
            </tr>
          </thead>
          <tbody>
            {comparativa.map((item, i) => (
              <tr key={i}>
                <td style={s.td}>{item.candidato}</td>
                <td style={s.td}>{item.votos}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {departamentos.length > 0 && (
        <table style={{ ...s.table, marginTop: 12 }}>
          <thead>
            <tr>
              <th style={s.th}>Departamento</th>
              <th style={s.th}>Actas</th>
              <th style={s.th}>Validos</th>
            </tr>
          </thead>
          <tbody>
            {departamentos.map((item, i) => (
              <tr key={i}>
                <td style={s.td}>{item.ubicacion}</td>
                <td style={s.td}>{item.totalActas}</td>
                <td style={s.td}>{item.votosValidos}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}

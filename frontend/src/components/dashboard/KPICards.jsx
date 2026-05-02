import React from 'react'

const cards = [
  { key: 'total_actas_rrv',      label: 'Total Actas RRV',       color: '#f97316' },
  { key: 'total_actas_oficial',  label: 'Total Actas Oficial',   color: '#3b82f6' },
  { key: 'actas_solo_rrv',       label: 'Solo RRV',              color: '#f97316' },
  { key: 'actas_solo_oficial',   label: 'Solo Oficial',          color: '#94a3b8' },
  { key: 'confiabilidad_rrv',    label: 'Confiabilidad RRV (%)', color: '#1a1a2e', format: 'pct' },
]

const s = {
  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(180px, 1fr))',
    gap: 14,
    marginBottom: 24,
  },
  card: (color) => ({
    background: '#fff',
    borderRadius: 10,
    padding: '18px 20px',
    boxShadow: '0 2px 8px rgba(0,0,0,.06)',
    borderLeft: `4px solid ${color}`,
  }),
  label: { fontSize: 11, fontWeight: 700, color: '#888', textTransform: 'uppercase', letterSpacing: .4, marginBottom: 6 },
  value: (color) => ({ fontSize: 28, fontWeight: 800, color }),
}

export default function KPICards({ data }) {
  if (!data) return null
  return (
    <div style={s.grid}>
      {cards.map(({ key, label, color, format }) => {
        const raw = data[key]
        const val = format === 'pct'
          ? (raw != null ? raw.toFixed(1) + '%' : '—')
          : (raw != null ? raw.toLocaleString() : '—')
        return (
          <div key={key} style={s.card(color)}>
            <div style={s.label}>{label}</div>
            <div style={s.value(color)}>{val}</div>
          </div>
        )
      })}
    </div>
  )
}

import React from 'react'

const s = {
  wrap: {
    textAlign: 'center', padding: '48px 24px',
    color: '#aaa', fontSize: 14,
  },
  icon: { fontSize: 32, marginBottom: 8, display: 'block' },
}

export default function EmptyState({ message = 'Sin datos disponibles.', icon = '📭' }) {
  return (
    <div style={s.wrap}>
      <span style={s.icon}>{icon}</span>
      {message}
    </div>
  )
}

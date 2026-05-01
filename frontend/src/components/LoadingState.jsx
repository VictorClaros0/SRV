import React from 'react'

const s = {
  wrap: {
    display: 'flex', flexDirection: 'column', alignItems: 'center',
    justifyContent: 'center', padding: '48px 24px', gap: 12,
  },
  spinner: {
    width: 32, height: 32, border: '3px solid #e0e0e0',
    borderTopColor: '#1a1a2e', borderRadius: '50%',
    animation: 'spin 0.8s linear infinite',
  },
  text: { color: '#888', fontSize: 14 },
}

export default function LoadingState({ message = 'Cargando...' }) {
  return (
    <div style={s.wrap}>
      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      <div style={s.spinner} />
      <span style={s.text}>{message}</span>
    </div>
  )
}

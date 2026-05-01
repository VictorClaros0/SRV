import React from 'react'

const s = {
  wrap: {
    background: '#fff0f0', border: '1px solid #fac0c0', borderRadius: 8,
    padding: '20px 24px', display: 'flex', flexDirection: 'column', gap: 8,
  },
  title: { fontWeight: 700, color: '#c0392b', fontSize: 14 },
  msg: { color: '#7b2d2d', fontSize: 13 },
  retry: {
    alignSelf: 'flex-start', marginTop: 4,
    background: 'none', border: '1px solid #c0392b', color: '#c0392b',
    borderRadius: 6, padding: '5px 14px', cursor: 'pointer', fontSize: 13,
  },
}

export default function ErrorState({ message = 'Ocurrió un error.', onRetry }) {
  return (
    <div style={s.wrap}>
      <div style={s.title}>Error</div>
      <div style={s.msg}>{message}</div>
      {onRetry && (
        <button style={s.retry} onClick={onRetry}>Reintentar</button>
      )}
    </div>
  )
}

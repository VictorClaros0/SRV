import React from 'react'

const s = {
  card: {
    background: '#fff', borderRadius: 10, padding: '20px 24px',
    boxShadow: '0 2px 8px rgba(0,0,0,.06)',
  },
  title: { fontSize: 14, fontWeight: 700, color: '#1a1a2e', marginBottom: 16 },
  grid: { display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 },
  item: {
    background: '#f7f8fa', borderRadius: 8, padding: '12px 16px',
  },
  label: { fontSize: 11, fontWeight: 700, color: '#888', textTransform: 'uppercase', letterSpacing: .4, marginBottom: 4 },
  value: { fontSize: 13, fontWeight: 600, color: '#1a1a2e' },
  dot: (ok) => ({
    display: 'inline-block', width: 8, height: 8, borderRadius: '50%',
    background: ok ? '#22c55e' : '#ef4444', marginRight: 6,
  }),
}

function Item({ label, children }) {
  return (
    <div style={s.item}>
      <div style={s.label}>{label}</div>
      <div style={s.value}>{children}</div>
    </div>
  )
}

export default function TecnicoPanel({ data }) {
  if (!data) return null
  const { fuentes, seguridad, modulo_comparacion, latencia_ms, throughput_actas_min } = data
  const pgOk = fuentes?.postgresql?.estado === 'connected'
  const mgOk = fuentes?.mongodb?.estado === 'connected'

  return (
    <div style={s.card}>
      <div style={s.title}>Estado técnico del sistema</div>
      <div style={s.grid}>
        <Item label="PostgreSQL (Oficial)">
          <span style={s.dot(pgOk)} />
          {pgOk ? 'Conectado' : 'Error'}
          {fuentes?.postgresql?.total_actas != null && (
            <span style={{ color: '#888', fontWeight: 400 }}>
              {' '}— {fuentes.postgresql.total_actas.toLocaleString()} actas
            </span>
          )}
        </Item>
        <Item label="MongoDB (RRV)">
          <span style={s.dot(mgOk)} />
          {mgOk ? 'Conectado' : 'Error'}
          {fuentes?.mongodb?.total_actas != null && (
            <span style={{ color: '#888', fontWeight: 400 }}>
              {' '}— {fuentes.mongodb.total_actas.toLocaleString()} actas
            </span>
          )}
        </Item>
        <Item label="Autenticación">
          {seguridad?.jwt_required ? 'JWT requerido' : 'Sin autenticación'}
        </Item>
        <Item label="Módulo comparación">
          {modulo_comparacion?.cqrs_mode === 'query_only' ? 'CQRS — Solo lectura' : modulo_comparacion?.cqrs_mode || '—'}
        </Item>
        <Item label="Latencia">
          {latencia_ms != null ? `${latencia_ms} ms` : 'No disponible'}
        </Item>
        <Item label="Throughput">
          {throughput_actas_min != null ? `${throughput_actas_min} actas/min` : 'No disponible'}
        </Item>
      </div>
    </div>
  )
}

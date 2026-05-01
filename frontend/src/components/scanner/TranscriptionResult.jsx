import React from 'react'

const ESTADO_COLOR = {
  transcrita: '#22c55e',
  observada:  '#f97316',
  ERROR:      '#ef4444',
}

const s = {
  wrap: { marginTop: 20 },
  card: (estado) => ({
    background: '#fff', border: `2px solid ${ESTADO_COLOR[estado] || '#ddd'}`,
    borderRadius: 10, padding: '20px 24px',
    boxShadow: '0 2px 10px rgba(0,0,0,.06)',
  }),
  header: { display: 'flex', alignItems: 'center', gap: 10, marginBottom: 16 },
  estadoBadge: (estado) => ({
    display: 'inline-block', padding: '3px 12px', borderRadius: 20,
    fontSize: 12, fontWeight: 700,
    background: (ESTADO_COLOR[estado] || '#888') + '20',
    color: ESTADO_COLOR[estado] || '#888',
    textTransform: 'uppercase',
  }),
  title: { fontSize: 15, fontWeight: 700, color: '#1a1a2e' },
  message: { fontSize: 13, color: '#555', marginBottom: 12 },
  grid: { display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(140px, 1fr))', gap: '10px 16px' },
  field: {},
  fieldLabel: { fontSize: 11, color: '#888', fontWeight: 700, textTransform: 'uppercase', letterSpacing: .3, marginBottom: 2 },
  fieldValue: { fontSize: 14, fontWeight: 700, color: '#1a1a2e' },
  obs: {
    marginTop: 12, background: '#fffbeb', border: '1px solid #fde68a',
    borderRadius: 6, padding: '8px 12px', fontSize: 13, color: '#92400e',
  },
  actualizadaBadge: {
    display: 'inline-block', padding: '2px 10px', borderRadius: 10,
    background: '#f0fff4', color: '#22c55e', fontSize: 12, fontWeight: 600,
    border: '1px solid #bbf7d0', marginLeft: 8,
  },
  noActualizadaBadge: {
    display: 'inline-block', padding: '2px 10px', borderRadius: 10,
    background: '#f8fafc', color: '#94a3b8', fontSize: 12, fontWeight: 600,
    border: '1px solid #e2e8f0', marginLeft: 8,
  },
}

function Field({ label, value }) {
  return (
    <div style={s.field}>
      <div style={s.fieldLabel}>{label}</div>
      <div style={s.fieldValue}>{value ?? '—'}</div>
    </div>
  )
}

export default function TranscriptionResult({ result }) {
  if (!result) return null

  const { estado, message, acta_id, data } = result

  return (
    <div style={s.wrap}>
      <div style={s.card(estado)}>
        <div style={s.header}>
          <span style={s.estadoBadge(estado)}>{estado}</span>
          <span style={s.title}>Resultado de transcripción</span>
          {data?.acta_actualizada
            ? <span style={s.actualizadaBadge}>Acta actualizada en BD</span>
            : <span style={s.noActualizadaBadge}>Acta no encontrada en BD</span>
          }
        </div>

        <div style={s.message}>{message}</div>

        {data && (
          <>
            <div style={s.grid}>
              <Field label="Código Acta" value={acta_id || data.codigo_acta} />
              <Field label="P1" value={data.p1} />
              <Field label="P2" value={data.p2} />
              <Field label="P3" value={data.p3} />
              <Field label="P4" value={data.p4} />
              <Field label="Votos Válidos" value={data.votos_validos} />
              <Field label="Votos Nulos" value={data.votos_nulos} />
              <Field label="Votos Blancos" value={data.votos_blancos} />
              <Field label="Papeletas Ánfora" value={data.papeletas_anfora} />
            </div>

            {data.observaciones && (
              <div style={s.obs}>
                <strong>Observación:</strong> {data.observaciones}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}

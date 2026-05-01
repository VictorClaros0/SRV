import React, { useEffect, useState } from 'react'
import { api } from '../../api.js'
import LoadingState from '../LoadingState.jsx'
import ErrorState from '../ErrorState.jsx'

const ESTADO_COLOR = {
  CONSISTENTE:  '#22c55e',
  INCONSISTENTE:'#ef4444',
  SOLO_RRV:     '#f97316',
  SOLO_OFICIAL: '#94a3b8',
}

const SEVERIDAD_COLOR = {
  ALTA:  '#ef4444',
  MEDIA: '#f97316',
  BAJA:  '#eab308',
}

const s = {
  overlay: {
    position: 'fixed', inset: 0, background: 'rgba(0,0,0,.45)',
    display: 'flex', alignItems: 'center', justifyContent: 'center',
    zIndex: 1000, padding: 16,
  },
  modal: {
    background: '#fff', borderRadius: 12, width: '100%', maxWidth: 740,
    maxHeight: '90vh', overflow: 'auto',
    boxShadow: '0 8px 40px rgba(0,0,0,.18)',
  },
  header: {
    display: 'flex', alignItems: 'center', padding: '18px 24px',
    borderBottom: '1px solid #f0f0f0', gap: 12,
  },
  title: { fontSize: 16, fontWeight: 700, color: '#1a1a2e', flex: 1 },
  close: {
    background: 'none', border: 'none', fontSize: 20, cursor: 'pointer',
    color: '#888', lineHeight: 1,
  },
  body: { padding: '20px 24px' },
  section: { marginBottom: 20 },
  sectionTitle: {
    fontSize: 11, fontWeight: 700, color: '#888', textTransform: 'uppercase',
    letterSpacing: .5, marginBottom: 10, paddingBottom: 6, borderBottom: '1px solid #f0f0f0',
  },
  grid2: { display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px 20px' },
  field: {},
  fieldLabel: { fontSize: 11, color: '#888', marginBottom: 2 },
  fieldValue: { fontSize: 13, fontWeight: 600, color: '#1a1a2e' },
  badge: (color) => ({
    display: 'inline-block', padding: '2px 10px', borderRadius: 20,
    fontSize: 12, fontWeight: 700, background: color + '20', color,
  }),
  table: { width: '100%', borderCollapse: 'collapse', fontSize: 13 },
  th: { background: '#f7f8fa', padding: '6px 10px', textAlign: 'left', fontSize: 11, fontWeight: 700, color: '#666' },
  td: { padding: '6px 10px', borderTop: '1px solid #f0f0f0' },
  incCard: {
    background: '#fff8f8', border: '1px solid #fac0c0', borderRadius: 6,
    padding: '8px 12px', marginBottom: 6, fontSize: 13,
  },
}

function Field({ label, children }) {
  return (
    <div style={s.field}>
      <div style={s.fieldLabel}>{label}</div>
      <div style={s.fieldValue}>{children || '—'}</div>
    </div>
  )
}

export default function ModalDetalleActa({ actaId, onClose }) {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!actaId) return
    setLoading(true)
    setError('')
    api.comparacion.getActa(actaId)
      .then(setData)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }, [actaId])

  return (
    <div style={s.overlay} onClick={e => { if (e.target === e.currentTarget) onClose() }}>
      <div style={s.modal}>
        <div style={s.header}>
          <div style={s.title}>Detalle del acta — {actaId}</div>
          <button style={s.close} onClick={onClose}>×</button>
        </div>
        <div style={s.body}>
          {loading && <LoadingState message="Cargando detalle..." />}
          {!loading && error && <ErrorState message={error} />}
          {!loading && !error && data && (
            <>
              {/* Identificación */}
              <div style={s.section}>
                <div style={s.sectionTitle}>Identificación</div>
                <div style={s.grid2}>
                  <Field label="Acta ID">{data.acta_id}</Field>
                  <Field label="Estado comparación">
                    <span style={s.badge(ESTADO_COLOR[data.estado_comparacion] || '#888')}>
                      {data.estado_comparacion}
                    </span>
                  </Field>
                  <Field label="Departamento">{data.departamento}</Field>
                  <Field label="Municipio">{data.municipio}</Field>
                  <Field label="Recinto">{data.recinto}</Field>
                  <Field label="Mesa">{data.mesa}</Field>
                </div>
              </div>

              {/* Datos RRV */}
              {data.datos_rrv && (
                <div style={s.section}>
                  <div style={s.sectionTitle}>Datos RRV (fuente preliminar)</div>
                  <table style={s.table}>
                    <thead>
                      <tr>
                        <th style={s.th}>Candidato</th>
                        <th style={s.th}>Votos RRV</th>
                      </tr>
                    </thead>
                    <tbody>
                      {(data.datos_rrv.candidatos || []).map((c, i) => (
                        <tr key={i}>
                          <td style={s.td}>{c.nombre || c.candidato_id}</td>
                          <td style={s.td}>{c.votos}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  <div style={{ marginTop: 8, fontSize: 12, color: '#666' }}>
                    Nulos: {data.datos_rrv.votos_nulos} · Blancos: {data.datos_rrv.votos_blancos} · Total: {data.datos_rrv.total_votos}
                  </div>
                </div>
              )}

              {/* Datos Oficial */}
              {data.datos_oficial && (
                <div style={s.section}>
                  <div style={s.sectionTitle}>Datos Oficial (fuente auditada)</div>
                  <table style={s.table}>
                    <thead>
                      <tr>
                        <th style={s.th}>Candidato</th>
                        <th style={s.th}>Votos Oficial</th>
                      </tr>
                    </thead>
                    <tbody>
                      {(data.datos_oficial.candidatos || []).map((c, i) => (
                        <tr key={i}>
                          <td style={s.td}>{c.nombre || c.candidato_id}</td>
                          <td style={s.td}>{c.votos}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  <div style={{ marginTop: 8, fontSize: 12, color: '#666' }}>
                    Nulos: {data.datos_oficial.votos_nulos} · Blancos: {data.datos_oficial.votos_blancos} · Total: {data.datos_oficial.total_votos}
                  </div>
                </div>
              )}

              {/* Diferencias */}
              {data.diferencias && (
                <div style={s.section}>
                  <div style={s.sectionTitle}>Diferencias (RRV − Oficial)</div>
                  <div style={s.grid2}>
                    <Field label="Diferencia total votos">{data.diferencias.diferencia_total_votos}</Field>
                    <Field label="Diferencia votos nulos">{data.diferencias.diferencia_votos_nulos}</Field>
                    <Field label="Diferencia votos blancos">{data.diferencias.diferencia_votos_blancos}</Field>
                  </div>
                  {data.diferencias.diferencia_por_candidato?.length > 0 && (
                    <table style={{ ...s.table, marginTop: 10 }}>
                      <thead>
                        <tr>
                          <th style={s.th}>Candidato</th>
                          <th style={s.th}>Oficial</th>
                          <th style={s.th}>RRV</th>
                          <th style={s.th}>Diferencia</th>
                        </tr>
                      </thead>
                      <tbody>
                        {data.diferencias.diferencia_por_candidato.map((d, i) => (
                          <tr key={i}>
                            <td style={s.td}>{d.nombre || d.candidato_id}</td>
                            <td style={s.td}>{d.votos_oficial}</td>
                            <td style={s.td}>{d.votos_rrv}</td>
                            <td style={{ ...s.td, color: d.diferencia !== 0 ? '#ef4444' : '#22c55e', fontWeight: 600 }}>
                              {d.diferencia > 0 ? '+' : ''}{d.diferencia}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  )}
                </div>
              )}

              {/* Inconsistencias */}
              {data.inconsistencias?.length > 0 && (
                <div style={s.section}>
                  <div style={s.sectionTitle}>Inconsistencias detectadas ({data.inconsistencias.length})</div>
                  {data.inconsistencias.map((inc, i) => (
                    <div key={i} style={s.incCard}>
                      <span style={s.badge(SEVERIDAD_COLOR[inc.severidad] || '#888')}>
                        {inc.severidad}
                      </span>{' '}
                      <strong>{inc.tipo_inconsistencia}</strong> — {inc.descripcion}
                    </div>
                  ))}
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  )
}

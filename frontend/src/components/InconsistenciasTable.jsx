import React from 'react'
import EmptyState from './EmptyState.jsx'

const ESTADO_COLOR = {
  CONSISTENTE:   '#22c55e',
  INCONSISTENTE: '#ef4444',
  SOLO_RRV:      '#f97316',
  SOLO_OFICIAL:  '#94a3b8',
}

const SEVERIDAD_COLOR = {
  ALTA:  '#ef4444',
  MEDIA: '#f97316',
  BAJA:  '#eab308',
}

const s = {
  wrap: { overflowX: 'auto' },
  table: {
    width: '100%', borderCollapse: 'collapse', background: '#fff',
    borderRadius: 10, overflow: 'hidden', boxShadow: '0 2px 8px rgba(0,0,0,.06)',
    fontSize: 13,
  },
  th: {
    background: '#f7f8fa', padding: '10px 14px', textAlign: 'left',
    fontSize: 11, fontWeight: 700, color: '#666', textTransform: 'uppercase', letterSpacing: .4,
  },
  td: { padding: '10px 14px', borderTop: '1px solid #f0f0f0' },
  badge: (color) => ({
    display: 'inline-block', padding: '2px 8px', borderRadius: 12,
    fontSize: 11, fontWeight: 700, background: color + '20', color,
  }),
  btnDetalle: {
    background: 'none', border: '1px solid #1a1a2e', color: '#1a1a2e',
    borderRadius: 5, padding: '3px 10px', cursor: 'pointer', fontSize: 12, fontWeight: 600,
  },
  pagination: { display: 'flex', gap: 8, marginTop: 14, alignItems: 'center', justifyContent: 'flex-end' },
  pageBtn: (active) => ({
    padding: '5px 11px', borderRadius: 5, border: '1px solid #ddd',
    background: active ? '#1a1a2e' : '#fff', color: active ? '#fff' : '#444',
    cursor: 'pointer', fontSize: 13,
  }),
  info: { fontSize: 12, color: '#888' },
}

export default function InconsistenciasTable({ rows = [], total, pagina, porPagina, onPageChange, onVerDetalle }) {
  if (!rows.length) return <EmptyState message="No hay inconsistencias con los filtros actuales." icon="✅" />

  const totalPages = Math.ceil(total / porPagina) || 1

  return (
    <div>
      <div style={s.wrap}>
        <table style={s.table}>
          <thead>
            <tr>
              <th style={s.th}>Acta ID</th>
              <th style={s.th}>Departamento</th>
              <th style={s.th}>Municipio</th>
              <th style={s.th}>Recinto</th>
              <th style={s.th}>Mesa</th>
              <th style={s.th}>Estado</th>
              <th style={s.th}>Dif. Votos</th>
              <th style={s.th}>Severidad</th>
              <th style={s.th}></th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row, i) => {
              const color = ESTADO_COLOR[row.estado_comparacion] || '#888'
              const sev = row.inconsistencias?.[0]?.severidad
              const sevColor = SEVERIDAD_COLOR[sev] || '#888'
              const difVotos = row.diferencias?.diferencia_total_votos ?? '—'
              return (
                <tr key={row.acta_id || i}>
                  <td style={{ ...s.td, fontWeight: 600 }}>{row.acta_id}</td>
                  <td style={s.td}>{row.departamento || '—'}</td>
                  <td style={s.td}>{row.municipio || '—'}</td>
                  <td style={s.td}>{row.recinto || '—'}</td>
                  <td style={s.td}>{row.mesa || '—'}</td>
                  <td style={s.td}>
                    <span style={s.badge(color)}>{row.estado_comparacion}</span>
                  </td>
                  <td style={s.td}>{difVotos}</td>
                  <td style={s.td}>
                    {sev ? <span style={s.badge(sevColor)}>{sev}</span> : '—'}
                  </td>
                  <td style={s.td}>
                    <button style={s.btnDetalle} onClick={() => onVerDetalle?.(row.acta_id)}>
                      Ver detalle
                    </button>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div style={s.pagination}>
          <span style={s.info}>{total} resultado(s) — pág. {pagina} de {totalPages}</span>
          <button style={s.pageBtn(false)} disabled={pagina === 1} onClick={() => onPageChange(pagina - 1)}>‹</button>
          {Array.from({ length: Math.min(totalPages, 7) }, (_, i) => i + 1).map(p => (
            <button key={p} style={s.pageBtn(p === pagina)} onClick={() => onPageChange(p)}>{p}</button>
          ))}
          <button style={s.pageBtn(false)} disabled={pagina >= totalPages} onClick={() => onPageChange(pagina + 1)}>›</button>
        </div>
      )}
    </div>
  )
}

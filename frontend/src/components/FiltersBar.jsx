import React from 'react'

const ESTADOS = ['', 'CONSISTENTE', 'INCONSISTENTE', 'SOLO_RRV', 'SOLO_OFICIAL']

const s = {
  bar: {
    background: '#fff', borderRadius: 10, padding: '14px 20px',
    boxShadow: '0 2px 8px rgba(0,0,0,.06)', marginBottom: 16,
    display: 'flex', flexWrap: 'wrap', gap: 10, alignItems: 'flex-end',
  },
  group: { display: 'flex', flexDirection: 'column', gap: 4 },
  label: { fontSize: 11, fontWeight: 700, color: '#888', textTransform: 'uppercase', letterSpacing: .4 },
  input: {
    padding: '7px 12px', border: '1.5px solid #ddd', borderRadius: 6,
    fontSize: 13, outline: 'none', minWidth: 140,
  },
  select: {
    padding: '7px 12px', border: '1.5px solid #ddd', borderRadius: 6,
    fontSize: 13, outline: 'none', background: '#fff', minWidth: 140,
  },
  btnApply: {
    background: '#1a1a2e', color: '#fff', border: 'none', borderRadius: 6,
    padding: '8px 18px', cursor: 'pointer', fontSize: 13, fontWeight: 600,
    alignSelf: 'flex-end',
  },
  btnClear: {
    background: '#f0f0f0', color: '#444', border: 'none', borderRadius: 6,
    padding: '8px 14px', cursor: 'pointer', fontSize: 13, alignSelf: 'flex-end',
  },
}

export default function FiltersBar({ filters, onChange, onApply, onClear }) {
  function set(field) {
    return e => onChange({ ...filters, [field]: e.target.value, pagina: 1 })
  }

  return (
    <div style={s.bar}>
      <div style={s.group}>
        <label style={s.label}>Estado</label>
        <select style={s.select} value={filters.estado || ''} onChange={set('estado')}>
          {ESTADOS.map(e => (
            <option key={e} value={e}>{e || 'Todos'}</option>
          ))}
        </select>
      </div>

      <div style={s.group}>
        <label style={s.label}>Departamento</label>
        <input
          style={s.input}
          placeholder="Ej. La Paz"
          value={filters.departamento || ''}
          onChange={set('departamento')}
        />
      </div>

      <div style={s.group}>
        <label style={s.label}>Municipio</label>
        <input
          style={s.input}
          placeholder="Ej. El Alto"
          value={filters.municipio || ''}
          onChange={set('municipio')}
        />
      </div>

      <div style={s.group}>
        <label style={s.label}>Recinto</label>
        <input
          style={s.input}
          placeholder="Ej. Colegio Ayacucho"
          value={filters.recinto || ''}
          onChange={set('recinto')}
        />
      </div>

      <div style={s.group}>
        <label style={s.label}>Resultados por página</label>
        <select style={s.select} value={filters.por_pagina || 20} onChange={set('por_pagina')}>
          {[10, 20, 50, 100].map(n => <option key={n} value={n}>{n}</option>)}
        </select>
      </div>

      <button style={s.btnApply} onClick={onApply}>Filtrar</button>
      <button style={s.btnClear} onClick={onClear}>Limpiar</button>
    </div>
  )
}

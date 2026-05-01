import React, { useState, useCallback, useEffect } from 'react'
import { api } from '../api.js'
import LoadingState from '../components/LoadingState.jsx'
import ErrorState from '../components/ErrorState.jsx'
import FiltersBar from '../components/FiltersBar.jsx'
import InconsistenciasTable from '../components/InconsistenciasTable.jsx'
import ModalDetalleActa from '../components/dashboard/ModalDetalleActa.jsx'

const INIT_FILTERS = {
  estado: '',
  departamento: '',
  municipio: '',
  recinto: '',
  pagina: 1,
  por_pagina: 20,
}

const s = {
  header: { display: 'flex', alignItems: 'center', marginBottom: 20, gap: 12 },
  title: { fontSize: 22, fontWeight: 800, color: '#1a1a2e', flex: 1 },
  resumen: {
    display: 'flex', gap: 16, marginBottom: 16, flexWrap: 'wrap',
  },
  chip: (color) => ({
    padding: '4px 14px', borderRadius: 20, fontSize: 13, fontWeight: 600,
    background: color + '20', color, border: '1px solid ' + color + '40',
  }),
}

export default function InconsistenciasPage() {
  const [filters, setFilters] = useState(INIT_FILTERS)
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [modalActaId, setModalActaId] = useState(null)

  const load = useCallback(async (f) => {
    setLoading(true)
    setError('')
    try {
      const params = {}
      if (f.estado) params.estado = f.estado
      if (f.departamento) params.departamento = f.departamento
      if (f.municipio) params.municipio = f.municipio
      if (f.recinto) params.recinto = f.recinto
      params.pagina = f.pagina || 1
      params.por_pagina = f.por_pagina || 20

      const res = await api.comparacion.list(params)
      setData(res)
    } catch (e) {
      setError(e.message || 'Error al cargar datos.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load(INIT_FILTERS) }, [load])

  function handleApply() { load(filters) }
  function handleClear() {
    const clean = INIT_FILTERS
    setFilters(clean)
    load(clean)
  }
  function handlePage(p) {
    const updated = { ...filters, pagina: p }
    setFilters(updated)
    load(updated)
  }

  const resumen = data?.resumen
  const rows = data?.actas || []
  const total = data?.total || 0
  const pagina = data?.pagina || 1
  const porPagina = data?.por_pagina || 20

  return (
    <div>
      <div style={s.header}>
        <div style={s.title}>Inconsistencias — Comparación RRV vs Oficial</div>
      </div>

      {/* Resumen rápido */}
      {resumen && (
        <div style={s.resumen}>
          <span style={s.chip('#22c55e')}>Consistentes: {resumen.actas_consistentes}</span>
          <span style={s.chip('#ef4444')}>Inconsistentes: {resumen.actas_inconsistentes}</span>
          <span style={s.chip('#f97316')}>Solo RRV: {resumen.actas_solo_rrv}</span>
          <span style={s.chip('#94a3b8')}>Solo Oficial: {resumen.actas_solo_oficial}</span>
          <span style={{ fontSize: 13, color: '#888', alignSelf: 'center' }}>
            Confiabilidad RRV: {resumen.confiabilidad_rrv?.toFixed(1)}%
          </span>
        </div>
      )}

      <FiltersBar
        filters={filters}
        onChange={setFilters}
        onApply={handleApply}
        onClear={handleClear}
      />

      {loading && <LoadingState message="Cargando comparaciones..." />}
      {!loading && error && <ErrorState message={error} onRetry={() => load(filters)} />}
      {!loading && !error && (
        <InconsistenciasTable
          rows={rows}
          total={total}
          pagina={pagina}
          porPagina={porPagina}
          onPageChange={handlePage}
          onVerDetalle={setModalActaId}
        />
      )}

      {modalActaId && (
        <ModalDetalleActa actaId={modalActaId} onClose={() => setModalActaId(null)} />
      )}
    </div>
  )
}

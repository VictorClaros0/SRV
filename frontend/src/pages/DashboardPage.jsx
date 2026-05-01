import React, { useEffect, useState, useCallback } from 'react'
import { api } from '../api.js'
import LoadingState from '../components/LoadingState.jsx'
import ErrorState from '../components/ErrorState.jsx'
import KPICards from '../components/dashboard/KPICards.jsx'
import RRVvsOficialChart from '../components/dashboard/RRVvsOficialChart.jsx'
import VotosCandidatoChart from '../components/dashboard/VotosCandidatoChart.jsx'
import ParticipacionChart from '../components/dashboard/ParticipacionChart.jsx'
import GeograficoPanel from '../components/dashboard/GeograficoPanel.jsx'
import TecnicoPanel from '../components/dashboard/TecnicoPanel.jsx'
import InconsistenciasTable from '../components/InconsistenciasTable.jsx'
import ModalDetalleActa from '../components/dashboard/ModalDetalleActa.jsx'

const s = {
  page: {},
  header: { display: 'flex', alignItems: 'center', marginBottom: 24, gap: 12 },
  title: { fontSize: 22, fontWeight: 800, color: '#1a1a2e', flex: 1 },
  refreshBtn: {
    background: '#f0f0f0', color: '#444', border: 'none', borderRadius: 6,
    padding: '7px 16px', cursor: 'pointer', fontSize: 13,
  },
  grid2: { display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginBottom: 16 },
  sectionTitle: {
    fontSize: 15, fontWeight: 700, color: '#1a1a2e', marginBottom: 12, marginTop: 8,
  },
  lastUpdate: { fontSize: 12, color: '#888' },
}

export default function DashboardPage() {
  const [kpis, setKpis] = useState(null)
  const [rrv, setRrv] = useState(null)
  const [votos, setVotos] = useState(null)
  const [part, setPart] = useState(null)
  const [tecnico, setTecnico] = useState(null)
  const [inconsistencias, setInconsistencias] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [modalActaId, setModalActaId] = useState(null)
  const [lastUpdate, setLastUpdate] = useState(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [k, r, v, p, t, inc] = await Promise.all([
        api.dashboard.getKPIs(),
        api.dashboard.getRRVvsOficial(),
        api.dashboard.getVotosCandidato(),
        api.dashboard.getParticipacion(),
        api.dashboard.getTecnico(),
        api.dashboard.getInconsistencias(),
      ])
      setKpis(k)
      setRrv(r)
      setVotos(v)
      setPart(p)
      setTecnico(t)
      setInconsistencias(inc)
      setLastUpdate(new Date().toLocaleTimeString())
    } catch (e) {
      setError(e.message || 'No se pudo conectar con el backend.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  if (loading) return <LoadingState message="Cargando dashboard..." />
  if (error) return <ErrorState message={error} onRetry={load} />

  const incRows = inconsistencias?.inconsistencias || []

  return (
    <div style={s.page}>
      <div style={s.header}>
        <div>
          <div style={s.title}>Dashboard Electoral</div>
          {lastUpdate && <div style={s.lastUpdate}>Actualizado: {lastUpdate}</div>}
        </div>
        <button style={s.refreshBtn} onClick={load}>↺ Actualizar</button>
      </div>

      {/* KPIs */}
      <KPICards data={kpis} />

      {/* RRV vs Oficial + Votos por candidato */}
      <div style={s.grid2}>
        <RRVvsOficialChart data={rrv} />
        <VotosCandidatoChart data={votos} />
      </div>

      {/* Participación + Técnico */}
      <div style={s.grid2}>
        <ParticipacionChart data={part} />
        <TecnicoPanel data={tecnico} />
      </div>

      {/* Geográfico — maneja su propio estado */}
      <div style={{ marginBottom: 16 }}>
        <GeograficoPanel />
      </div>

      {/* Inconsistencias */}
      <div style={s.sectionTitle}>
        Inconsistencias detectadas
        {incRows.length > 0 && (
          <span style={{ fontSize: 13, color: '#888', fontWeight: 400, marginLeft: 8 }}>
            ({incRows.length})
          </span>
        )}
      </div>
      <InconsistenciasTable
        rows={incRows}
        total={incRows.length}
        pagina={1}
        porPagina={incRows.length || 1}
        onVerDetalle={setModalActaId}
      />

      {modalActaId && (
        <ModalDetalleActa actaId={modalActaId} onClose={() => setModalActaId(null)} />
      )}
    </div>
  )
}

import React, { useEffect, useState, useCallback } from 'react'
import { api } from '../api.js'
import LoadingState from '../components/LoadingState.jsx'
import ErrorState from '../components/ErrorState.jsx'
import KPICards from '../components/dashboard/KPICards.jsx'
import RRVvsOficialChart from '../components/dashboard/RRVvsOficialChart.jsx'
import VotosCandidatoChart from '../components/dashboard/VotosCandidatoChart.jsx'
import VotosDistribucionChart from '../components/dashboard/VotosDistribucionChart.jsx'
import MobileScansPanel from '../components/dashboard/MobileScansPanel.jsx'
import GeograficoPanel from '../components/dashboard/GeograficoPanel.jsx'
import TecnicoPanel from '../components/dashboard/TecnicoPanel.jsx'
import ResultadosOficialesPanel from '../components/dashboard/ResultadosOficialesPanel.jsx'
import AuditoriaOficialPanel from '../components/dashboard/AuditoriaOficialPanel.jsx'
import EventosRRVPanel from '../components/dashboard/EventosRRVPanel.jsx'

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
  const [tecnico, setTecnico] = useState(null)
  const [resultadosOficiales, setResultadosOficiales] = useState(null)
  const [auditoriaOficial, setAuditoriaOficial] = useState(null)
  const [eventosRRV, setEventosRRV] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [lastUpdate, setLastUpdate] = useState(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [k, r, v, t, ro, ao, ev] = await Promise.all([
        api.dashboard.getKPIs(),
        api.dashboard.getRRVvsOficial(),
        api.dashboard.getVotosCandidato(),
        api.dashboard.getTecnico(),
        api.oficial.getResultados(),
        api.oficial.getAuditoria(),
        api.dashboard.getEventosRRV(),
      ])
      setKpis(k)
      setRrv(r)
      setVotos(v)
      setTecnico(t)
      setResultadosOficiales(ro)
      setAuditoriaOficial(ao)
      setEventosRRV(ev)
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

      {/* Distribución de votos + Técnico */}
      <div style={s.grid2}>
        <VotosDistribucionChart data={votos} />
        <TecnicoPanel data={tecnico} />
      </div>

      {/* Escaneos desde app móvil (Firebase) */}
      <div style={{ marginBottom: 16 }}>
        <MobileScansPanel />
      </div>

      {/* Geográfico — maneja su propio estado */}
      <div style={{ marginBottom: 16 }}>
        <GeograficoPanel />
      </div>

      <div style={s.grid2}>
        <ResultadosOficialesPanel data={resultadosOficiales} />
        <AuditoriaOficialPanel data={auditoriaOficial} />
      </div>

      <div style={{ marginBottom: 16 }}>
        <EventosRRVPanel data={eventosRRV} />
      </div>
    </div>
  )
}

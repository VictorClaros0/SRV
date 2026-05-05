import React from 'react'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { Doughnut } from 'react-chartjs-2'

ChartJS.register(ArcElement, Tooltip, Legend)

const s = {
  card: {
    background: '#fff', borderRadius: 10, padding: '20px 24px',
    boxShadow: '0 2px 8px rgba(0,0,0,.06)',
  },
  title: { fontSize: 14, fontWeight: 700, color: '#1a1a2e', marginBottom: 4 },
  note: { fontSize: 12, color: '#888', marginBottom: 16 },
  wrap: { display: 'flex', justifyContent: 'center', maxHeight: 260 },
  empty: {
    background: '#f8fafc', border: '1px solid #e2e8f0', borderRadius: 6,
    padding: '12px 16px', color: '#64748b', fontSize: 13,
  },
}

const OPTIONS = {
  responsive: true,
  plugins: {
    legend: { position: 'bottom' },
    tooltip: {
      callbacks: {
        label: (ctx) => {
          const total = ctx.dataset.data.reduce((a, b) => a + b, 0)
          const pct = total > 0 ? ((ctx.parsed / total) * 100).toFixed(1) : 0
          return ` ${ctx.label}: ${ctx.parsed.toLocaleString()} (${pct}%)`
        },
      },
    },
  },
}

export default function VotosDistribucionChart({ data }) {
  if (!data) return null

  // data viene del endpoint votos-candidato; extrae totales RRV
  const rrv = data.datasets?.find(ds => /rrv/i.test(ds.label))
  const oficial = data.datasets?.find(ds => /oficial/i.test(ds.label))
  const source = rrv || oficial
  if (!source) {
    return (
      <div style={s.card}>
        <div style={s.title}>Distribución de votos (RRV)</div>
        <div style={s.empty}>Sin datos disponibles aún.</div>
      </div>
    )
  }

  const labels = data.labels || []
  const values = source.data || []
  const total = values.reduce((a, b) => a + b, 0)

  const chartData = {
    labels,
    datasets: [{
      data: values,
      backgroundColor: ['#3b82f6', '#f97316', '#22c55e', '#a855f7', '#ef4444', '#eab308'],
      borderWidth: 1,
    }],
  }

  return (
    <div style={s.card}>
      <div style={s.title}>Distribución de votos — RRV</div>
      <div style={s.note}>Total: {total.toLocaleString()} votos</div>
      {total === 0 ? (
        <div style={s.empty}>Sin votos registrados aún en el RRV.</div>
      ) : (
        <div style={s.wrap}>
          <Doughnut data={chartData} options={OPTIONS} />
        </div>
      )}
    </div>
  )
}

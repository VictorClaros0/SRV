import React from 'react'
import {
  Chart as ChartJS,
  ArcElement,
  Tooltip,
  Legend,
} from 'chart.js'
import { Doughnut } from 'react-chartjs-2'

ChartJS.register(ArcElement, Tooltip, Legend)

const s = {
  card: {
    background: '#fff', borderRadius: 10, padding: '20px 24px',
    boxShadow: '0 2px 8px rgba(0,0,0,.06)',
  },
  title: { fontSize: 14, fontWeight: 700, color: '#1a1a2e', marginBottom: 16 },
  wrap: { maxWidth: 320, margin: '0 auto' },
}

const OPTIONS = {
  responsive: true,
  plugins: {
    legend: { position: 'bottom' },
    tooltip: { callbacks: { label: (ctx) => ` ${ctx.label}: ${ctx.parsed}` } },
  },
}

export default function RRVvsOficialChart({ data }) {
  if (!data) return null
  return (
    <div style={s.card}>
      <div style={s.title}>RRV vs Oficial — Distribución de estados</div>
      <div style={s.wrap}>
        <Doughnut data={data} options={OPTIONS} />
      </div>
    </div>
  )
}

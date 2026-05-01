import React from 'react'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Tooltip,
  Legend,
} from 'chart.js'
import { Bar } from 'react-chartjs-2'

ChartJS.register(CategoryScale, LinearScale, BarElement, Tooltip, Legend)

const s = {
  card: {
    background: '#fff', borderRadius: 10, padding: '20px 24px',
    boxShadow: '0 2px 8px rgba(0,0,0,.06)',
  },
  title: { fontSize: 14, fontWeight: 700, color: '#1a1a2e', marginBottom: 12 },
  notice: {
    background: '#f8fafc', border: '1px solid #e2e8f0', borderRadius: 6,
    padding: '12px 16px', color: '#64748b', fontSize: 13,
  },
}

const OPTIONS = {
  responsive: true,
  plugins: { legend: { position: 'top' } },
  scales: { y: { beginAtZero: true, max: 100, ticks: { callback: v => v + '%' } } },
}

export default function ParticipacionChart({ data }) {
  if (!data) return null
  return (
    <div style={s.card}>
      <div style={s.title}>Participación electoral</div>
      {!data.available ? (
        <div style={s.notice}>{data.message}</div>
      ) : (
        <Bar data={{ labels: data.labels, datasets: data.datasets }} options={OPTIONS} />
      )}
    </div>
  )
}

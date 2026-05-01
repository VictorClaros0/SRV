import React from 'react'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js'
import { Bar } from 'react-chartjs-2'

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend)

const s = {
  card: {
    background: '#fff', borderRadius: 10, padding: '20px 24px',
    boxShadow: '0 2px 8px rgba(0,0,0,.06)',
  },
  title: { fontSize: 14, fontWeight: 700, color: '#1a1a2e', marginBottom: 4 },
  note: { fontSize: 12, color: '#888', marginBottom: 16 },
}

const OPTIONS = {
  responsive: true,
  plugins: {
    legend: { position: 'top' },
    title: { display: false },
  },
  scales: {
    y: { beginAtZero: true, ticks: { precision: 0 } },
  },
}

export default function VotosCandidatoChart({ data }) {
  if (!data) return null

  const todosEnCero = data.datasets?.every(ds => ds.data?.every(v => v === 0))

  return (
    <div style={s.card}>
      <div style={s.title}>Votos por candidato — RRV vs Oficial</div>
      {todosEnCero && (
        <div style={s.note}>
          Los votos oficiales dependen de la transcripción/carga oficial procesada.
        </div>
      )}
      <Bar data={data} options={OPTIONS} />
    </div>
  )
}

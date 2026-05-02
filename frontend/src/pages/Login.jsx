import React, { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api.js'

const s = {
  page: {
    minHeight: '100vh', display: 'flex', alignItems: 'center',
    justifyContent: 'center', background: '#f0f2f5',
  },
  card: {
    background: '#fff', borderRadius: 12, padding: '48px 40px',
    boxShadow: '0 4px 24px rgba(0,0,0,0.10)', width: 360,
  },
  title: { fontSize: 26, fontWeight: 700, color: '#1a1a2e', marginBottom: 6 },
  sub: { color: '#888', fontSize: 14, marginBottom: 32 },
  label: { display: 'block', fontSize: 13, fontWeight: 600, color: '#444', marginBottom: 6 },
  input: {
    width: '100%', padding: '10px 14px', border: '1.5px solid #ddd',
    borderRadius: 8, fontSize: 15, outline: 'none', marginBottom: 18,
    transition: 'border-color .2s',
  },
  btn: {
    width: '100%', padding: '12px 0', background: '#1a1a2e', color: '#fff',
    border: 'none', borderRadius: 8, fontSize: 15, fontWeight: 600,
    cursor: 'pointer', marginTop: 4,
  },
  error: {
    background: '#fff0f0', color: '#c0392b', border: '1px solid #fac0c0',
    borderRadius: 6, padding: '10px 14px', fontSize: 13, marginBottom: 16,
  },
}

export default function Login() {
  const [usuario, setUsuario] = useState('admin')
  const [contrasena, setContrasena] = useState('admin123')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const nav = useNavigate()

  useEffect(() => {
    handleSubmit({ preventDefault: () => {} })
  }, [])

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const data = await api.login(usuario, contrasena)
      localStorage.setItem('token', data.token)
      nav('/actas')
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={s.page}>
      <div style={s.card}>
        <div style={s.title}>SROV</div>
        <div style={s.sub}>Sistema de Registro de Votos</div>
        {error && <div style={s.error}>{error}</div>}
        <form onSubmit={handleSubmit}>
          <label style={s.label}>Usuario</label>
          <input
            style={s.input}
            value={usuario}
            onChange={e => setUsuario(e.target.value)}
            autoFocus
            required
          />
          <label style={s.label}>Contraseña</label>
          <input
            style={s.input}
            type="password"
            value={contrasena}
            onChange={e => setContrasena(e.target.value)}
            required
          />
          <button style={s.btn} disabled={loading}>
            {loading ? 'Ingresando...' : 'Ingresar'}
          </button>
        </form>
      </div>
    </div>
  )
}

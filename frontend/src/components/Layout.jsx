import React from 'react'
import { Outlet, NavLink, useNavigate } from 'react-router-dom'

const s = {
  shell: { display: 'flex', minHeight: '100vh', flexDirection: 'column' },
  nav: {
    background: '#1a1a2e', color: '#fff', display: 'flex',
    alignItems: 'center', padding: '0 24px', height: 56,
    gap: 24,
  },
  brand: { fontWeight: 700, fontSize: 18, letterSpacing: 1, color: '#e0e0e0', marginRight: 'auto' },
  link: { color: '#aaa', textDecoration: 'none', fontSize: 14, padding: '4px 8px', borderRadius: 4 },
  activeLink: { color: '#fff', background: '#2d2d5e' },
  logout: {
    background: 'none', border: '1px solid #555', color: '#ccc',
    cursor: 'pointer', borderRadius: 4, padding: '4px 12px', fontSize: 13,
  },
  main: { flex: 1, padding: 24 },
}

export default function Layout() {
  const nav = useNavigate()
  function logout() {
    localStorage.removeItem('token')
    nav('/login')
  }
  return (
    <div style={s.shell}>
      <nav style={s.nav}>
        <span style={s.brand}>SROV</span>
        <NavLink to="/actas" style={({ isActive }) => ({ ...s.link, ...(isActive ? s.activeLink : {}) })}>
          Actas
        </NavLink>
        <button style={s.logout} onClick={logout}>Cerrar sesión</button>
      </nav>
      <main style={s.main}>
        <Outlet />
      </main>
    </div>
  )
}

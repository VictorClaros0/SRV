import React from 'react'
import { Outlet, NavLink, useNavigate } from 'react-router-dom'

const s = {
  shell: { display: 'flex', minHeight: '100vh', flexDirection: 'column' },
  nav: {
    background: '#1a1a2e', color: '#fff', display: 'flex',
    alignItems: 'center', padding: '0 24px', height: 56,
    gap: 4, flexWrap: 'wrap',
  },
  brand: { fontWeight: 700, fontSize: 17, letterSpacing: 1, color: '#e0e0e0', marginRight: 12 },
  link: {
    color: '#aaa', textDecoration: 'none', fontSize: 13, fontWeight: 500,
    padding: '5px 11px', borderRadius: 5, whiteSpace: 'nowrap',
  },
  activeLink: { color: '#fff', background: '#2d2d5e' },
  spacer: { flex: 1 },
  logout: {
    background: 'none', border: '1px solid #555', color: '#ccc',
    cursor: 'pointer', borderRadius: 5, padding: '5px 14px', fontSize: 13,
    marginLeft: 8,
  },
  main: { flex: 1, padding: 24, background: '#f0f2f5' },
}

const LINKS = [
  { to: '/dashboard',       label: 'Dashboard' },
  { to: '/scanner',         label: 'Scanner' },
  { to: '/inconsistencias', label: 'Inconsistencias' },
  { to: '/actas',           label: 'Actas' },
]

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
        {LINKS.map(({ to, label }) => (
          <NavLink
            key={to}
            to={to}
            style={({ isActive }) => ({ ...s.link, ...(isActive ? s.activeLink : {}) })}
          >
            {label}
          </NavLink>
        ))}
        <div style={s.spacer} />
        <button style={s.logout} onClick={logout}>Cerrar sesión</button>
      </nav>
      <main style={s.main}>
        <Outlet />
      </main>
    </div>
  )
}

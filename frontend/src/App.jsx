import React from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import Login from './pages/Login.jsx'
import Actas from './pages/Actas.jsx'
import ActaForm from './pages/ActaForm.jsx'
import DashboardPage from './pages/DashboardPage.jsx'
import InconsistenciasPage from './pages/InconsistenciasPage.jsx'
import ScannerPage from './pages/ScannerPage.jsx'
import Layout from './components/Layout.jsx'

function ProtectedRoute({ children }) {
  const token = localStorage.getItem('token')
  if (!token) return <Navigate to="/login" replace />
  return children
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<DashboardPage />} />
        <Route path="inconsistencias" element={<InconsistenciasPage />} />
        <Route path="scanner" element={<ScannerPage />} />
        <Route path="actas" element={<Actas />} />
        <Route path="actas/nueva" element={<ActaForm />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

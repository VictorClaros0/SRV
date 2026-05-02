import React, { useEffect, useState, useCallback } from 'react'
import { api } from '../api.js'

const s = {
  page: {},
  header: { marginBottom: 24 },
  title: { fontSize: 22, fontWeight: 800, color: '#1a1a2e', marginBottom: 4 },
  subtitle: { fontSize: 13, color: '#666' },

  card: {
    background: '#fff', borderRadius: 10, padding: '20px 24px',
    boxShadow: '0 2px 8px rgba(0,0,0,.06)', marginBottom: 16,
  },
  cardTitle: { fontSize: 14, fontWeight: 700, color: '#1a1a2e', marginBottom: 16 },

  table: { width: '100%', borderCollapse: 'collapse', fontSize: 13 },
  th: {
    background: '#f7f8fa', padding: '8px 12px', textAlign: 'left',
    fontSize: 11, color: '#666', fontWeight: 600,
  },
  td: { padding: '10px 12px', borderTop: '1px solid #f0f0f0', verticalAlign: 'middle' },

  badge: {
    display: 'inline-block', background: '#e6f9f0', color: '#1a8a4a',
    borderRadius: 20, padding: '2px 10px', fontSize: 11, fontWeight: 600,
  },

  removeBtn: {
    background: 'none', border: '1px solid #e0e0e0', color: '#c0392b',
    borderRadius: 5, padding: '4px 12px', cursor: 'pointer', fontSize: 12,
    fontWeight: 500,
  },

  addRow: { display: 'flex', gap: 8, marginTop: 16, alignItems: 'center' },
  input: {
    flex: 1, border: '1px solid #ddd', borderRadius: 6, padding: '8px 12px',
    fontSize: 13, outline: 'none',
  },
  addBtn: {
    background: '#1a1a2e', color: '#fff', border: 'none', borderRadius: 6,
    padding: '8px 18px', cursor: 'pointer', fontSize: 13, fontWeight: 600,
    whiteSpace: 'nowrap',
  },
  addBtnDisabled: { opacity: 0.5, cursor: 'not-allowed' },

  hint: { fontSize: 11, color: '#999', marginTop: 6 },
  errorMsg: { color: '#c0392b', fontSize: 12, marginTop: 6 },
  successMsg: { color: '#1a8a4a', fontSize: 12, marginTop: 6 },

  infoCard: {
    background: '#f7f8fa', borderRadius: 10, padding: '16px 20px',
    marginBottom: 16, borderLeft: '3px solid #1a1a2e',
  },
  infoTitle: { fontSize: 13, fontWeight: 700, color: '#1a1a2e', marginBottom: 8 },
  infoCode: {
    fontFamily: 'monospace', background: '#eee', padding: '6px 10px',
    borderRadius: 5, fontSize: 12, display: 'block', marginBottom: 4,
  },
  infoText: { fontSize: 12, color: '#555', marginBottom: 4 },

  emptyMsg: { fontSize: 13, color: '#aaa', padding: '16px 0', textAlign: 'center' },
}

function formatBoliviano(numero) {
  return `+591 ${numero}`
}

export default function SMSPage() {
  const [numeros, setNumeros] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [nuevoNumero, setNuevoNumero] = useState('')
  const [addError, setAddError] = useState('')
  const [addSuccess, setAddSuccess] = useState('')
  const [adding, setAdding] = useState(false)
  const [removiendo, setRemoviendo] = useState(null)
  
  const [historial, setHistorial] = useState([])
  const [loadingHistorial, setLoadingHistorial] = useState(true)

  const cargar = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const data = await api.sms.getNumeros()
      setNumeros(data.numeros || [])
    } catch (e) {
      setError(e.message || 'Error al cargar números')
    } finally {
      setLoading(false)
    }
  }, [])

  const cargarHistorial = useCallback(async () => {
    try {
      const data = await api.sms.getHistorial()
      setHistorial(data.historial || [])
    } catch (e) {
      console.error('Error al cargar historial SMS:', e)
    } finally {
      setLoadingHistorial(false)
    }
  }, [])

  useEffect(() => { 
    cargar() 
    cargarHistorial()
    
    const intervalId = setInterval(() => {
      cargarHistorial()
    }, 3000)
    
    return () => clearInterval(intervalId)
  }, [cargar, cargarHistorial])

  async function agregar() {
    const n = nuevoNumero.trim()
    if (!n) return
    setAdding(true)
    setAddError('')
    setAddSuccess('')
    try {
      await api.sms.addNumero(n)
      setNuevoNumero('')
      setAddSuccess(`Número autorizado agregado correctamente.`)
      await cargar()
    } catch (e) {
      setAddError(e.message || 'Error al agregar número')
    } finally {
      setAdding(false)
    }
  }

  async function quitar(numero) {
    setRemoviendo(numero)
    try {
      await api.sms.removeNumero(numero)
      await cargar()
    } catch (e) {
      setError(e.message || 'Error al eliminar número')
    } finally {
      setRemoviendo(null)
    }
  }

  function handleKeyDown(e) {
    if (e.key === 'Enter') agregar()
  }

  return (
    <div style={s.page}>
      <div style={s.header}>
        <div style={s.title}>Canal SMS — RRV</div>
        <div style={s.subtitle}>
          Administración de números bolivianos autorizados para enviar actas por SMS al flujo de Recuento Rápido de Votos.
        </div>
      </div>

      {/* ── Números autorizados ─────────────────────────────────────────── */}
      <div style={s.card}>
        <div style={s.cardTitle}>Números autorizados</div>

        {error && <div style={s.errorMsg}>{error}</div>}

        {loading ? (
          <div style={s.emptyMsg}>Cargando...</div>
        ) : numeros.length === 0 ? (
          <div style={s.emptyMsg}>No hay números autorizados configurados.</div>
        ) : (
          <table style={s.table}>
            <thead>
              <tr>
                <th style={s.th}>#</th>
                <th style={s.th}>Número (normalizado)</th>
                <th style={s.th}>Formato internacional</th>
                <th style={s.th}>Estado</th>
                <th style={s.th}>Acciones</th>
              </tr>
            </thead>
            <tbody>
              {numeros.map((numero, i) => (
                <tr key={numero}>
                  <td style={s.td}>{i + 1}</td>
                  <td style={{ ...s.td, fontFamily: 'monospace', fontWeight: 600 }}>{numero}</td>
                  <td style={{ ...s.td, fontFamily: 'monospace', color: '#555' }}>
                    {formatBoliviano(numero)}
                  </td>
                  <td style={s.td}>
                    <span style={s.badge}>Autorizado</span>
                  </td>
                  <td style={s.td}>
                    <button
                      style={s.removeBtn}
                      onClick={() => quitar(numero)}
                      disabled={removiendo === numero}
                    >
                      {removiendo === numero ? 'Quitando...' : 'Quitar'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {/* Agregar número */}
        <div style={s.addRow}>
          <input
            style={s.input}
            type="text"
            placeholder="Ej: 65707079 · +59165707079 · 59165707079"
            value={nuevoNumero}
            onChange={e => { setNuevoNumero(e.target.value); setAddError(''); setAddSuccess('') }}
            onKeyDown={handleKeyDown}
            disabled={adding}
          />
          <button
            style={{ ...s.addBtn, ...(adding ? s.addBtnDisabled : {}) }}
            onClick={agregar}
            disabled={adding || !nuevoNumero.trim()}
          >
            {adding ? 'Agregando...' : '+ Autorizar número'}
          </button>
        </div>
        <div style={s.hint}>
          Acepta cualquier formato: <code>65707079</code>, <code>+59165707079</code> o <code>59165707079</code>.
          El sistema normaliza automáticamente quitando el prefijo <code>+591</code>.
        </div>
        {addError && <div style={s.errorMsg}>{addError}</div>}
        {addSuccess && <div style={s.successMsg}>{addSuccess}</div>}
      </div>

      {/* ── Formato del SMS ─────────────────────────────────────────────── */}
      <div style={s.infoCard}>
        <div style={s.infoTitle}>Formato de SMS para enviar un acta</div>
        <code style={s.infoCode}>
          RRV|MESA=10101001001|P1=120|P2=95|P3=40|P4=30|BLANCOS=4|NULOS=3|TOTAL=292
        </code>
        <div style={s.infoText}>
          Campos obligatorios: <strong>MESA, P1, P2, P3, P4, BLANCOS, NULOS, TOTAL</strong>.
          El orden de los campos no importa. El TOTAL debe coincidir con P1+P2+P3+P4+BLANCOS+NULOS
          para que el acta quede como <strong>VALIDADA</strong>; si no coincide queda como <strong>PENDIENTE_REVISION</strong>.
        </div>
      </div>

      {/* ── Endpoint webhook ────────────────────────────────────────────── */}
      <div style={s.infoCard}>
        <div style={s.infoTitle}>Configuración en SMS Forwarder (Android)</div>
        <div style={s.infoText}><strong>URL:</strong></div>
        <code style={s.infoCode}>POST /api/v1/rrv/sms/inbound</code>
        <div style={s.infoText}><strong>Content-Type:</strong> application/json</div>
        <div style={s.infoText}><strong>Body personalizado:</strong></div>
        <code style={s.infoCode}>{`{"text":"Desde : {{sender}}\\n{{message}}"}`}</code>
        <div style={{ ...s.infoText, marginTop: 4 }}>
          Solo se procesan SMS de números que aparecen en la lista de autorizados de arriba.
          Los demás son rechazados sin crear ningún registro.
        </div>
      </div>

      {/* ── Historial de Mensajes Recibidos ─────────────────────────────── */}
      <div style={s.card}>
        <div style={s.cardTitle}>
          Mensajes Recibidos (Auto-refresh)
          <span style={{ float: 'right', fontSize: 12, fontWeight: 400, color: '#888' }}>
            Total: {historial.length}
          </span>
        </div>

        {loadingHistorial ? (
          <div style={s.emptyMsg}>Cargando historial...</div>
        ) : historial.length === 0 ? (
          <div style={s.emptyMsg}>No se han recibido mensajes aún.</div>
        ) : (
          <table style={s.table}>
            <thead>
              <tr>
                <th style={s.th}>Fecha</th>
                <th style={s.th}>Remitente</th>
                <th style={s.th}>Estado</th>
                <th style={s.th}>Mensaje</th>
                <th style={s.th}>Raw JSON</th>
              </tr>
            </thead>
            <tbody>
              {historial.map((msg, i) => (
                <tr key={msg.acta_id || i}>
                  <td style={{ ...s.td, fontSize: 11, whiteSpace: 'nowrap' }}>
                    {new Date(msg.fecha_recepcion).toLocaleString()}
                  </td>
                  <td style={{ ...s.td, fontWeight: 600 }}>{msg.remitente || 'Desconocido'}</td>
                  <td style={s.td}>
                    <span style={{ 
                      ...s.badge, 
                      background: msg.estado === 'VALIDADA' ? '#e6f9f0' : (msg.estado === 'RECHAZADA' ? '#feecec' : '#fef4e6'),
                      color: msg.estado === 'VALIDADA' ? '#1a8a4a' : (msg.estado === 'RECHAZADA' ? '#c0392b' : '#d35400')
                    }}>
                      {msg.estado}
                    </span>
                  </td>
                  <td style={{ ...s.td, maxWidth: 300, wordWrap: 'break-word', fontSize: 11 }}>
                    {msg.sms_body}
                  </td>
                  <td style={{ ...s.td, maxWidth: 200, fontSize: 10 }}>
                    <pre style={{ margin: 0, whiteSpace: 'pre-wrap', fontFamily: 'monospace', color: '#666', background: '#f5f5f5', padding: 4, borderRadius: 4 }}>
                      {JSON.stringify(msg.raw_payload, null, 2)}
                    </pre>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

    </div>
  )
}

const BASE = '/api/v1'

function getToken() {
  return localStorage.getItem('token')
}

function authHeaders() {
  return { 'Content-Type': 'application/json', Authorization: `Bearer ${getToken()}` }
}

async function request(method, path, body) {
  const res = await fetch(BASE + path, {
    method,
    headers: authHeaders(),
    body: body ? JSON.stringify(body) : undefined,
  })
  if (res.status === 401) {
    localStorage.removeItem('token')
    window.location.href = '/login'
    return
  }
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || res.statusText)
  }
  if (res.status === 204) return null
  return res.json()
}

export const api = {
  login: (usuario, contrasena) =>
    fetch(BASE + '/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ usuario, contrasena }),
    }).then(async (r) => {
      const data = await r.json()
      if (!r.ok) throw new Error(data.error || 'Error de autenticación')
      return data
    }),

  me: () => request('GET', '/auth/me'),

  actas: {
    list: (params = {}) => {
      const qs = new URLSearchParams(params).toString()
      return request('GET', `/actas${qs ? '?' + qs : ''}`)
    },
    get: (id) => request('GET', `/actas/${id}`),
    create: (body) => request('POST', '/actas', body),
    update: (id, body) => request('PUT', `/actas/${id}`, body),
    delete: (id) => request('DELETE', `/actas/${id}`),
    resumen: () => request('GET', '/actas/resumen'),
  },

  recintos: {
    list: () => request('GET', '/recintos'),
  },

  resultados: {
    get: () => request('GET', '/resultados'),
    auditoria: () => request('GET', '/auditoria'),
  },

  // Dispara el workflow de n8n que simula la transcripción masiva.
  // n8n lee /api/v1/actas/para-transcribir y llama al webhook por cada acta.
  simularTranscripcion: () =>
    fetch('/n8n/webhook/trigger-transcripcion', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{}',
    }).then(async r => {
      if (!r.ok) {
        const txt = await r.text().catch(() => r.statusText)
        throw new Error(txt || 'n8n no respondió. ¿El workflow está activo?')
      }
      return r.json().catch(() => ({ ok: true }))
    }),
}

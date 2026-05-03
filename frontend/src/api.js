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

function qs(params) {
  const p = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => { if (v) p.set(k, v) })
  const s = p.toString()
  return s ? '?' + s : ''
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
    list: (params = {}) => request('GET', `/actas${qs(params)}`),
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
    get: (params = {}) => request('GET', `/resultados${qs(params)}`),
    auditoria: () => request('GET', '/auditoria'),
  },

  filtros: {
    departamentos: () =>
      request('GET', '/filtros/departamentos'),
    municipios: (departamento = '') =>
      request('GET', `/filtros/municipios${qs({ departamento })}`),
    provincias: (departamento = '', municipio = '') =>
      request('GET', `/filtros/provincias${qs({ departamento, municipio })}`),
    recintos: (departamento = '', municipio = '', provincia = '') =>
      request('GET', `/filtros/recintos${qs({ departamento, municipio, provincia })}`),
    mesas: (recinto = '') =>
      request('GET', `/filtros/mesas${qs({ recinto })}`),
  },

  procesarTranscripciones: () =>
    fetch(BASE + '/actas/procesar-transcripciones', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    }).then(async r => {
      if (!r.ok) throw new Error(await r.text().catch(() => r.statusText))
      return r.json()
    }),
}

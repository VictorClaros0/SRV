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

  // ─── Transcripción masiva (flujo: React → Backend → n8n → Backend webhook → PostgreSQL) ──
  // El frontend NUNCA llama a n8n directamente.

  transcripcion: {
    // Llama al backend que orquesta n8n.
    // opts.fallback = true → si n8n falla, backend procesa Transcripciones.csv directamente.
    simular: (opts = {}) => {
      const qs = opts.fallback ? '?fallback=true' : ''
      return request('POST', `/transcripcion/simular${qs}`, {})
    },
  },

  // ─── Dashboard ───────────────────────────────────────────────────────────────
  // Todos los endpoints son GET de solo lectura. No calculan nada en frontend.

  dashboard: {
    getKPIs: () => request('GET', '/dashboard/kpis'),
    getRRVvsOficial: () => request('GET', '/dashboard/rrv-vs-oficial'),
    getVotosCandidato: () => request('GET', '/dashboard/votos-candidato'),
    getParticipacion: () => request('GET', '/dashboard/participacion'),
    getGeografico: (groupBy = 'departamento') =>
      request('GET', `/dashboard/geografico?group_by=${groupBy}`),
    getTecnico: () => request('GET', '/dashboard/tecnico'),
    getInconsistencias: () => request('GET', '/dashboard/inconsistencias'),
  },

  // ─── Comparación RRV vs Oficial ──────────────────────────────────────────────

  comparacion: {
    list: (params = {}) => {
      const qs = new URLSearchParams(params).toString()
      return request('GET', `/comparacion${qs ? '?' + qs : ''}`)
    },
    resumen: () => request('GET', '/comparacion/resumen'),
    getActa: (actaId) => request('GET', `/comparacion/${actaId}`),
  },

  // ─── Scanner / Transcriptor ──────────────────────────────────────────────────
  // Envía archivo (imagen/PDF) al backend. NO usa JSON, usa FormData.

  scanner: {
    transcribir: (file, opts = {}) => {
      const fd = new FormData()
      fd.append('archivo', file)
      if (opts.codigoActa) fd.append('codigo_acta', String(opts.codigoActa))
      if (opts.actaId) fd.append('acta_id', String(opts.actaId))

      return fetch(BASE + '/scanner/transcribir', {
        method: 'POST',
        // No poner Content-Type: el browser lo setea con boundary para FormData
        headers: { Authorization: `Bearer ${getToken()}` },
        body: fd,
      }).then(async r => {
        if (r.status === 401) {
          localStorage.removeItem('token')
          window.location.href = '/login'
          return
        }
        const data = await r.json().catch(() => ({ error: r.statusText }))
        if (!r.ok) throw new Error(data.error || r.statusText)
        return data
      })
    },
  },
}

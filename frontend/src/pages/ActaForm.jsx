import React, { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { api } from '../api.js'

const s = {
  header: { display: 'flex', alignItems: 'center', gap: 12, marginBottom: 28 },
  title: { fontSize: 20, fontWeight: 700, color: '#1a1a2e' },
  back: { background: 'none', border: '1px solid #ddd', borderRadius: 6, padding: '6px 14px', cursor: 'pointer', fontSize: 14 },
  card: { background: '#fff', borderRadius: 12, padding: 28, boxShadow: '0 2px 10px rgba(0,0,0,.07)', maxWidth: 860 },
  section: { marginBottom: 28 },
  sectionTitle: { fontSize: 13, fontWeight: 700, color: '#888', textTransform: 'uppercase', letterSpacing: .5, marginBottom: 16, paddingBottom: 8, borderBottom: '1px solid #f0f0f0' },
  grid: { display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(180px, 1fr))', gap: '12px 20px' },
  grid2: { display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px 20px' },
  group: { display: 'flex', flexDirection: 'column', gap: 4 },
  label: { fontSize: 12, fontWeight: 600, color: '#555' },
  input: { padding: '8px 12px', border: '1.5px solid #ddd', borderRadius: 7, fontSize: 14, outline: 'none' },
  select: { padding: '8px 12px', border: '1.5px solid #ddd', borderRadius: 7, fontSize: 14, outline: 'none', background: '#fff' },
  textarea: { padding: '8px 12px', border: '1.5px solid #ddd', borderRadius: 7, fontSize: 14, outline: 'none', resize: 'vertical', minHeight: 72 },
  footer: { display: 'flex', gap: 12, marginTop: 24 },
  btnPrimary: { background: '#1a1a2e', color: '#fff', border: 'none', borderRadius: 8, padding: '10px 24px', cursor: 'pointer', fontSize: 14, fontWeight: 600 },
  btnSecondary: { background: '#f0f0f0', color: '#444', border: 'none', borderRadius: 8, padding: '10px 20px', cursor: 'pointer', fontSize: 14 },
  error: { background: '#fff0f0', color: '#c0392b', border: '1px solid #fac0c0', borderRadius: 6, padding: '10px 14px', fontSize: 13, marginBottom: 16 },
  success: { background: '#f0fff4', color: '#27ae60', border: '1px solid #b2dfdb', borderRadius: 6, padding: '10px 14px', fontSize: 13, marginBottom: 16 },
}

const INIT = {
  codigoActa: '', codigoRecinto: '', nroMesa: 1, estado: 'impresa',
  papeletasAnfora: 0, papeletasNoUsadas: 0,
  p1: 0, p2: 0, p3: 0, p4: 0,
  votosValidos: 0, votosNulos: 0, votosBlanco: 0,
  observaciones: '',
  aperturaHora: 8, aperturaMinutos: 0, cierreHora: 18, cierreMinutos: 0,
}

export default function ActaForm() {
  const { id } = useParams()
  const nav = useNavigate()
  const editing = Boolean(id)
  const [form, setForm] = useState(INIT)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [loading, setLoading] = useState(false)
  const [fetchLoading, setFetchLoading] = useState(editing)

  useEffect(() => {
    if (!editing) return
    api.actas.get(id).then(data => {
      setForm({
        codigoActa: data.codigoActa ?? '',
        codigoRecinto: data.codigoRecinto ?? '',
        nroMesa: data.nroMesa ?? 1,
        estado: data.estado ?? 'impresa',
        papeletasAnfora: data.papeletasAnfora ?? 0,
        papeletasNoUsadas: data.papeletasNoUsadas ?? 0,
        p1: data.p1 ?? 0, p2: data.p2 ?? 0, p3: data.p3 ?? 0, p4: data.p4 ?? 0,
        votosValidos: data.votosValidos ?? 0,
        votosNulos: data.votosNulos ?? 0,
        votosBlanco: data.votosBlanco ?? 0,
        observaciones: data.observaciones ?? '',
        aperturaHora: data.aperturaHora ?? 8,
        aperturaMinutos: data.aperturaMinutos ?? 0,
        cierreHora: data.cierreHora ?? 18,
        cierreMinutos: data.cierreMinutos ?? 0,
      })
      setFetchLoading(false)
    }).catch(e => { setError(e.message); setFetchLoading(false) })
  }, [id, editing])

  function set(field) {
    return e => setForm(f => ({ ...f, [field]: e.target.value }))
  }

  function setNum(field) {
    return e => setForm(f => ({ ...f, [field]: Number(e.target.value) || 0 }))
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    setSuccess('')
    setLoading(true)
    const body = {
      ...form,
      codigoActa: Number(form.codigoActa) || 0,
      codigoRecinto: Number(form.codigoRecinto) || 0,
      nroMesa: Number(form.nroMesa) || 0,
    }
    try {
      if (editing) {
        await api.actas.update(id, body)
        setSuccess('Acta actualizada correctamente')
      } else {
        await api.actas.create(body)
        setSuccess('Acta creada correctamente')
        setTimeout(() => nav('/actas'), 1200)
      }
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  if (fetchLoading) return <div>Cargando...</div>

  return (
    <div>
      <div style={s.header}>
        <button style={s.back} onClick={() => nav('/actas')}>← Volver</button>
        <div style={s.title}>{editing ? 'Editar acta' : 'Nueva acta'}</div>
      </div>

      <div style={s.card}>
        {error && <div style={s.error}>{error}</div>}
        {success && <div style={s.success}>{success}</div>}

        <form onSubmit={handleSubmit}>

          {/* Identificación */}
          <div style={s.section}>
            <div style={s.sectionTitle}>Identificación</div>
            <div style={s.grid}>
              <div style={s.group}>
                <label style={s.label}>Código Acta</label>
                <input style={s.input} value={form.codigoActa} onChange={set('codigoActa')} required />
              </div>
              <div style={s.group}>
                <label style={s.label}>Código Recinto</label>
                <input style={s.input} value={form.codigoRecinto} onChange={set('codigoRecinto')} required />
              </div>
              <div style={s.group}>
                <label style={s.label}>Nro Mesa</label>
                <input style={s.input} type="number" min={1} value={form.nroMesa} onChange={setNum('nroMesa')} required />
              </div>
              <div style={s.group}>
                <label style={s.label}>Estado</label>
                <select style={s.select} value={form.estado} onChange={set('estado')}>
                  <option value="impresa">Impresa</option>
                  <option value="transcrita">Transcrita</option>
                </select>
              </div>
            </div>
          </div>

          {/* Papeletas */}
          <div style={s.section}>
            <div style={s.sectionTitle}>Papeletas</div>
            <div style={s.grid}>
              <div style={s.group}>
                <label style={s.label}>Papeletas en Ánfora</label>
                <input style={s.input} type="number" min={0} value={form.papeletasAnfora} onChange={setNum('papeletasAnfora')} />
              </div>
              <div style={s.group}>
                <label style={s.label}>Papeletas No Usadas</label>
                <input style={s.input} type="number" min={0} value={form.papeletasNoUsadas} onChange={setNum('papeletasNoUsadas')} />
              </div>
            </div>
          </div>

          {/* Votos */}
          <div style={s.section}>
            <div style={s.sectionTitle}>Votos por partido</div>
            <div style={s.grid}>
              {['p1','p2','p3','p4'].map(p => (
                <div key={p} style={s.group}>
                  <label style={s.label}>{p.toUpperCase()}</label>
                  <input style={s.input} type="number" min={0} value={form[p]} onChange={setNum(p)} />
                </div>
              ))}
            </div>
          </div>

          {/* Totales */}
          <div style={s.section}>
            <div style={s.sectionTitle}>Totales</div>
            <div style={s.grid}>
              <div style={s.group}>
                <label style={s.label}>Votos Válidos</label>
                <input style={s.input} type="number" min={0} value={form.votosValidos} onChange={setNum('votosValidos')} />
              </div>
              <div style={s.group}>
                <label style={s.label}>Votos Nulos</label>
                <input style={s.input} type="number" min={0} value={form.votosNulos} onChange={setNum('votosNulos')} />
              </div>
              <div style={s.group}>
                <label style={s.label}>Votos en Blanco</label>
                <input style={s.input} type="number" min={0} value={form.votosBlanco} onChange={setNum('votosBlanco')} />
              </div>
            </div>
          </div>

          {/* Horarios */}
          <div style={s.section}>
            <div style={s.sectionTitle}>Horario</div>
            <div style={s.grid}>
              <div style={s.group}>
                <label style={s.label}>Apertura – Hora</label>
                <input style={s.input} type="number" min={0} max={23} value={form.aperturaHora} onChange={setNum('aperturaHora')} />
              </div>
              <div style={s.group}>
                <label style={s.label}>Apertura – Minutos</label>
                <input style={s.input} type="number" min={0} max={59} value={form.aperturaMinutos} onChange={setNum('aperturaMinutos')} />
              </div>
              <div style={s.group}>
                <label style={s.label}>Cierre – Hora</label>
                <input style={s.input} type="number" min={0} max={23} value={form.cierreHora} onChange={setNum('cierreHora')} />
              </div>
              <div style={s.group}>
                <label style={s.label}>Cierre – Minutos</label>
                <input style={s.input} type="number" min={0} max={59} value={form.cierreMinutos} onChange={setNum('cierreMinutos')} />
              </div>
            </div>
          </div>

          {/* Observaciones */}
          <div style={s.section}>
            <div style={s.sectionTitle}>Observaciones</div>
            <textarea style={{ ...s.textarea, width: '100%' }} value={form.observaciones} onChange={set('observaciones')} placeholder="Observaciones opcionales..." />
          </div>

          <div style={s.footer}>
            <button style={s.btnPrimary} type="submit" disabled={loading}>
              {loading ? 'Guardando...' : (editing ? 'Guardar cambios' : 'Crear acta')}
            </button>
            <button style={s.btnSecondary} type="button" onClick={() => nav('/actas')}>
              Cancelar
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

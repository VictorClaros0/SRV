import React, { useRef, useState } from 'react'

const ALLOWED_EXT = ['jpg', 'jpeg', 'png', 'pdf']
const MAX_SIZE_MB = 10

const s = {
  zone: (active) => ({
    border: `2px dashed ${active ? '#1a1a2e' : '#ddd'}`,
    borderRadius: 10, padding: '32px 20px', textAlign: 'center',
    cursor: 'pointer', transition: 'border-color .2s',
    background: active ? '#f0f2f5' : '#fafafa',
  }),
  icon: { fontSize: 32, marginBottom: 8, display: 'block' },
  hint: { fontSize: 13, color: '#888', marginBottom: 4 },
  ext: { fontSize: 11, color: '#bbb' },
  input: { display: 'none' },
  preview: {
    marginTop: 14, display: 'flex', alignItems: 'center', gap: 12,
    background: '#f0fff4', border: '1px solid #b2dfdb', borderRadius: 8, padding: '10px 14px',
  },
  previewImg: { width: 60, height: 60, objectFit: 'cover', borderRadius: 6, border: '1px solid #ddd' },
  previewInfo: { flex: 1 },
  previewName: { fontSize: 13, fontWeight: 600, color: '#1a1a2e' },
  previewSize: { fontSize: 12, color: '#888' },
  removeBtn: {
    background: 'none', border: '1px solid #ef4444', color: '#ef4444',
    borderRadius: 5, padding: '3px 10px', cursor: 'pointer', fontSize: 12,
  },
  error: {
    marginTop: 8, background: '#fff0f0', border: '1px solid #fac0c0',
    borderRadius: 6, padding: '8px 12px', color: '#c0392b', fontSize: 13,
  },
}

export default function FileUploader({ file, onFile, onError }) {
  const inputRef = useRef(null)
  const [drag, setDrag] = useState(false)
  const [localError, setLocalError] = useState('')

  function validate(f) {
    setLocalError('')
    if (!f) return false
    const ext = f.name.split('.').pop().toLowerCase()
    if (!ALLOWED_EXT.includes(ext)) {
      const msg = `Extensión no permitida: .${ext}. Use: ${ALLOWED_EXT.join(', ')}`
      setLocalError(msg)
      onError?.(msg)
      return false
    }
    if (f.size > MAX_SIZE_MB * 1024 * 1024) {
      const msg = `El archivo supera ${MAX_SIZE_MB} MB.`
      setLocalError(msg)
      onError?.(msg)
      return false
    }
    return true
  }

  function handleFiles(files) {
    const f = files[0]
    if (!f) return
    if (validate(f)) {
      onFile(f)
      onError?.('')
    }
  }

  function onDrop(e) {
    e.preventDefault()
    setDrag(false)
    handleFiles(e.dataTransfer.files)
  }

  const imgUrl = file && file.type?.startsWith('image/') ? URL.createObjectURL(file) : null

  return (
    <div>
      {!file ? (
        <div
          style={s.zone(drag)}
          onClick={() => inputRef.current?.click()}
          onDragOver={e => { e.preventDefault(); setDrag(true) }}
          onDragLeave={() => setDrag(false)}
          onDrop={onDrop}
        >
          <span style={s.icon}>📎</span>
          <div style={s.hint}>Haz clic o arrastra un archivo aquí</div>
          <div style={s.ext}>jpg · jpeg · png · pdf — máx. {MAX_SIZE_MB} MB</div>
          <input
            ref={inputRef}
            type="file"
            accept=".jpg,.jpeg,.png,.pdf"
            style={s.input}
            onChange={e => handleFiles(e.target.files)}
          />
        </div>
      ) : (
        <div style={s.preview}>
          {imgUrl ? (
            <img src={imgUrl} alt="preview" style={s.previewImg} />
          ) : (
            <span style={{ fontSize: 32 }}>📄</span>
          )}
          <div style={s.previewInfo}>
            <div style={s.previewName}>{file.name}</div>
            <div style={s.previewSize}>{(file.size / 1024).toFixed(1)} KB</div>
          </div>
          <button style={s.removeBtn} onClick={() => { onFile(null); setLocalError('') }}>
            Quitar
          </button>
        </div>
      )}

      {localError && <div style={s.error}>{localError}</div>}
    </div>
  )
}

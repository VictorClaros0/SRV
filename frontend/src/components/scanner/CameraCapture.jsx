import React, { useRef, useState, useEffect } from 'react'

const s = {
  wrap: {},
  video: {
    width: '100%', borderRadius: 8, maxHeight: 320, background: '#000',
    display: 'block',
  },
  canvas: { display: 'none' },
  btnRow: { display: 'flex', gap: 10, marginTop: 10, flexWrap: 'wrap' },
  btn: (variant) => ({
    padding: '9px 18px', borderRadius: 7, border: 'none', cursor: 'pointer',
    fontSize: 13, fontWeight: 600,
    ...(variant === 'primary' ? { background: '#1a1a2e', color: '#fff' } :
        variant === 'danger'  ? { background: '#fee2e2', color: '#c0392b' } :
                                { background: '#f0f0f0', color: '#444' }),
  }),
  preview: {
    marginTop: 10, position: 'relative', display: 'inline-block',
  },
  previewImg: { width: '100%', borderRadius: 8, border: '2px solid #22c55e' },
  previewLabel: {
    position: 'absolute', bottom: 8, left: 8,
    background: 'rgba(0,0,0,.55)', color: '#fff',
    fontSize: 12, borderRadius: 4, padding: '2px 8px',
  },
  error: {
    background: '#fff0f0', border: '1px solid #fac0c0', borderRadius: 6,
    padding: '10px 14px', color: '#c0392b', fontSize: 13, marginTop: 8,
  },
  hint: { fontSize: 12, color: '#888', marginBottom: 8 },
}

export default function CameraCapture({ onCapture }) {
  const videoRef = useRef(null)
  const canvasRef = useRef(null)
  const streamRef = useRef(null)
  const [active, setActive] = useState(false)
  const [captured, setCaptured] = useState(null) // URL
  const [capturedFile, setCapturedFile] = useState(null) // File
  const [error, setError] = useState('')

  useEffect(() => {
    return () => { stopStream() }
  }, [])

  function stopStream() {
    if (streamRef.current) {
      streamRef.current.getTracks().forEach(t => t.stop())
      streamRef.current = null
    }
  }

  async function startCamera() {
    setError('')
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: { ideal: 'environment' } },
        audio: false,
      })
      streamRef.current = stream
      if (videoRef.current) {
        videoRef.current.srcObject = stream
        videoRef.current.play()
      }
      setActive(true)
      setCaptured(null)
      setCapturedFile(null)
    } catch (e) {
      setError('No se pudo acceder a la cámara. ' + (e.message || ''))
    }
  }

  function capturePhoto() {
    const video = videoRef.current
    const canvas = canvasRef.current
    if (!video || !canvas) return
    canvas.width = video.videoWidth
    canvas.height = video.videoHeight
    canvas.getContext('2d').drawImage(video, 0, 0)
    const url = canvas.toDataURL('image/jpeg', 0.9)
    setCaptured(url)

    canvas.toBlob(blob => {
      const file = new File([blob], `captura_${Date.now()}.jpg`, { type: 'image/jpeg' })
      setCapturedFile(file)
      onCapture?.(file)
    }, 'image/jpeg', 0.9)

    stopStream()
    setActive(false)
  }

  function retake() {
    setCaptured(null)
    setCapturedFile(null)
    onCapture?.(null)
    startCamera()
  }

  function discard() {
    setCaptured(null)
    setCapturedFile(null)
    onCapture?.(null)
    stopStream()
    setActive(false)
  }

  return (
    <div style={s.wrap}>
      <div style={s.hint}>
        La cámara funcionará desde localhost. En dispositivos móviles requiere HTTPS.
      </div>

      {!active && !captured && (
        <button style={s.btn('primary')} onClick={startCamera}>
          📷 Activar cámara
        </button>
      )}

      {active && (
        <>
          <video ref={videoRef} style={s.video} playsInline muted />
          <div style={s.btnRow}>
            <button style={s.btn('primary')} onClick={capturePhoto}>📸 Capturar acta</button>
            <button style={s.btn('danger')} onClick={discard}>Cancelar</button>
          </div>
        </>
      )}

      {captured && (
        <div style={s.preview}>
          <img src={captured} alt="Foto capturada" style={s.previewImg} />
          <span style={s.previewLabel}>Capturada</span>
          <div style={s.btnRow}>
            <button style={s.btn('default')} onClick={retake}>↺ Repetir</button>
            <button style={s.btn('danger')} onClick={discard}>Descartar</button>
          </div>
        </div>
      )}

      <canvas ref={canvasRef} style={s.canvas} />

      {error && <div style={s.error}>{error}</div>}
    </div>
  )
}

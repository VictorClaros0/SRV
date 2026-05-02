/* ══════════════════════════════════════════════════════════════
   Escáner Móvil de Actas - RRV Bolivia
   App Logic: Login → Capture → Upload → Results → Firebase RTDB
   ══════════════════════════════════════════════════════════════ */

// ─── CONFIGURACIÓN ──────────────────────────────────────────
const CONFIG = {
  // API URL - se detecta automáticamente o se pide al usuario
  API_URL: localStorage.getItem('api_url') || (location.hostname === 'localhost' ? 'http://localhost:8080' : ''),

  // Firebase Realtime Database
  FIREBASE_RTDB: 'https://lawdocs-75aef-default-rtdb.firebaseio.com',

  // Cloudinary
  CLOUDINARY_CLOUD_NAME: 'dk5u8dljb',
  CLOUDINARY_API_KEY: '779231566857641',
  CLOUDINARY_UPLOAD_PRESET: 'rrv_unsigned',
};

// Si no hay API_URL configurada (ej: Firebase hosting), pedir al usuario
if (!CONFIG.API_URL) {
  const saved = prompt('Ingresa la URL de tu backend (ej: https://xxxx.ngrok-free.app):');
  if (saved) {
    CONFIG.API_URL = saved.replace(/\/$/, '');
    localStorage.setItem('api_url', CONFIG.API_URL);
  }
}

// ─── STATE ──────────────────────────────────────────────────
let state = {
  jwt: null,
  user: null,
  capturedFile: null,
  processing: false,
};

// ─── DOM REFERENCES ─────────────────────────────────────────
const $ = id => document.getElementById(id);
const views = {
  login: $('view-login'),
  scanner: $('view-scanner'),
  history: $('view-history'),
};

// ─── VIEW MANAGEMENT ────────────────────────────────────────
function showView(name) {
  Object.values(views).forEach(v => v.classList.remove('active'));
  views[name].classList.add('active');
  if (name === 'history') loadHistory();
}

// ─── LOGIN ──────────────────────────────────────────────────
$('login-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const btn = $('btn-login');
  const errorEl = $('login-error');
  const user = $('login-user').value.trim();
  const pass = $('login-pass').value;

  btn.disabled = true;
  btn.querySelector('.btn-text').textContent = 'Ingresando...';
  errorEl.classList.add('hidden');

  try {
    const res = await fetch(`${CONFIG.API_URL}/api/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ usuario: user, contrasena: pass }),
    });

    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || 'Credenciales incorrectas');
    }

    const data = await res.json();
    state.jwt = data.token;
    state.user = data.usuario || user;

    // Persist session
    sessionStorage.setItem('jwt', state.jwt);
    sessionStorage.setItem('user', JSON.stringify(state.user));

    showView('scanner');
  } catch (err) {
    errorEl.textContent = err.message;
    errorEl.classList.remove('hidden');
  } finally {
    btn.disabled = false;
    btn.querySelector('.btn-text').textContent = 'Ingresar';
  }
});

// ─── LOGOUT ─────────────────────────────────────────────────
$('btn-logout').addEventListener('click', () => {
  state.jwt = null;
  state.user = null;
  state.capturedFile = null;
  sessionStorage.clear();
  showView('login');
});

// ─── NAVIGATION ─────────────────────────────────────────────
$('btn-history').addEventListener('click', () => showView('history'));
$('btn-back').addEventListener('click', () => showView('scanner'));

// ─── IMAGE CAPTURE ──────────────────────────────────────────
const fileInput = $('file-input');
const imgPreview = $('img-preview');
const placeholder = $('placeholder');
const previewArea = $('preview-area');
const btnUpload = $('btn-upload');

$('btn-capture').addEventListener('click', () => {
  fileInput.setAttribute('capture', 'environment');
  fileInput.click();
});

$('btn-gallery').addEventListener('click', () => {
  fileInput.removeAttribute('capture');
  fileInput.click();
});

fileInput.addEventListener('change', (e) => {
  const file = e.target.files[0];
  if (!file) return;

  state.capturedFile = file;

  const reader = new FileReader();
  reader.onload = (ev) => {
    imgPreview.src = ev.target.result;
    imgPreview.classList.remove('hidden');
    placeholder.classList.add('hidden');
    previewArea.classList.add('has-image');
    btnUpload.classList.remove('hidden');
    btnUpload.disabled = false;
  };
  reader.readAsDataURL(file);
});

// Allow tapping preview area to capture
previewArea.addEventListener('click', () => {
  if (!state.capturedFile) {
    fileInput.setAttribute('capture', 'environment');
    fileInput.click();
  }
});

// ─── UPLOAD & PROCESS ───────────────────────────────────────
btnUpload.addEventListener('click', async () => {
  if (!state.capturedFile || state.processing) return;

  state.processing = true;
  const btnText = btnUpload.querySelector('.btn-text');
  const btnLoader = btnUpload.querySelector('.btn-loader');
  btnText.classList.add('hidden');
  btnLoader.classList.remove('hidden');
  btnUpload.classList.add('processing');
  btnUpload.disabled = true;

  const resultSection = $('result-section');
  resultSection.classList.add('hidden');

  try {
    // 1. Upload to Cloudinary (client-side unsigned) if configured
    let cloudinaryUrl = null;
    if (CONFIG.CLOUDINARY_CLOUD_NAME && CONFIG.CLOUDINARY_UPLOAD_PRESET) {
      try {
        cloudinaryUrl = await uploadToCloudinary(state.capturedFile);
      } catch (cloudErr) {
        console.warn('Cloudinary upload failed, continuing without URL:', cloudErr);
      }
    }

    // 2. Send to backend for processing
    const formData = new FormData();
    formData.append('archivo', state.capturedFile);
    const codigoActa = $('codigo-acta').value.trim();
    if (codigoActa) {
      formData.append('codigo_acta', codigoActa);
    }
    if (cloudinaryUrl) {
      formData.append('imagen_url', cloudinaryUrl);
    }

    const res = await fetch(`${CONFIG.API_URL}/api/v1/scanner/transcribir`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${state.jwt}` },
      body: formData,
    });

    if (res.status === 401) {
      alert('Sesión expirada. Inicia sesión de nuevo.');
      showView('login');
      return;
    }

    const result = await res.json();

    if (!result.success) {
      throw new Error(result.error || 'Error al procesar el acta');
    }

    // 3. Display result
    displayResult(result, cloudinaryUrl);

    // 4. Save to Firebase RTDB
    await saveToFirebase(result, cloudinaryUrl);

  } catch (err) {
    displayError(err.message);
  } finally {
    state.processing = false;
    btnText.classList.remove('hidden');
    btnLoader.classList.add('hidden');
    btnUpload.classList.remove('processing');
    btnUpload.disabled = false;
  }
});

// ─── CLOUDINARY UPLOAD (Unsigned) ───────────────────────────
async function uploadToCloudinary(file) {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('upload_preset', CONFIG.CLOUDINARY_UPLOAD_PRESET);
  formData.append('folder', 'rrv_actas');

  const res = await fetch(
    `https://api.cloudinary.com/v1_1/${CONFIG.CLOUDINARY_CLOUD_NAME}/image/upload`,
    { method: 'POST', body: formData }
  );

  if (!res.ok) throw new Error('Cloudinary upload failed');
  const data = await res.json();
  return data.secure_url;
}

// ─── DISPLAY RESULT ─────────────────────────────────────────
function displayResult(result, imageUrl) {
  const section = $('result-section');
  const header = $('result-header');
  const icon = $('result-icon');
  const status = $('result-status');
  const details = $('result-details');

  const data = result.data || {};
  const estado = result.estado || 'desconocido';

  let headerClass, iconText, statusText;
  if (estado === 'transcrita') {
    headerClass = 'validated';
    iconText = '✅';
    statusText = 'Acta Validada';
  } else if (estado === 'observada') {
    headerClass = 'observed';
    iconText = '⚠️';
    statusText = 'Acta Observada';
  } else {
    headerClass = 'error';
    iconText = '❌';
    statusText = 'Acta con Error';
  }

  header.className = 'result-header ' + headerClass;
  icon.textContent = iconText;
  status.textContent = statusText;
  status.className = 'status-text';

  const totalVotos = (data.p1 || 0) + (data.p2 || 0) + (data.p3 || 0) + (data.p4 || 0);

  details.innerHTML = `
    <div class="detail-item">
      <div class="label">Código Acta</div>
      <div class="value">${data.codigo_acta || '-'}</div>
    </div>
    <div class="detail-item">
      <div class="label">Estado</div>
      <div class="value" style="color: var(--${headerClass === 'validated' ? 'success' : headerClass === 'observed' ? 'warning' : 'danger'})">${estado.toUpperCase()}</div>
    </div>
    <div class="detail-item">
      <div class="label">Partido 1</div>
      <div class="value">${data.p1 ?? '-'}</div>
    </div>
    <div class="detail-item">
      <div class="label">Partido 2</div>
      <div class="value">${data.p2 ?? '-'}</div>
    </div>
    <div class="detail-item">
      <div class="label">Partido 3</div>
      <div class="value">${data.p3 ?? '-'}</div>
    </div>
    <div class="detail-item">
      <div class="label">Partido 4</div>
      <div class="value">${data.p4 ?? '-'}</div>
    </div>
    <div class="detail-item">
      <div class="label">Votos Válidos</div>
      <div class="value">${data.votos_validos ?? totalVotos}</div>
    </div>
    <div class="detail-item">
      <div class="label">Votos Nulos</div>
      <div class="value">${data.votos_nulos ?? '-'}</div>
    </div>
    <div class="detail-item">
      <div class="label">Votos Blancos</div>
      <div class="value">${data.votos_blancos ?? '-'}</div>
    </div>
    <div class="detail-item">
      <div class="label">Papeletas Ánfora</div>
      <div class="value">${data.papeletas_anfora ?? '-'}</div>
    </div>
    ${data.observaciones ? `
    <div class="detail-item full-width">
      <div class="label">Observaciones</div>
      <div class="value" style="font-size:14px">${data.observaciones}</div>
    </div>` : ''}
    ${result.acta_actualizada ? `
    <div class="detail-item full-width">
      <div class="label">Base de Datos</div>
      <div class="value" style="font-size:14px; color: var(--success)">✅ Acta actualizada en PostgreSQL</div>
    </div>` : ''}
  `;

  section.classList.remove('hidden');
  section.scrollIntoView({ behavior: 'smooth' });
}

function displayError(message) {
  const section = $('result-section');
  const header = $('result-header');
  const icon = $('result-icon');
  const status = $('result-status');
  const details = $('result-details');

  header.className = 'result-header error';
  icon.textContent = '❌';
  status.textContent = 'Error';
  status.className = 'status-text';
  details.innerHTML = `
    <div class="detail-item full-width">
      <div class="label">Detalle</div>
      <div class="value" style="font-size:14px; color: var(--danger)">${message}</div>
    </div>
  `;
  section.classList.remove('hidden');
}

// ─── FIREBASE RTDB ──────────────────────────────────────────
async function saveToFirebase(result, imageUrl) {
  try {
    const scanId = Date.now().toString();
    const scanData = {
      timestamp: new Date().toISOString(),
      codigo_acta: result.data?.codigo_acta || null,
      estado: result.estado || 'desconocido',
      imagen_url: imageUrl || null,
      p1: result.data?.p1 || 0,
      p2: result.data?.p2 || 0,
      p3: result.data?.p3 || 0,
      p4: result.data?.p4 || 0,
      votos_validos: result.data?.votos_validos || 0,
      votos_nulos: result.data?.votos_nulos || 0,
      votos_blancos: result.data?.votos_blancos || 0,
      observaciones: result.data?.observaciones || '',
      acta_actualizada: result.acta_actualizada || false,
      usuario: typeof state.user === 'string' ? state.user : state.user?.nombre || 'unknown',
    };

    await fetch(`${CONFIG.FIREBASE_RTDB}/scans/${scanId}.json`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(scanData),
    });

    console.log('[Firebase] Scan saved:', scanId);
  } catch (err) {
    console.warn('[Firebase] Save failed:', err);
  }
}

// ─── HISTORY ────────────────────────────────────────────────
async function loadHistory() {
  const list = $('history-list');
  list.innerHTML = '<p class="loading-text">Cargando historial...</p>';

  try {
    const res = await fetch(
      `${CONFIG.FIREBASE_RTDB}/scans.json?orderBy="$key"&limitToLast=30`
    );
    const data = await res.json();

    if (!data || Object.keys(data).length === 0) {
      list.innerHTML = `
        <div class="empty-state">
          <span>📭</span>
          <p>No hay escaneos registrados aún</p>
        </div>
      `;
      return;
    }

    const entries = Object.entries(data).reverse();
    list.innerHTML = entries.map(([id, scan]) => {
      let icon, badgeClass;
      if (scan.estado === 'transcrita') {
        icon = '✅'; badgeClass = 'badge-success';
      } else if (scan.estado === 'observada') {
        icon = '⚠️'; badgeClass = 'badge-warning';
      } else {
        icon = '❌'; badgeClass = 'badge-danger';
      }

      const date = scan.timestamp ? new Date(scan.timestamp).toLocaleString() : '-';

      return `
        <div class="history-item">
          <span class="hi-icon">${icon}</span>
          <div class="hi-info">
            <div class="hi-title">Acta ${scan.codigo_acta || 'S/N'}</div>
            <div class="hi-sub">${date} · ${scan.usuario || ''}</div>
          </div>
          <span class="hi-badge ${badgeClass}">${scan.estado || '?'}</span>
        </div>
      `;
    }).join('');

  } catch (err) {
    list.innerHTML = `
      <div class="empty-state">
        <span>⚠️</span>
        <p>Error cargando historial: ${err.message}</p>
      </div>
    `;
  }
}

// ─── SESSION RESTORE ────────────────────────────────────────
(function init() {
  const savedJwt = sessionStorage.getItem('jwt');
  const savedUser = sessionStorage.getItem('user');

  if (savedJwt) {
    state.jwt = savedJwt;
    try { state.user = JSON.parse(savedUser); } catch { state.user = savedUser; }
    showView('scanner');
  } else {
    showView('login');
  }
})();

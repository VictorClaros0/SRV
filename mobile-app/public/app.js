/* ══════════════════════════════════════════════════════════════
   Escáner Móvil de Actas - RRV Bolivia
   Claude Vision IA + Observaciones OEP + Firebase + Backend RRV
   ══════════════════════════════════════════════════════════════ */

const CONFIG = {
  FIREBASE_RTDB: 'https://lawdocs-75aef-default-rtdb.firebaseio.com',
  CLOUDINARY_CLOUD_NAME: 'dk5u8dljb',
  CLOUDINARY_API_KEY: '779231566857641',
  CLOUDINARY_API_SECRET: 'GFzuC8LmqeqAVZQfnOFBa5yrKzA',
  ANTHROPIC_API_KEY: 'TU_API_KEY_AQUI',
  // URL del backend local. Se puede sobreescribir desde Firebase RTDB en /config/backend_url
  BACKEND_URL: 'http://192.168.100.31:8080',
};

let state = {
  user: null,
  capturedFile: null,
  processing: false,
  currentScanId: null,
  currentImgUrl: null,
  currentFields: null,
  currentObservaciones: [],
  currentEsAnulada: false,
  currentDuplicado: null,
  backendUrl: CONFIG.BACKEND_URL,
};

const $ = id => document.getElementById(id);

// ── Cargar backend_url desde Firebase RTDB al inicio ─────────────────────────
async function loadBackendUrl() {
  try {
    var res = await fetch(CONFIG.FIREBASE_RTDB + '/config/backend_url.json');
    var url = await res.json();
    if (url && typeof url === 'string' && url.startsWith('http')) {
      state.backendUrl = url;
    }
  } catch (e) { /* usar default */ }
}

function showView(name) {
  document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
  $(('view-' + name)).classList.add('active');
  if (name === 'history') loadHistory();
}

// ── LOGIN ─────────────────────────────────────────────────────────────────────
$('btn-login').addEventListener('click', async () => {
  var btn = $('btn-login'), errorEl = $('login-error');
  var user = $('login-user').value.trim(), pass = $('login-pass').value;
  if (!user || !pass) { errorEl.textContent = 'Ingresa usuario y contraseña'; errorEl.classList.remove('hidden'); return; }
  btn.disabled = true; btn.querySelector('.btn-text').textContent = 'Verificando...'; errorEl.classList.add('hidden');
  try {
    var res = await fetch(CONFIG.FIREBASE_RTDB + '/config/users/' + encodeURIComponent(user) + '.json');
    var data = await res.json();
    if (!data || data.password !== pass) throw new Error('Usuario o contraseña incorrectos');
    state.user = { usuario: user, nombre: data.nombre, rol: data.rol };
    sessionStorage.setItem('user', JSON.stringify(state.user));
    showView('scanner');
  } catch (err) { errorEl.textContent = err.message; errorEl.classList.remove('hidden'); }
  finally { btn.disabled = false; btn.querySelector('.btn-text').textContent = 'Ingresar'; }
});

$('btn-logout').addEventListener('click', () => {
  state = { user:null, capturedFile:null, processing:false, currentScanId:null,
    currentImgUrl:null, currentFields:null, currentObservaciones:[], currentEsAnulada:false,
    backendUrl: state.backendUrl };
  sessionStorage.clear(); showView('login');
});
$('btn-history').addEventListener('click', () => showView('history'));
$('btn-back').addEventListener('click', () => showView('scanner'));

// ── CAPTURA DE IMAGEN ─────────────────────────────────────────────────────────
var fileInput = $('file-input'), imgPreview = $('img-preview'),
    placeholder = $('placeholder'), previewArea = $('preview-area'), btnUpload = $('btn-upload');

$('btn-capture').addEventListener('click', () => { fileInput.setAttribute('capture', 'environment'); fileInput.click(); });
$('btn-gallery').addEventListener('click', () => { fileInput.removeAttribute('capture'); fileInput.click(); });
previewArea.addEventListener('click', () => { if (!state.capturedFile) { fileInput.setAttribute('capture','environment'); fileInput.click(); }});

fileInput.addEventListener('change', (e) => {
  var file = e.target.files[0]; if (!file) return;
  state.capturedFile = file;
  var reader = new FileReader();
  reader.onload = (ev) => {
    imgPreview.src = ev.target.result; imgPreview.classList.remove('hidden');
    placeholder.classList.add('hidden'); previewArea.classList.add('has-image');
    btnUpload.classList.remove('hidden'); btnUpload.disabled = false;
    $('result-section').classList.add('hidden');
    $('final-status').classList.add('hidden');
    $('obs-section').classList.add('hidden');
    $('anulada-banner').classList.add('hidden');
    $('bordes-info').classList.add('hidden');
    $('duplicate-banner')?.classList.add('hidden');
    state.currentDuplicado = null;
  };
  reader.readAsDataURL(file);
});

// ── CLOUDINARY ────────────────────────────────────────────────────────────────
async function sha1(msg) {
  var buf = await crypto.subtle.digest('SHA-1', new TextEncoder().encode(msg));
  return Array.from(new Uint8Array(buf)).map(b => b.toString(16).padStart(2,'0')).join('');
}
async function uploadToCloudinary(file) {
  var ts = Math.floor(Date.now()/1000);
  var sig = await sha1('folder=rrv_actas&timestamp='+ts+CONFIG.CLOUDINARY_API_SECRET);
  var fd = new FormData();
  fd.append('file', file); fd.append('api_key', CONFIG.CLOUDINARY_API_KEY);
  fd.append('timestamp', ts); fd.append('signature', sig); fd.append('folder', 'rrv_actas');
  var res = await fetch('https://api.cloudinary.com/v1_1/'+CONFIG.CLOUDINARY_CLOUD_NAME+'/image/upload',{method:'POST',body:fd});
  if (!res.ok) throw new Error('Error subiendo imagen');
  return (await res.json()).secure_url;
}

// ── PDF → JPEG ────────────────────────────────────────────────────────────────
async function pdfToImage(file) {
  var arrayBuffer = await file.arrayBuffer();
  var pdf = await pdfjsLib.getDocument({ data: arrayBuffer }).promise;
  var page = await pdf.getPage(1);
  var viewport = page.getViewport({ scale: 2.0 });
  var canvas = document.createElement('canvas');
  canvas.width = viewport.width; canvas.height = viewport.height;
  await page.render({ canvasContext: canvas.getContext('2d'), viewport: viewport }).promise;
  return new Promise(function(resolve){ canvas.toBlob(resolve, 'image/jpeg', 0.92); });
}

function blobToBase64(blob) {
  return new Promise(function(resolve, reject) {
    var reader = new FileReader();
    reader.onload = function() { resolve(reader.result.split(',')[1]); };
    reader.onerror = reject;
    reader.readAsDataURL(blob);
  });
}

// ── PROMPT OCR con reglas OEP Bolivia ────────────────────────────────────────
var OCR_PROMPT =
  'Analiza esta acta electoral boliviana. Extrae los datos y detecta problemas físicos.\n\n' +
  'REGLAS DE LECTURA:\n' +
  '- Votos en casillas de 3 dígitos: |0|9|5| = 95\n' +
  '- Código de mesa: número largo en esquina superior izquierda\n' +
  '- El ÁREA CENTRAL es el recuadro grande del conteo de votos por candidato\n\n' +
  'DETECCIÓN DE OBSERVACIONES — responde true/false:\n' +
  '- anulada: el documento dice "ANULADA", "NULA" o tiene sello de anulación\n' +
  '- area_central_danada: el recuadro central está mojado, roto o manchado de forma que impide leer los votos; O la parte donde están los votos no se puede leer\n' +
  '- tiene_letras_en_numeros: hay letras (a,b,c...) en casillas que deben ser numéricas\n' +
  '- tiene_numeros_dudosos: hay un número sobreescrito o corregido con tachón encima que no permite saber con certeza qué número es\n' +
  '- tiene_tachaduras_no_aclaradas: hay tachaduras/borrones sin aclaración en la sección Observaciones\n' +
  '- solo_bordes_danados: SOLO bordes/esquinas mojados o arrugados pero el área central de votos está legible\n\n' +
  'IMPORTANTE:\n' +
  '- Esquina mojada pero centro legible → solo_bordes_danados=true, area_central_danada=false\n' +
  '- Votos no legibles por daño → area_central_danada=true\n' +
  '- Dice "ANULADA" → anulada=true\n\n' +
  'Devuelve ÚNICAMENTE este JSON:\n' +
  '{\n' +
  '  "codigo_mesa": número largo,\n' +
  '  "numero_mesa": número corto,\n' +
  '  "p1": votos Daenerys Targaryen,\n' +
  '  "p2": votos Sansa Stark,\n' +
  '  "p3": votos Robert Baratheon,\n' +
  '  "p4": votos Tyrion Lannister,\n' +
  '  "partido_1": nombre del partido que corresponde a P1,\n' +
  '  "partido_2": nombre del partido que corresponde a P2,\n' +
  '  "partido_3": nombre del partido que corresponde a P3,\n' +
  '  "partido_4": nombre del partido que corresponde a P4,\n' +
  '  "validos": total votos válidos,\n' +
  '  "blancos": votos blancos,\n' +
  '  "nulos": votos nulos,\n' +
  '  "habilitados": electores habilitados,\n' +
  '  "papeletas": papeletas en ánfora,\n' +
  '  "no_utilizadas": papeletas no utilizadas,\n' +
  '  "apertura_h": hora apertura,\n' +
  '  "apertura_m": minutos apertura,\n' +
  '  "cierre_h": hora cierre,\n' +
  '  "cierre_m": minutos cierre,\n' +
  '  "anulada": true/false,\n' +
  '  "observaciones": texto de observaciones del acta o "",\n' +
  '  "area_central_danada": true/false,\n' +
  '  "tiene_letras_en_numeros": true/false,\n' +
  '  "tiene_numeros_dudosos": true/false,\n' +
  '  "tiene_tachaduras_no_aclaradas": true/false,\n' +
  '  "solo_bordes_danados": true/false\n' +
  '}';

// ── CLAUDE VISION OCR ─────────────────────────────────────────────────────────
function updateProgress(pct, text) {
  $('progress-section').classList.remove('hidden');
  $('progress-bar').style.width = pct + '%';
  $('progress-text').textContent = text;
}

async function runOCR(file) {
  var ocrInput = file;
  var isPdf = file.name && file.name.toLowerCase().endsWith('.pdf');
  if (isPdf) {
    updateProgress(5, 'Convirtiendo PDF a imagen...');
    ocrInput = await pdfToImage(file);
  }
  updateProgress(20, 'Cargando...');
  var b64 = await blobToBase64(ocrInput);
  var resp = await fetch('https://api.anthropic.com/v1/messages', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'x-api-key': CONFIG.ANTHROPIC_API_KEY,
      'anthropic-version': '2023-06-01',
      'anthropic-dangerous-direct-browser-access': 'true',
    },
    body: JSON.stringify({
      model: 'claude-haiku-4-5-20251001',
      max_tokens: 1024,
      messages: [{ role: 'user', content: [
        { type: 'image', source: { type: 'base64', media_type: 'image/jpeg', data: b64 } },
        { type: 'text', text: OCR_PROMPT },
      ]}],
    }),
  });
  var raw = await resp.json();
  if (!resp.ok) throw new Error(raw.error && raw.error.message ? raw.error.message : 'Claude API error ' + resp.status);
  var text = (raw.content && raw.content[0] && raw.content[0].text) || '';
  var match = text.match(/\{[\s\S]*\}/);
  if (!match) throw new Error('Claude no devolvió JSON válido');
  return JSON.parse(match[0]);
}

// ── DETECCIÓN DE OBSERVACIONES OEP ───────────────────────────────────────────
// Regla: solo_bordes_danados NO genera observación — el acta se procesa normalmente.
function detectarObservaciones(result) {
  var obs = [];
  if (result.area_central_danada)
    obs.push({ tipo: 'AREA_CENTRAL_DAÑADA',
      texto: 'El área central del recuadro de conteo de votos está dañada, mojada o ilegible. Los datos de votación no son verificables con certeza. Causal de observación según Ley 026 Art. 190.' });
  if (result.tiene_letras_en_numeros)
    obs.push({ tipo: 'LETRAS_EN_CASILLAS_NUMÉRICAS',
      texto: 'Se detectaron letras en casillas que deben contener únicamente números. Acta mal llenada. La Guía OEP exige números claros en cada casilla.' });
  if (result.tiene_numeros_dudosos)
    obs.push({ tipo: 'NÚMERO_SOBREESCRITO_ILEGIBLE',
      texto: 'Existe un número corregido con tachón o sobreescrito que no puede distinguirse con certeza. Dato dudoso — no puede entrar al cómputo sin revisión humana.' });
  if (result.tiene_tachaduras_no_aclaradas)
    obs.push({ tipo: 'TACHADURA_SIN_ACLARACIÓN',
      texto: 'Hay tachaduras o correcciones no señaladas en la sección de Observaciones. Causal de nulidad según Ley 026: alteración de datos sin señalamiento en Observaciones.' });
  // solo_bordes_danados: NO genera observación — solo informativo
  return obs;
}

// ── VALIDACIÓN ARITMÉTICA ─────────────────────────────────────────────────────
function validateActa(d) {
  var alerts = [];
  if (d.anulada) alerts.push({ t:'err', m:'⛔ Acta marcada como ANULADA — no puede ingresar al cómputo.' });
  var pv = [d.p1,d.p2,d.p3,d.p4].filter(function(v){return v!=null;});
  var sum = pv.reduce(function(a,b){return a+b;},0);
  if (d.validos!=null && pv.length>0 && sum!==d.validos)
    alerts.push({ t:'err', m:'Inconsistencia: Σ candidatos ('+sum+') ≠ válidos ('+d.validos+')' });
  if (d.validos!=null && d.blancos!=null && d.nulos!=null && d.papeletas!=null) {
    var exp = d.validos+d.blancos+d.nulos;
    if (exp!==d.papeletas) alerts.push({ t:'err', m:'Papeletas ('+d.papeletas+') ≠ válidos+blancos+nulos ('+exp+')' });
  }
  if (pv.length===0) alerts.push({ t:'err', m:'No se pudieron extraer votos' });
  if (!d.anulada && alerts.filter(function(a){return a.t==='err';}).length===0 && pv.length>0)
    alerts.push({ t:'ok', m:'Datos aritméticos consistentes ✓' });
  return alerts;
}

// ── PROCESO PRINCIPAL ─────────────────────────────────────────────────────────
btnUpload.addEventListener('click', async () => {
  if (!state.capturedFile || state.processing) return;
  state.processing = true;
  var btnText = btnUpload.querySelector('.btn-text'), btnLoader = btnUpload.querySelector('.btn-loader');
  btnText.classList.add('hidden'); btnLoader.classList.remove('hidden');
  btnUpload.classList.add('processing'); btnUpload.disabled = true;
  $('result-section').classList.add('hidden');
  $('final-status').classList.add('hidden');
  $('obs-section').classList.add('hidden');
  $('anulada-banner').classList.add('hidden');
  $('bordes-info').classList.add('hidden');

  try {
    updateProgress(10, 'Iniciando análisis...');
    var actaData = await runOCR(state.capturedFile);

    updateProgress(65, 'Verificando duplicados en Firebase...');
    state.currentDuplicado = null;
    $('duplicate-banner')?.classList.add('hidden');
    if (actaData.codigo_mesa) {
      try {
        // Verificación de duplicados directo en Firebase (funciona desde cualquier red)
        var idxRes = await fetch(CONFIG.FIREBASE_RTDB + '/mesa_index/' + actaData.codigo_mesa + '.json');
        var idxData = await idxRes.json();
        if (idxData && idxData.estado) {
          state.currentDuplicado = idxData;
          var banner = $('duplicate-banner');
          if (banner) {
            banner.innerHTML = '<strong>⚠️ Acta duplicada:</strong> Ya existe un acta para la mesa <strong>' + actaData.codigo_mesa + '</strong> registrada desde <strong>' + (idxData.fuente || 'scanner') + '</strong>. No puede aceptarse nuevamente; marque Observada o Rechazada.';
            banner.classList.remove('hidden');
          }
        }
      } catch(e) { console.warn('[DUP CHECK Firebase]', e); }
    }

    updateProgress(75, 'Detectando observaciones OEP...');
    var obsDetectadas = detectarObservaciones(actaData);
    state.currentObservaciones = obsDetectadas;
    state.currentFields = actaData;
    state.currentEsAnulada = !!actaData.anulada;

    updateProgress(85, 'Validando datos...');
    var alerts = validateActa(actaData);

    updateProgress(93, 'Subiendo imagen...');
    state.currentImgUrl = null;
    try { state.currentImgUrl = await uploadToCloudinary(state.capturedFile); } catch(e){ console.warn(e); }

    state.currentScanId = Date.now().toString();
    updateProgress(100, 'Listo ✓');
    setTimeout(function(){ $('progress-section').classList.add('hidden'); }, 1000);

    // Mostrar banner si acta anulada
    if (actaData.anulada) {
      $('anulada-banner').classList.remove('hidden');
    }

    // Mostrar info bordes dañados (solo informativo, no bloquea)
    if (actaData.solo_bordes_danados && !actaData.area_central_danada) {
      $('bordes-info').classList.remove('hidden');
    }

    // Mostrar observaciones si las hay
    if (obsDetectadas.length > 0) {
      var obsHtml = obsDetectadas.map(function(o) {
        return '<div class="obs-item"><div class="obs-badge">'+o.tipo+'</div>' +
               '<div class="obs-text">'+o.texto+'</div></div>';
      }).join('');
      $('obs-list').innerHTML = obsHtml;
      $('obs-count').textContent = obsDetectadas.length + ' observación' + (obsDetectadas.length>1?'es':'') + ' detectada' + (obsDetectadas.length>1?'s':'');
      $('obs-section').classList.remove('hidden');
    }

    fillEditableFields(actaData);
    renderAlerts(alerts);
    $('result-section').classList.remove('hidden');
    $('reason-group').style.display = 'none';
    $('r-reason').value = '';
    $('final-status').classList.add('hidden');

    // Botón observada: solo si hay observaciones y NO está anulada
    var btnObs = $('btn-mark-observed');
    if (btnObs) {
      btnObs.style.display = (obsDetectadas.length > 0 && !actaData.anulada) ? 'block' : 'none';
    }

    // Botón confirmar anulada: solo si acta anulada
    var btnAnulada = $('btn-confirm-anulada');
    if (btnAnulada) {
      btnAnulada.style.display = actaData.anulada ? 'block' : 'none';
    }

    // Aceptar: siempre disponible salvo si acta anulada o existe duplicado
    var acceptDisabled = actaData.anulada || !!state.currentDuplicado;
    $('btn-accept').disabled = acceptDisabled;
    $('btn-accept').style.opacity = acceptDisabled ? '0.6' : '1';
    if (actaData.anulada) $('btn-accept').style.display = 'none';
    else $('btn-accept').style.display = '';

    // Rechazar siempre disponible (excepto si anulada, que usamos btn-confirm-anulada)
    $('btn-reject').style.display = actaData.anulada ? 'none' : '';
    $('btn-reject').disabled = false;

    $('result-section').scrollIntoView({ behavior:'smooth' });

  } catch(err) {
    $('progress-section').classList.add('hidden');
    alert('Error: ' + err.message);
  } finally {
    state.processing = false;
    btnText.classList.remove('hidden'); btnLoader.classList.add('hidden');
    btnUpload.classList.remove('processing'); btnUpload.disabled = false;
  }
});

function fillEditableFields(d) {
  $('r-codigo').value = d.codigo_mesa || '';
  $('r-mesa').value = d.numero_mesa || '';
  $('r-partido-1').value = d.partido_1 || 'Partido 1';
  $('r-partido-2').value = d.partido_2 || 'Partido 2';
  $('r-partido-3').value = d.partido_3 || 'Partido 3';
  $('r-partido-4').value = d.partido_4 || 'Partido 4';
  $('r-p1').value = d.p1 != null ? d.p1 : '';
  $('r-p2').value = d.p2 != null ? d.p2 : '';
  $('r-p3').value = d.p3 != null ? d.p3 : '';
  $('r-p4').value = d.p4 != null ? d.p4 : '';
  $('r-validos').value = d.validos != null ? d.validos : '';
  $('r-nulos').value = d.nulos != null ? d.nulos : '';
  $('r-blancos').value = d.blancos != null ? d.blancos : '';
  $('r-papeletas').value = d.papeletas != null ? d.papeletas : '';
  $('r-apertura').value = d.apertura_h != null ? (String(d.apertura_h).padStart(2,'0')+':'+(d.apertura_m||0).toString().padStart(2,'0')) : '';
  $('r-cierre').value = d.cierre_h != null ? (String(d.cierre_h).padStart(2,'0')+':'+(d.cierre_m||0).toString().padStart(2,'0')) : '';
}

function renderAlerts(alerts) {
  $('validation-alerts').innerHTML = alerts.map(function(a) {
    return '<div class="val-alert alert-'+(a.t==='ok'?'ok':a.t==='warn'?'warn':'err')+'">' +
      '<span>'+(a.t==='ok'?'✅':a.t==='warn'?'⚠️':'❌')+'</span><span>'+a.m+'</span></div>';
  }).join('');
}

function getEditedValues() {
  return {
    codigo_mesa: parseInt($('r-codigo').value)||0,
    numero_mesa: parseInt($('r-mesa').value)||0,
    partido_1: $('r-partido-1').value.trim()||'Partido 1',
    partido_2: $('r-partido-2').value.trim()||'Partido 2',
    partido_3: $('r-partido-3').value.trim()||'Partido 3',
    partido_4: $('r-partido-4').value.trim()||'Partido 4',
    p1: parseInt($('r-p1').value)||0,
    p2: parseInt($('r-p2').value)||0,
    p3: parseInt($('r-p3').value)||0,
    p4: parseInt($('r-p4').value)||0,
    validos: parseInt($('r-validos').value)||0,
    nulos: parseInt($('r-nulos').value)||0,
    blancos: parseInt($('r-blancos').value)||0,
    papeletas: parseInt($('r-papeletas').value)||0,
    hora_apertura: $('r-apertura').value.trim()||null,
    hora_cierre: $('r-cierre').value.trim()||null,
  };
}

// ── LLAMAR AL BACKEND LOCAL ───────────────────────────────────────────────────
async function postToBackend(path, data) {
  var url = state.backendUrl + path;
  var resp = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  return resp;
}

// ── BOTÓN OBSERVADA ───────────────────────────────────────────────────────────
var btnMarkObs = $('btn-mark-observed');
if (btnMarkObs) {
  btnMarkObs.addEventListener('click', function() { saveScan('observada', null); });
}

// ── BOTÓN CONFIRMAR ANULADA ───────────────────────────────────────────────────
var btnConfirmAnulada = $('btn-confirm-anulada');
if (btnConfirmAnulada) {
  btnConfirmAnulada.addEventListener('click', function() { saveScan('anulada', null); });
}

// ── BOTONES ACEPTAR / RECHAZAR ────────────────────────────────────────────────
$('btn-reject').addEventListener('click', () => {
  var rg = $('reason-group');
  if (rg.style.display === 'none') {
    rg.style.display = 'block'; $('r-reason').focus();
    $('btn-reject').textContent = '❌ Confirmar Rechazo'; return;
  }
  var reason = $('r-reason').value.trim();
  if (!reason) { alert('Ingresa la razón del rechazo'); return; }
  saveScan('rechazada', reason);
});
$('btn-accept').addEventListener('click', () => { saveScan('aceptada', null); });

// ── GUARDAR ───────────────────────────────────────────────────────────────────
async function saveScan(decision, reason) {
  $('btn-accept').disabled = true; $('btn-reject').disabled = true;
  if ($('btn-mark-observed')) $('btn-mark-observed').disabled = true;
  if ($('btn-confirm-anulada')) $('btn-confirm-anulada').disabled = true;

  var vals = getEditedValues();
  var now = new Date();

  // Toda la comunicación va por Firebase (sin backend local — funciona desde celular en cualquier red)
  if (decision === 'observada' || decision === 'anulada') {
    var obsParaLog = decision === 'anulada'
      ? [{ tipo: 'ACTA_ANULADA', descripcion: 'El acta tiene marcación explícita de ANULADA. No puede ingresar al cómputo.' }]
      : state.currentObservaciones.map(function(o){ return { tipo:o.tipo, descripcion:o.texto }; });
    try {
      await fetch(CONFIG.FIREBASE_RTDB+'/logs_observadas/'+state.currentScanId+'.json', {
        method:'PUT', headers:{'Content-Type':'application/json'},
        body: JSON.stringify({
          id: state.currentScanId, timestamp: now.toISOString(),
          hora: now.toLocaleTimeString('es-BO'), fuente: 'mobile_scanner',
          estado: decision, usuario: state.user ? state.user.usuario : null,
          codigo_acta: vals.codigo_mesa,
          imagen_url: state.currentImgUrl || null,
          anulada: decision === 'anulada',
          razon_principal: obsParaLog.length>0 ? obsParaLog[0].descripcion : 'Acta observada',
          observaciones_detectadas: obsParaLog,
          datos: vals,
        }),
      });
    } catch(e){ console.warn('Firebase log observada:', e); }
    showFinalStatus(decision === 'anulada' ? 'anulada' : 'observada', vals, null);

  } else if (decision === 'aceptada') {
    if (state.currentDuplicado) {
      alert('Acta duplicada detectada. No puede aceptarse nuevamente. Marque Observada o Rechazada.');
      $('btn-accept').disabled = false; $('btn-reject').disabled = false;
      if ($('btn-mark-observed')) $('btn-mark-observed').disabled = false;
      if ($('btn-confirm-anulada')) $('btn-confirm-anulada').disabled = false;
      return;
    }
    try {
      // Guardar datos completos en Firebase /mobile_scans/ para el dashboard
      await fetch(CONFIG.FIREBASE_RTDB+'/mobile_scans/'+state.currentScanId+'.json', {
        method:'PUT', headers:{'Content-Type':'application/json'},
        body: JSON.stringify({
          id: state.currentScanId, timestamp: now.toISOString(),
          hora: now.toLocaleTimeString('es-BO'), fuente: 'mobile_scanner',
          estado: 'aceptada', usuario: state.user ? state.user.usuario : null,
          codigo_acta: vals.codigo_mesa, imagen_url: state.currentImgUrl||null,
          // Votos planos (para que el dashboard los grafique directamente)
          p1: vals.p1, p2: vals.p2, p3: vals.p3, p4: vals.p4,
          partido_1: vals.partido_1, partido_2: vals.partido_2,
          partido_3: vals.partido_3, partido_4: vals.partido_4,
          votos_validos: vals.validos, votos_blancos: vals.blancos,
          votos_nulos: vals.nulos, papeletas: vals.papeletas,
          datos: vals,
        }),
      });
      // Índice por mesa para detección rápida de duplicados
      await fetch(CONFIG.FIREBASE_RTDB+'/mesa_index/'+vals.codigo_mesa+'.json', {
        method:'PUT', headers:{'Content-Type':'application/json'},
        body: JSON.stringify({
          scan_id: state.currentScanId, timestamp: now.toISOString(),
          fuente: 'mobile_scanner', estado: 'aceptada',
        }),
      });
    } catch(e){ console.warn('Firebase aceptada:', e); }
    showFinalStatus('aceptada', vals, null);

  } else {
    // rechazada — solo Firebase
    try {
      await fetch(CONFIG.FIREBASE_RTDB+'/mobile_scans/'+state.currentScanId+'.json', {
        method:'PUT', headers:{'Content-Type':'application/json'},
        body: JSON.stringify({
          id: state.currentScanId, timestamp: now.toISOString(),
          hora: now.toLocaleTimeString('es-BO'), fuente: 'mobile_scanner',
          estado: 'rechazada', usuario: state.user ? state.user.usuario : null,
          codigo_acta: vals.codigo_mesa, imagen_url: state.currentImgUrl||null,
          razon_rechazo: reason||null, datos: vals,
        }),
      });
    } catch(e){ console.warn('Firebase rechazada:', e); }
    showFinalStatus('rechazada', vals, reason);
  }
}

function showFinalStatus(decision, vals, reason) {
  var fs = $('final-status');
  if (decision === 'aceptada') {
    fs.className = 'final-status final-accepted';
    fs.innerHTML = '<div class="fs-icon">✅</div><div class="fs-title">Acta Aceptada — RRV</div>' +
      '<div class="fs-detail">Mesa '+(vals.codigo_mesa||'S/N')+' registrada en conteo rápido</div>' +
      '<div class="fs-meta">'+(vals.partido_1||'P1')+':'+vals.p1+' · '+(vals.partido_2||'P2')+':'+vals.p2+' · '+(vals.partido_3||'P3')+':'+vals.p3+' · '+(vals.partido_4||'P4')+':'+vals.p4+
      ' · Válidos:'+vals.validos+'</div>';
  } else if (decision === 'observada') {
    fs.className = 'final-status final-observed';
    var obsTexts = state.currentObservaciones.map(function(o){ return '<li>'+o.texto+'</li>'; }).join('');
    fs.innerHTML = '<div class="fs-icon">📋</div><div class="fs-title">Acta Observada — No entra al cómputo</div>' +
      '<div class="fs-detail">Guardada en log de observadas para revisión humana</div>' +
      (obsTexts ? '<ul class="obs-list-final">'+obsTexts+'</ul>' : '');
  } else if (decision === 'anulada') {
    fs.className = 'final-status final-anulada';
    fs.innerHTML = '<div class="fs-icon">🚫</div><div class="fs-title">Acta Anulada</div>' +
      '<div class="fs-detail">Registrada en log. El TED debe revisar el material físico.</div>';
  } else {
    fs.className = 'final-status final-rejected';
    fs.innerHTML = '<div class="fs-icon">❌</div><div class="fs-title">Acta Rechazada</div>' +
      '<div class="fs-detail">Mesa '+(vals.codigo_mesa||'S/N')+' no tomada en cuenta</div>' +
      '<div class="fs-reason"><strong>Razón:</strong> '+reason+'</div>';
  }
  fs.classList.remove('hidden');
  fs.scrollIntoView({ behavior:'smooth' });
}

// ── HISTORIAL ─────────────────────────────────────────────────────────────────
async function loadHistory() {
  var list = $('history-list');
  list.innerHTML = '<p class="loading-text">Cargando historial...</p>';
  try {
    var [scansRes, obsRes] = await Promise.all([
      fetch(CONFIG.FIREBASE_RTDB+'/mobile_scans.json?orderBy="$key"&limitToLast=20'),
      fetch(CONFIG.FIREBASE_RTDB+'/logs_observadas.json?orderBy="$key"&limitToLast=10'),
    ]);
    var scans = await scansRes.json() || {};
    var observadas = await obsRes.json() || {};

    var all = [];
    Object.values(scans).forEach(function(s){ all.push(Object.assign({}, s, {_tipo:'scan'})); });
    Object.values(observadas).forEach(function(s){ if (s.fuente==='mobile_scanner') all.push(Object.assign({}, s, {_tipo:'observada'})); });
    all.sort(function(a,b){ return (b.timestamp||'').localeCompare(a.timestamp||''); });

    if (all.length === 0) {
      list.innerHTML = '<div class="empty-state"><span>📭</span><p>No hay escaneos aún</p></div>'; return;
    }
    list.innerHTML = all.map(function(s) {
      var estado = s._tipo==='observada' ? (s.estado||'observada') : (s.estado||'?');
      var icon = estado==='aceptada'?'✅':estado==='observada'?'📋':estado==='anulada'?'🚫':'❌';
      var badge = estado==='aceptada'?'badge-success':estado==='observada'?'badge-warning':estado==='anulada'?'badge-danger':'badge-danger';
      var date = s.timestamp ? new Date(s.timestamp).toLocaleString('es-BO') : '-';
      var datos = s.datos || {};
      var razon = (s.razon_principal||s.razon_rechazo||'').substring(0,60);
      var mesa = s.codigo_acta || (datos.codigo_mesa) || 'S/N';
      return '<div class="history-item">' +
        '<span class="hi-icon">'+icon+'</span>' +
        '<div class="hi-info"><div class="hi-title">Acta '+mesa+'</div>' +
        '<div class="hi-sub">'+date+' · '+(s.usuario||s.fuente||'')+'</div>' +
        '<div class="hi-votes">P1:'+(datos.p1||0)+' P2:'+(datos.p2||0)+' P3:'+(datos.p3||0)+' P4:'+(datos.p4||0)+'</div>' +
        (razon?'<div class="hi-reason">'+razon+'</div>':'') +
        '</div>' +
        '<span class="hi-badge '+badge+'">'+estado+'</span></div>';
    }).join('');
  } catch(err) {
    list.innerHTML = '<div class="empty-state"><span>⚠️</span><p>Error: '+err.message+'</p></div>';
  }
}

// ── INIT ──────────────────────────────────────────────────────────────────────
(function() {
  loadBackendUrl();
  var saved = sessionStorage.getItem('user');
  if (saved) { try { state.user = JSON.parse(saved); showView('scanner'); } catch(e){ showView('login'); } }
  else { showView('login'); }
})();

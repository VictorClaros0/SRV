const webhookBase = window.location.origin;
document.getElementById('webhook-url').textContent = webhookBase + '/webhook';

const curlCmd =
  `curl -X POST ${webhookBase}/webhook ` +
  `-H "Content-Type: application/json" ` +
  `-d "{\\"from\\":\\"+59165707079\\",\\"body\\":\\"RRV|MESA=10101001001|P1=120|P2=95|P3=40|P4=30|BLANCOS=4|NULOS=3|TOTAL=292\\"}"`;
document.getElementById('curl-example').textContent = curlCmd;

function esc(str) {
  if (str === null || str === undefined) return '';
  return String(str)
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function fmtDate(iso) {
  return new Date(iso).toLocaleString('es-BO');
}

function statusBadge(status, isDuplicate) {
  const map = {
    VALIDADA: 'badge-validada',
    PENDIENTE_REVISION: 'badge-pendiente',
    RECHAZADA: 'badge-rechazada',
    NO_AUTORIZADO: 'badge-noauth',
  };
  const cls = map[status] || 'badge-default';
  const dup = isDuplicate ? ' <span class="badge-dup">DUPLICADO</span>' : '';
  return `<span class="badge ${cls}">${esc(status)}</span>${dup}`;
}

function renderSmsCard(m) {
  const p = m.parsed;
  const hasFields = p && p.valid && p.fields;

  let votos = '';
  if (hasFields) {
    const f = p.fields;
    const matchIcon = p.totalsMatch ? '✓' : '✗';
    const matchCls = p.totalsMatch ? 'match-ok' : 'match-fail';

    const obsHtml = f.OBS
      ? `<div class="obs-item"><span class="vk">OBSERVACIONES</span><span class="obs-val">${esc(f.OBS)}</span></div>`
      : '';

    votos = `
      <div class="votos-grid">
        <div class="voto-item"><span class="vk">MESA</span><span class="vv">${esc(f.MESA)}</span></div>
        <div class="voto-item"><span class="vk">P1</span><span class="vv">${f.P1}</span></div>
        <div class="voto-item"><span class="vk">P2</span><span class="vv">${f.P2}</span></div>
        <div class="voto-item"><span class="vk">P3</span><span class="vv">${f.P3}</span></div>
        <div class="voto-item"><span class="vk">P4</span><span class="vv">${f.P4}</span></div>
        <div class="voto-item"><span class="vk">BLANCOS</span><span class="vv">${f.BLANCOS}</span></div>
        <div class="voto-item"><span class="vk">NULOS</span><span class="vv">${f.NULOS}</span></div>
        <div class="voto-item total-item"><span class="vk">TOTAL declarado</span><span class="vv">${f.TOTAL}</span></div>
        <div class="voto-item total-item ${matchCls}">
          <span class="vk">TOTAL calculado</span>
          <span class="vv">${p.calculated} ${matchIcon}</span>
        </div>
      </div>
      ${obsHtml}`;
  }

  const reasonHtml = p?.reason
    ? `<div class="card-reason">⚠ ${esc(p.reason)}</div>`
    : '';

  const authIcon = m.senderAuthorized ? '✓ Autorizado' : '✗ No autorizado';
  const authCls = m.senderAuthorized ? 'auth-ok' : 'auth-fail';

  return `
    <div class="sms-card status-${(m.status || '').toLowerCase()}">
      <div class="card-header">
        <div class="card-left">
          <span class="card-id">#${m.id}</span>
          <span class="card-date">${fmtDate(m.receivedAt)}</span>
        </div>
        <div class="card-right">
          ${statusBadge(m.status, m.isDuplicate)}
        </div>
      </div>
      <div class="card-body">
        <div class="card-meta">
          <span class="meta-label">Remitente:</span>
          <span class="meta-val">${esc(m.sender)}</span>
          <span class="meta-auth ${authCls}">${authIcon}</span>
        </div>
        <div class="card-sms-raw">
          <span class="meta-label">SMS raw:</span>
          <code>${esc(m.smsBody)}</code>
        </div>
        ${reasonHtml}
        ${votos}
      </div>
    </div>`;
}

async function loadMessages() {
  try {
    const res = await fetch('/messages');
    const data = await res.json();

    document.getElementById('last-updated').textContent =
      'Actualizado: ' + new Date().toLocaleTimeString('es-BO');

    const counts = { VALIDADA: 0, PENDIENTE_REVISION: 0, RECHAZADA: 0, NO_AUTORIZADO: 0 };
    data.forEach((m) => { if (m.status in counts) counts[m.status]++; });
    document.getElementById('stat-total').textContent = data.length;
    document.getElementById('stat-validada').textContent = counts.VALIDADA;
    document.getElementById('stat-pendiente').textContent = counts.PENDIENTE_REVISION;
    document.getElementById('stat-rechazada').textContent = counts.RECHAZADA;
    document.getElementById('stat-noauth').textContent = counts.NO_AUTORIZADO;

    const cardsEl = document.getElementById('sms-cards');
    if (data.length === 0) {
      cardsEl.innerHTML = '<p class="empty-cards">Sin mensajes todavía.</p>';
    } else {
      cardsEl.innerHTML = data.map(renderSmsCard).join('');
    }

    const tbody = document.getElementById('messages-body');
    if (data.length === 0) {
      tbody.innerHTML = '<tr><td colspan="6" class="empty">Sin mensajes todavía.</td></tr>';
    } else {
      tbody.innerHTML = data.map((m) => `
        <tr>
          <td class="col-id">${m.id}</td>
          <td class="col-date">${esc(fmtDate(m.receivedAt))}</td>
          <td class="col-sender ${m.senderAuthorized ? 'auth-ok' : 'auth-fail'}">${esc(m.sender)}</td>
          <td>${statusBadge(m.status, m.isDuplicate)}</td>
          <td class="col-message">${esc(m.smsBody)}</td>
          <td class="col-raw"><pre>${esc(JSON.stringify(m.raw, null, 2))}</pre></td>
        </tr>`).join('');
    }
  } catch (err) {
    console.error('Error cargando mensajes:', err);
  }
}

loadMessages();
setInterval(loadMessages, 3000);

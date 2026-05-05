const express = require('express');
const path = require('path');
const http = require('http');

const app = express();
const PORT = 3001;

// URL del SRV-ocr2
const OCR2_HOST = 'localhost';
const OCR2_PORT = 8081;
const OCR2_INBOUND_PATH = '/api/rrv/sms/inbound';
const OCR2_NUMEROS_PATH = '/api/v1/rrv/sms/numeros';

const messages = [];
let nextId = 1;

const AUTHORIZED_SENDERS = ['65707079'];

function normalizeSender(raw) {
  if (!raw) return null;
  let s = String(raw).trim().replace(/\s/g, '');
  if (s.startsWith('+591')) s = s.slice(4);
  else if (s.startsWith('591') && s.length > 3) s = s.slice(3);
  return s;
}

function extractFromText(text) {
  if (!text) return { sender: null, smsBody: null };
  const str = String(text);
  const desdeMatch = str.match(/Desde\s*:\s*(\+?[\d]+)\s*\n?([\s\S]*)/i);
  if (desdeMatch) {
    return { sender: desdeMatch[1].trim(), smsBody: desdeMatch[2].trim() };
  }
  return { sender: null, smsBody: str.trim() };
}

function normalizeSmsBody(body) {
  if (!body) return body;
  return body.replace(/@/g, '|');
}

// Parse RRV SMS: RRV|MESA=X|P1=X|...|TOTAL=X[|OBS=texto opcional]
function parseSms(rawBody) {
  const body = normalizeSmsBody(rawBody);
  const parts = body.split('|').map((p) => p.trim()).filter(Boolean);

  if (!parts[0] || parts[0].toUpperCase() !== 'RRV') {
    return { valid: false, reason: 'El mensaje no comienza con RRV', fields: null };
  }

  const rawFields = {};
  for (let i = 1; i < parts.length; i++) {
    const eq = parts[i].indexOf('=');
    if (eq === -1) continue;
    const key = parts[i].slice(0, eq).toUpperCase().trim();
    const val = parts[i].slice(eq + 1).trim();
    rawFields[key] = val;
  }

  const required = ['MESA', 'P1', 'P2', 'P3', 'P4', 'BLANCOS', 'NULOS', 'TOTAL'];
  const missing = required.filter((k) => !(k in rawFields));
  if (missing.length > 0) {
    return { valid: false, reason: `Campos faltantes: ${missing.join(', ')}`, fields: rawFields };
  }

  const numericKeys = ['P1', 'P2', 'P3', 'P4', 'BLANCOS', 'NULOS', 'TOTAL'];
  const fields = { MESA: rawFields.MESA };

  for (const k of numericKeys) {
    const n = Number(rawFields[k]);
    if (!Number.isInteger(n)) {
      return { valid: false, reason: `${k} no es entero: "${rawFields[k]}"`, fields: rawFields };
    }
    if (n < 0) {
      return { valid: false, reason: `${k} es negativo`, fields: rawFields };
    }
    fields[k] = n;
  }

  // OBS es opcional
  if (rawFields.OBS !== undefined) {
    fields.OBS = rawFields.OBS;
  }

  const calculated = fields.P1 + fields.P2 + fields.P3 + fields.P4 + fields.BLANCOS + fields.NULOS;
  const declared = fields.TOTAL;
  const totalsMatch = calculated === declared;

  return {
    valid: true,
    fields,
    calculated,
    declared,
    totalsMatch,
    reason: totalsMatch ? null : `Total declarado (${declared}) ≠ calculado (${calculated})`,
  };
}

function determineStatus(parseResult, senderAuthorized) {
  if (!senderAuthorized) return 'NO_AUTORIZADO';
  if (!parseResult.valid) return 'RECHAZADA';
  if (parseResult.totalsMatch) return 'VALIDADA';
  return 'PENDIENTE_REVISION';
}

// Registrar número autorizado en SRV-ocr2 al arrancar
function registrarNumeroEnOCR2(numero) {
  const body = JSON.stringify({ numero });
  const req = http.request(
    {
      hostname: OCR2_HOST,
      port: OCR2_PORT,
      path: OCR2_NUMEROS_PATH,
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(body) },
    },
    (res) => {
      let data = '';
      res.on('data', (chunk) => (data += chunk));
      res.on('end', () => {
        console.log(`[OCR2] Número ${numero} registrado en SRV-ocr2: status=${res.statusCode} resp=${data}`);
      });
    }
  );
  req.on('error', (e) => console.warn(`[OCR2] No se pudo registrar número (¿está corriendo SRV-ocr2?): ${e.message}`));
  req.write(body);
  req.end();
}

// Reenviar SMS al SRV-ocr2 y actualizar el campo srvStatus en la entrada
function reenviarAOCR2(sender, smsBody, entry) {
  const payload = JSON.stringify({ from: sender, body: smsBody });
  const req = http.request(
    {
      hostname: OCR2_HOST,
      port: OCR2_PORT,
      path: OCR2_INBOUND_PATH,
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(payload) },
    },
    (res) => {
      let data = '';
      res.on('data', (chunk) => (data += chunk));
      res.on('end', () => {
        let parsed = {};
        try {
          parsed = JSON.parse(data);
        } catch (_) {
          parsed = {};
        }
        const savedInSrv = res.statusCode >= 200 && res.statusCode < 300 && parsed.ok === true && parsed.ocr_guardado === true;
        entry.srvStatus = savedInSrv ? 'GUARDADO' : 'OMITIDO';
        entry.srvResponse = parsed;
        console.log(`[OCR2] Reenviado: status=${res.statusCode} srvStatus=${entry.srvStatus} resp=${data}`);
      });
    }
  );
  req.on('error', (e) => {
    entry.srvStatus = 'ERROR';
    console.error(`[OCR2] Error al reenviar: ${e.message}`);
  });
  req.write(payload);
  req.end();
}

app.use((req, res, next) => {
  res.header('Access-Control-Allow-Origin', '*');
  res.header('Access-Control-Allow-Methods', 'GET,POST,HEAD,OPTIONS');
  res.header('Access-Control-Allow-Headers', 'Content-Type,Authorization');
  if (req.method === 'OPTIONS') return res.sendStatus(204);
  next();
});
app.use(express.json());
app.use(express.urlencoded({ extended: true }));
app.use(express.static(path.join(__dirname, 'public')));

app.get('/health', (req, res) => res.json({ status: 'ok' }));

app.get('/messages', (req, res) => res.json(messages));

app.post('/webhook', (req, res) => {
  const body = req.body;

  let sender = body.from || body.sender || body.phone || body.number || body.From || null;
  let smsBody = body.body || body.message || body.content || body.sms || body.Body || null;

  if (!sender && !smsBody && body.text) {
    const extracted = extractFromText(body.text);
    sender = extracted.sender;
    smsBody = extracted.smsBody;
  } else if (!smsBody && body.text) {
    smsBody = body.text;
  }

  const normalizedSender = normalizeSender(sender);
  const senderAuthorized = AUTHORIZED_SENDERS.includes(normalizedSender);

  const parseResult = smsBody
    ? parseSms(smsBody)
    : { valid: false, reason: 'Sin cuerpo de SMS', fields: null };

  const status = determineStatus(parseResult, senderAuthorized);

  const isDuplicate =
    senderAuthorized &&
    parseResult.valid &&
    messages.some(
      (m) =>
        m.parsed?.fields?.MESA === parseResult.fields?.MESA &&
        m.status === 'VALIDADA'
    );

  // Seguridad: solo se guarda en SRV si autorizado + formato válido + totales correctos + no duplicado
  const guardarEnSrv = senderAuthorized && parseResult.valid && parseResult.totalsMatch === true && !isDuplicate;

  const entry = {
    id: nextId++,
    receivedAt: new Date().toISOString(),
    sender: sender || 'Unknown',
    senderNormalized: normalizedSender,
    senderAuthorized,
    smsBody: smsBody || '(vacío)',
    parsed: parseResult,
    status,
    isDuplicate: isDuplicate || false,
    srvStatus: guardarEnSrv ? 'PENDIENTE' : 'OMITIDO',
    raw: body,
  };

  messages.unshift(entry);

  console.log('--- Webhook recibido ---');
  console.log('Remitente     :', entry.sender, '| Autorizado:', senderAuthorized);
  console.log('SMS body      :', entry.smsBody);
  console.log('Estado        :', status);
  if (!parseResult.valid) console.log('Razón rechazo :', parseResult.reason);
  if (parseResult.valid) {
    console.log('Totales       :', `declarado=${parseResult.declared} calculado=${parseResult.calculated} match=${parseResult.totalsMatch}`);
    if (parseResult.fields?.OBS) console.log('Observaciones :', parseResult.fields.OBS);
  }
  console.log('SRV           :', guardarEnSrv ? 'se guardará' : 'omitido (no autorizado o formato inválido)');
  console.log('Raw payload   :', JSON.stringify(body));
  console.log('------------------------');

  res.json({ ok: true, id: entry.id, status, srvStatus: entry.srvStatus });

  if (guardarEnSrv) {
    reenviarAOCR2(sender, smsBody, entry);
  }
});

app.get('/', (req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'index.html'));
});

app.listen(PORT, () => {
  console.log(`SMS Webhook Test corriendo en http://localhost:${PORT}`);
  console.log(`Webhook endpoint: POST http://localhost:${PORT}/webhook`);
  console.log(`Reenvío configurado a SRV-ocr2: http://${OCR2_HOST}:${OCR2_PORT}${OCR2_INBOUND_PATH}`);

  // Registrar números autorizados en SRV-ocr2 al arrancar
  setTimeout(() => {
    AUTHORIZED_SENDERS.forEach(registrarNumeroEnOCR2);
  }, 2000);
});

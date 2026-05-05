# SMS Webhook Test

Recibe SMS reenviados desde tu iPhone y los muestra en una página web.

## Instalación y arranque

```bash
cd sms-test
npm install
npm start
```

El servidor corre en http://localhost:3001

---

## Exponer al internet con ngrok

```bash
ngrok http 3001
```

Copia la URL que te da ngrok, por ejemplo:
```
https://abc123.ngrok-free.app
```

---

## Configuración en la app del iPhone

**App:** Forward SMS: Text Forwarding

| Campo | Valor |
|-------|-------|
| Webhook URL | `https://TU-URL.ngrok-free.app/webhook` |
| Method | `POST` |
| Content-Type | `application/json` |

Si la app permite body JSON personalizado:
```json
{
  "from": "{{sender}}",
  "body": "{{message}}"
}
```

---

## Prueba con curl

```bash
curl -X POST http://localhost:3001/webhook -H "Content-Type: application/json" -d "{\"from\":\"+59170000000\",\"body\":\"RRV|MESA=10101001001|P1=120|P2=95|TOTAL=215\"}"
```

---

## Endpoints

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/` | Página web con los mensajes |
| POST | `/webhook` | Recibe el SMS reenviado |
| GET | `/messages` | Devuelve todos los mensajes en JSON |
| GET | `/health` | Health check `{"status":"ok"}` |

# Integración módulo RRV (Sebastian Villaroel) → proyecto srov

## Arquitectura general — 2 bases de datos

El sistema usa **dos bases de datos distintas** con responsabilidades separadas:

| Base de datos | Motor | Responsable | Módulo | Estado |
|---|---|---|---|---|
| **MongoDB** (Replica Set 3 nodos) | MongoDB 7 | Sebastian | RRV — conteo rápido por imagen y SMS | ⚠️ Pendiente conectar |
| **PostgreSQL** (primary + standby + Pgpool) | PostgreSQL 17 | Anett / Victor | Cómputo oficial | ⏳ Pendiente implementar |

### Por qué dos bases de datos

- **MongoDB** es ideal para el módulo RRV porque recibe documentos de estructura variable (candidatos como array embebido, payloads arbitrarios de auditoría). El conteo rápido no necesita esquema rígido ni transacciones complejas.
- **PostgreSQL** es ideal para el cómputo oficial porque requiere validaciones estrictas, relaciones normalizadas (distribución → recinto → mesa → acta) y trazabilidad transaccional.

---

## Qué hizo Sebastian (commits portados)

Sebastian implementó el módulo RRV en su rama `srrv2` usando **MongoDB nativo** con 2 commits:

### Commit `00c8f2e` — Módulo RRV completo

```
internal/models/rrv_acta.go       — struct RRVActa (bson.ObjectID, []Candidato embebido)
internal/models/evento.go         — struct Evento (bson.M como payload arbitrario)
internal/repository/rrv_acta.go   — queries MongoDB (ExistsByHash, ExistsByActaID, Create, GetAll, GetByActaID)
internal/repository/evento.go     — queries MongoDB (Create, GetAll)
internal/services/ocr.go          — SimulateOCR determinístico por SHA-256
internal/handlers/rrv_handler.go  — 5 endpoints bajo /api/rrv/
```

### Commit `501d2e2` — Integración Twilio SMS

```
internal/services/twilio.go       — TwilioClient HTTP puro (SendSMS, SendConfirmation)
internal/config/config.go         — 4 variables de entorno Twilio
internal/handlers/rrv_handler.go  — endpoint webhook /api/rrv/webhook/sms agregado
docker-compose.yml                — 3 nodos MongoDB + arbiter + TWILIO_* vars en api
```

El `docker-compose.yml` de Sebastian levantaba:
- `mongo1` (primary, puerto 27017)
- `mongo2` (secondary, puerto 27018)
- `mongo-arbiter` (arbiter, puerto 27019)
- `mongo-setup` (contenedor que inicializa el Replica Set `rs0`)
- `api` conectando a `mongodb://mongo1:27017,mongo2:27018/?replicaSet=rs0`

---

## Estado actual del código (lo que se hizo y lo que falta)

### Lo que se integró en esta sesión

Se tomó la **lógica de negocio** de Sebastian y se agregó al proyecto `srov`. Los archivos creados fueron:

| Archivo | Contenido | Estado |
|---|---|---|
| `backend/handlers/rrv.go` | 6 endpoints RRV (Upload, SMS, WebhookSMS, GetAll, GetByID, GetEventos) | ✅ Creado |
| `backend/services/ocr.go` | HashBytes + SimulateOCR | ✅ Creado |
| `backend/services/twilio.go` | TwilioClient | ✅ Creado |
| `backend/models/rrv_acta.go` | Modelos RRVActa + RRVCandidato | ⚠️ Ver nota |
| `backend/models/rrv_evento.go` | Modelo RRVEvento | ⚠️ Ver nota |
| `backend/config/config.go` | Variables Twilio agregadas | ✅ Correcto |
| `backend/routes/routes.go` | Rutas /api/rrv/* registradas | ✅ Correcto |
| `.env.example` | Variables TWILIO_* | ✅ Correcto |
| `docker-compose.yml` | Variables TWILIO_* en servicio api | ✅ Correcto |

### ⚠️ Lo que está mal y debe corregirse

Los modelos `rrv_acta.go` y `rrv_evento.go` fueron escritos para **PostgreSQL/GORM**, y el handler `rrv.go` actualmente guarda en PostgreSQL. Esto es incorrecto.

**El módulo RRV de Sebastian debe guardar en MongoDB, no en PostgreSQL.**

Lo que hay que hacer en la siguiente sesión:

1. **Agregar MongoDB al `docker-compose.yml`** — incorporar los 3 nodos de Sebastian (mongo1, mongo2, mongo-arbiter, mongo-setup) al `docker-compose.yml` existente que ya tiene pg-0, pg-1, pgpool y api.

2. **Agregar conexión MongoDB al proyecto Go** — crear `backend/database/mongo.go` con la función `ConnectMongo(cfg)` usando `go.mongodb.org/mongo-driver/v2`.

3. **Agregar `MONGO_URI` y `MONGO_DB` al `config.go`** — para que la API sepa a qué MongoDB conectarse.

4. **Reemplazar los modelos GORM por structs bson** — los archivos `backend/models/rrv_acta.go` y `backend/models/rrv_evento.go` deben usar `bson.ObjectID` y `bson.M` como en Sebastian, no structs GORM.

5. **Reemplazar las queries GORM en `handlers/rrv.go`** por queries MongoDB usando el driver `go.mongodb.org/mongo-driver/v2` — o restaurar la capa `repository/` de Sebastian.

6. **Quitar `RRVActa`, `RRVCandidato`, `RRVEvento` del `AutoMigrate`** en `backend/database/database.go` — esas tablas no deben existir en PostgreSQL.

---

## Arquitectura objetivo (a implementar)

```
docker-compose.yml
│
├── pg-0 (PostgreSQL primary)  ──┐
├── pg-1 (PostgreSQL standby)  ──┤ repmgr
├── pgpool                     ──┘
│
├── mongo1 (MongoDB primary)   ──┐
├── mongo2 (MongoDB secondary) ──┤ Replica Set rs0
├── mongo-arbiter              ──┘
├── mongo-setup (init rs0, corre una vez)
│
└── api (Go/Gin)
      ├── conecta a pgpool:5432      → cómputo oficial (Anett)
      └── conecta a mongo1,mongo2   → módulo RRV (Sebastian)
```

### Flujo de datos

```
Fiscal de mesa
    │
    ├── Sube imagen/PDF ──► POST /api/rrv/actas/upload ──► MongoDB (rrv_actas)
    │
    └── Envía SMS ──────► Twilio ──► POST /api/rrv/webhook/sms ──► MongoDB (rrv_actas)
                                                                 └► SMS confirmación

Operador oficial
    └── Carga CSV ──────► POST /api/oficial/actas/csv ──► PostgreSQL (acta) [PENDIENTE]
```

---

## Colecciones MongoDB del módulo RRV

| Colección | Descripción |
|---|---|
| `rrv_actas` | Actas recibidas por imagen o SMS. Cada documento tiene candidatos como array embebido. |
| `rrv_eventos` | Audit log de cada operación (recepción, duplicado, PIN inválido, inconsistencia aritmética). |

### Estructura de un documento `rrv_actas`

```json
{
  "_id": ObjectId("..."),
  "acta_id": "ACTA-001",
  "departamento": "La Paz",
  "municipio": "La Paz",
  "recinto": "Colegio Central",
  "mesa": "001",
  "candidatos": [
    {"candidato_id": "CAND-01", "nombre": "Candidato A", "votos": 50},
    {"candidato_id": "CAND-02", "nombre": "Candidato B", "votos": 30}
  ],
  "votos_nulos": 2,
  "votos_blancos": 3,
  "total_votos": 85,
  "estado": "PROCESADA",
  "fuente": "RRV",
  "tipo_entrada": "IMAGEN",
  "hash_origen": "sha256...",
  "fecha_recepcion": ISODate("...")
}
```

---

## Endpoints RRV — referencia rápida

Todos bajo `/api/rrv/`, sin JWT (seguridad por PIN en SMS).

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/rrv/actas/upload` | Subir imagen/PDF de acta (multipart) |
| POST | `/api/rrv/sms` | Recibir acta por JSON con formato pipe |
| POST | `/api/rrv/webhook/sms` | Webhook Twilio (SMS real entrante) |
| GET | `/api/rrv/actas` | Listar todas las actas RRV |
| GET | `/api/rrv/actas/:acta_id` | Obtener acta por acta_id |
| GET | `/api/rrv/eventos` | Audit log de eventos |

### Formato SMS

```
ACTA:<id>|DEP:<departamento>|MUN:<municipio>|REC:<recinto>|MESA:<mesa>|C1:<votos>|C2:<votos>|...|NULOS:<n>|BLANCOS:<b>|TOTAL:<t>|PIN:<pin>
```

PINs válidos por defecto: `1234`, `5678`, `9012`, `2025`, `4321`
PINs extra (sin recompilar): variable de entorno `VALID_PINS=9999,8888`

### Flujo SMS real (Twilio)

```
Fiscal                    Twilio                    API
  │── SMS ───────────────►│                           │
  │  "ACTA:001|...|PIN:1234│── POST /api/rrv/webhook ─►│── valida PIN
  │                        │   Body=..., From=+591XX   │── detecta duplicado (SHA-256)
  │                        │                           │── valida aritmética
  │                        │                           │── guarda en MongoDB
  │◄── SMS confirmación ───│◄── SendConfirmation        │── registra evento
```

---

## Variables de entorno a agregar (próxima sesión)

```
# MongoDB (módulo RRV — Sebastian)
MONGO_URI=mongodb://mongo1:27017,mongo2:27018/?replicaSet=rs0
MONGO_DB=electoral_db

# Twilio (ya agregadas)
TWILIO_ACCOUNT_SID=
TWILIO_AUTH_TOKEN=
TWILIO_MESSAGING_SERVICE_SID=
TWILIO_VERIFIED_NUMBER=
```

---

## Pendientes por módulo

| Módulo | Tarea | Responsable |
|---|---|---|
| RRV | Agregar MongoDB al docker-compose.yml | Anett |
| RRV | Crear backend/database/mongo.go (ConnectMongo) | Anett |
| RRV | Agregar MONGO_URI/MONGO_DB a config.go | Anett |
| RRV | Reescribir models/rrv_acta.go y rrv_evento.go con bson | Anett |
| RRV | Reescribir handlers/rrv.go para usar MongoDB driver | Anett |
| RRV | Quitar RRVActa/RRVCandidato/RRVEvento del AutoMigrate | Anett |
| Oficial | Modelo PostgreSQL para actas oficiales | Anett |
| Oficial | Endpoint POST /api/oficial/actas/csv | Anett |
| Oficial | Validaciones (totales, recintos/mesas existentes) | Anett |
| Oficial | Auditoría oficial | Anett |
| Dashboard | Comparación RRV vs Oficial + KPIs | Pablo |

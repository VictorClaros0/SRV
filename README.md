# Sistema Nacional de Cómputo Electoral — Bolivia
## Práctica 4 · Sistemas Distribuidos

---

## Estado del proyecto

| Módulo | Responsable | Estado |
|---|---|---|
| Arquitectura C4 + contratos entre módulos | Victor | ✅ Completado |
| Modelos MongoDB + Replica Set Docker | Victor | ✅ Completado |
| Seeder CSV (distribuciones, recintos, mesas) | Victor | ✅ Completado |
| CRUD base `/api/v1/...` | Victor | ✅ Completado |
| Módulo RRV completo (imagen/PDF + SMS) | Sebastian | ✅ Completado |
| **Modelo PostgreSQL para Oficial** | **Victor / Anett** | ⏳ Pendiente |
| **Carga CSV oficial** | **Anett** | ⏳ Pendiente |
| **Validaciones oficiales** | **Anett** | ⏳ Pendiente |
| **Auditoría oficial** | **Anett** | ⏳ Pendiente |
| Comparación RRV vs Oficial | Pablo | ⏳ Pendiente |
| Endpoints de KPIs + Dashboard React | Pablo | ⏳ Pendiente |

---

## Stack tecnológico

- **Lenguaje:** Go 1.25
- **Framework HTTP:** Gin
- **Base de datos RRV:** MongoDB 7 (Replica Set 3 nodos)
- **Base de datos Oficial:** PostgreSQL ← **Anett debe configurar esto**
- **Contenedores:** Docker + Docker Compose
- **Driver Mongo:** `go.mongodb.org/mongo-driver/v2` v2.5.0
- **Módulo Go:** `srrv`

---

## Requisitos para correr el proyecto

1. **Docker Desktop** instalado y corriendo
2. **Go 1.25+** (solo si se quiere correr fuera de Docker)
3. Puerto `8080` libre en el host

---

## Cómo levantar el sistema

```bash
# Desde la raíz del proyecto
docker compose up --build

# O si la imagen ya fue compilada:
docker compose up
```

La API queda disponible en `http://localhost:8080`.

Para apagar:
```bash
docker compose down
```

---

## Poblar la base de datos (Seeder)

El seeder carga los datos iniciales desde los CSV en `internal/data/`:

```bash
# Desde la raíz del proyecto (requiere Go instalado y MongoDB corriendo)
go run cmd/seed/main.go
```

Esto limpia y recarga las colecciones:
- `distribuciones_territoriales` — desde `DistribucionTerritorial.csv`
- `recintos_electorales` — desde `RecintosElectorales.csv`
- `mesas` — desde `Mesas.csv`

---

## Variables de entorno

Las variables se configuran en `docker-compose.yml`. Las relevantes son:

| Variable | Default | Descripción |
|---|---|---|
| `MONGO_URI` | `mongodb://mongo1:27017,...` | URI del Replica Set de MongoDB |
| `DB_NAME` | `electoral` | Nombre de la base de datos Mongo |
| `PORT` | `8080` | Puerto del servidor HTTP |
| `VALID_PINS` | _(vacío)_ | PINs SMS adicionales separados por coma |

Para agregar PINs SMS extra (sin recompilar): `VALID_PINS=9999,8888`

---

## Arquitectura de carpetas

```
c:\Practica 4\
├── cmd/
│   ├── main.go              # Punto de entrada principal
│   └── seed/main.go         # Seeder de datos CSV → MongoDB
├── internal/
│   ├── config/config.go     # Carga de variables de entorno
│   ├── database/mongo.go    # Conexión y desconexión MongoDB
│   ├── models/
│   │   ├── acta.go                        # Modelo Acta (Victor)
│   │   ├── distribucion_territorial.go    # Modelo (Victor)
│   │   ├── mesa.go                        # Modelo (Victor)
│   │   ├── recinto_electoral.go           # Modelo (Victor)
│   │   ├── rrv_acta.go                    # Modelo RRVActa + Candidato (Sebastian)
│   │   └── evento.go                      # Modelo Evento/audit (Sebastian)
│   ├── repository/          # Capa de acceso a MongoDB
│   ├── handlers/            # Controladores HTTP (Gin)
│   ├── services/ocr.go      # Simulación determinística de OCR
│   └── services/twilio.go  # Cliente Twilio para confirmaciones SMS
├── internal/data/           # CSV de datos electorales
├── Dockerfile
└── docker-compose.yml
```

---

## Endpoints disponibles

### Base (Victor) — `/api/v1/`

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/api/v1/distribuciones` | Listar distribuciones territoriales |
| GET | `/api/v1/recintos` | Listar recintos electorales |
| GET | `/api/v1/mesas` | Listar mesas |
| GET | `/api/v1/actas` | Listar actas (colección general) |
| GET | `/health` | Health check |

Todos los grupos también tienen POST / PUT `/:id` / DELETE `/:id`.

### RRV (Sebastian) — `/api/rrv/`

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/rrv/actas/upload` | Subir imagen/PDF de acta (multipart) |
| POST | `/api/rrv/sms` | Recibir acta por SMS vía JSON (pruebas directas) |
| POST | `/api/rrv/webhook/sms` | **Webhook Twilio** — Twilio llama aquí cuando llega un SMS real |
| GET | `/api/rrv/actas` | Listar todas las actas RRV |
| GET | `/api/rrv/actas/:acta_id` | Obtener acta RRV por ID |
| GET | `/api/rrv/eventos` | Audit log de eventos |

#### Ejemplo: subir acta por imagen

```bash
curl -X POST http://localhost:8080/api/rrv/actas/upload \
  -F "file=@/ruta/al/acta.jpg"
```

Con datos manuales (simula OCR externo):
```bash
curl -X POST http://localhost:8080/api/rrv/actas/upload \
  -F "file=@/ruta/al/acta.jpg" \
  -F 'acta_data={"acta_id":"ACTA-001","departamento":"La Paz","municipio":"La Paz","recinto":"Colegio Central","mesa":"001","candidatos":[{"candidato_id":"CAND-01","nombre":"Candidato A","votos":50}],"votos_nulos":2,"votos_blancos":3,"total_votos":55}'
```

#### Ejemplo: recibir acta por SMS

```bash
curl -X POST http://localhost:8080/api/rrv/sms \
  -H "Content-Type: application/json" \
  -d "{\"mensaje\": \"ACTA:ACTA-SMS-001|DEP:La Paz|MUN:La Paz|REC:Colegio|MESA:001|C1:50|C2:30|NULOS:2|BLANCOS:3|TOTAL:85|PIN:1234\"}"
```

**PINs válidos:** `1234`, `5678`, `9012`, `2025`, `4321`

**Formato SMS:**
```
ACTA:<id>|DEP:<departamento>|MUN:<municipio>|REC:<recinto>|MESA:<mesa>|C1:<votos>|C2:<votos>|...|NULOS:<n>|BLANCOS:<b>|TOTAL:<t>|PIN:<pin>
```

---

## Integración Twilio (SMS real)

### Credenciales configuradas

Las credenciales ya están cargadas en `docker-compose.yml`. No hay que tocar nada para que funcionen al levantar con `docker compose up`.

| Variable | Dónde obtenerla |
|---|---|
| `TWILIO_ACCOUNT_SID` | [console.twilio.com](https://console.twilio.com) → Account Info |
| `TWILIO_AUTH_TOKEN` | [console.twilio.com](https://console.twilio.com) → Account Info |
| `TWILIO_MESSAGING_SERVICE_SID` | Console → Messaging → Services |
| `TWILIO_VERIFIED_NUMBER` | Console → Phone Numbers → Verified Caller IDs |

Las credenciales reales se guardan en el archivo `.env` (no commiteado, ver `.env.example`).
El equipo comparte las credenciales por canal privado.

> **Nota cuenta trial:** Twilio trial solo permite enviar SMS a números verificados en la consola.
> El sistema envía todas las confirmaciones al número `+59162665211`.
> En producción (cuenta pagada) se envían al número real del remitente.

### Cómo funciona el flujo SMS real

```
Fiscal de mesa                Twilio                     SRRV API
     │                           │                           │
     │── SMS al +1XXXXXXXXX ───► │                           │
     │   "ACTA:001|...|PIN:1234" │                           │
     │                           │── POST /api/rrv/webhook ─►│
     │                           │   Body=ACTA:001|...        │── valida PIN
     │                           │   From=+591XXXXXXX         │── detecta duplicados
     │                           │                            │── valida aritmética
     │                           │                            │── guarda en MongoDB
     │◄── SMS confirmación ──────│◄─ Twilio.SendConfirmation  │
     │   "✅ Acta 001 registrada" │                           │
```

### Configurar el webhook en Twilio (para SMS reales entrantes)

1. Ir a [console.twilio.com](https://console.twilio.com) → **Phone Numbers** → tu número virtual
2. En **Messaging Configuration** → **A Message Comes In** → seleccionar **Webhook**
3. URL: `http://<tu-ip-publica>:8080/api/rrv/webhook/sms`
4. Método: **HTTP POST**

Para exponer el servidor local en pruebas usar **ngrok**:
```bash
ngrok http 8080
# Copia la URL https://xxxx.ngrok.io y ponla en Twilio como webhook
```

### Enviar SMS de prueba manualmente (sin Twilio entrante)

```bash
# Simula exactamente lo que Twilio enviaría al webhook
curl -X POST http://localhost:8080/api/rrv/webhook/sms \
  -d "Body=ACTA:ACTA-TW-001|DEP:Cochabamba|MUN:Cochabamba|REC:Escuela Central|MESA:005|C1:80|C2:60|C3:20|NULOS:3|BLANCOS:2|TOTAL:165|PIN:1234" \
  -d "From=%2B59176131645" \
  -d "To=%2B19XXXXXXXX"
```

La respuesta será `<Response/>` (TwiML vacío) y el sistema enviará confirmación al número verificado.

### Enviar SMS desde Twilio hacia un número (prueba de envío)

```bash
# Reemplaza con tus credenciales del .env
curl -X POST https://api.twilio.com/2010-04-01/Accounts/$TWILIO_ACCOUNT_SID/Messages.json \
  -u "$TWILIO_ACCOUNT_SID:$TWILIO_AUTH_TOKEN" \
  --data-urlencode "To=$TWILIO_VERIFIED_NUMBER" \
  --data-urlencode "MessagingServiceSid=$TWILIO_MESSAGING_SERVICE_SID" \
  --data-urlencode "Body=Prueba SRRV: sistema funcionando correctamente"
```

---

## Colecciones MongoDB

| Colección | Descripción |
|---|---|
| `distribuciones_territoriales` | Departamento/municipio/provincia |
| `recintos_electorales` | Recintos con referencia a distribución |
| `mesas` | Mesas con referencia a recinto |
| `actas` | Actas base (Victor) |
| `rrv_actas` | Actas procesadas por el módulo RRV (Sebastian) |
| `rrv_eventos` | Audit log de todos los eventos del sistema (Sebastian) |

---

## Lo que hace el módulo RRV (Sebastian) — resumen

1. **Upload de imagen/PDF:** recibe un archivo multipart, calcula SHA-256 para detectar duplicados, simula OCR de forma determinística, valida aritmética (suma candidatos + nulos + blancos = total), guarda en `rrv_actas` con estado `PROCESADA` o `INCONSISTENTE`.

2. **Recepción SMS:** recibe JSON `{"mensaje": "..."}` en formato pipe-separado, valida PIN de seguridad, detecta duplicados por hash, valida aritmética, guarda en `rrv_actas`.

3. **Event sourcing:** cada operación (recepción, duplicado, error aritmético, PIN inválido) genera un registro en `rrv_eventos` con tipo, fuente, payload y timestamp.

4. **Idempotencia:** si el mismo archivo o mensaje se envía dos veces, retorna `409 Conflict` sin insertar duplicados.

---

## Lo que debe hacer Anett

### Tareas asignadas

1. **Crear modelo PostgreSQL para Oficial** (en coordinación con Victor)
   - Tablas normalizadas para el módulo de cómputo oficial
   - Sugerencia: crear `internal/database/postgres.go` con conexión via `database/sql` + `pgx` driver
   - Agregar variable de entorno `POSTGRES_URI` en `docker-compose.yml`

2. **Implementar carga CSV oficial**
   - Endpoint `POST /api/oficial/actas/csv` que acepte un archivo CSV
   - Parsear y validar cada fila antes de insertar
   - Retornar resumen: filas procesadas, errores por fila

3. **Implementar validaciones oficiales**
   - Validar que los totales de votos cuadren
   - Validar que los recintos/mesas existan en la base MongoDB (puede consultar `/api/v1/mesas`)
   - Marcar actas con estado `VALIDA` / `INVALIDA` / `PENDIENTE`

4. **Registrar auditoría oficial**
   - Trazabilidad completa: qué CSV se cargó, cuándo, qué filas fallaron
   - Sugerencia: tabla `oficial_auditoria` en PostgreSQL o colección `oficial_eventos` en MongoDB (siguiendo el patrón de `rrv_eventos`)

### Cómo agregar el módulo Oficial

Seguir el mismo patrón que el módulo RRV de Sebastian:

```
internal/
├── models/oficial_acta.go          # Nuevo struct para actas oficiales
├── repository/oficial_acta.go      # Acceso a PostgreSQL (o Mongo)
├── handlers/oficial_handler.go     # Controladores HTTP
└── services/csv_parser.go          # Lógica de parseo CSV
```

Registrar las rutas en `cmd/main.go`:
```go
oficial := r.Group("/api/oficial")
{
    oficial.POST("/actas/csv",    oficialH.UploadCSV)
    oficial.GET("/actas",         oficialH.GetAll)
    oficial.GET("/auditoria",     oficialH.GetAuditoria)
}
```

### Patrón de repositorio a seguir

Ver `internal/repository/rrv_acta.go` como referencia. Estructura básica:
```go
type OficialActaRepository struct {
    collection *mongo.Collection // o *sql.DB si usan PostgreSQL
}

func NewOficialActaRepository(db *mongo.Database) *OficialActaRepository {
    return &OficialActaRepository{collection: db.Collection("oficial_actas")}
}
```

---

## Contacto del equipo

| Integrante | Módulo | Rama |
|---|---|---|
| Victor | Arquitectura + MongoDB base | `main` / `srrv` |
| Sebastian | Módulo RRV | `srrv` → `anett` |
| Anett | Módulo Oficial | branch propio desde `anett` |
| Pablo | Dashboard + KPIs | branch propio |

# Sistema de Registro de Votos (SROV)

Backend REST en **Go (Gin + GORM + JWT)** para el registro y transcripción de actas electorales. La base de datos es **PostgreSQL** en alta disponibilidad con repmgr y Pgpool-II. La transcripción masiva se simula mediante un workflow de **n8n** que procesa los datos del Excel como si fuera un operador humano.

## ¿Cómo funciona el sistema?

SROV es una plataforma electoral que digitaliza el proceso de transcripción de actas de votación. El frontend en React se comunica con una API REST en Go, que persiste los datos en un clúster PostgreSQL de alta disponibilidad compuesto por un nodo primario y un standby sincronizados mediante repmgr, y balanceados a través de Pgpool-II. Las actas se cargan inicialmente desde archivos CSV con estado `impresa` y votos en cero; al activar la simulación, n8n actúa como operador automatizado leyendo los resultados reales y enviándolos al webhook de la API acta por acta. Según el contenido de cada acta, el sistema la marca como `transcrita` (válida para el conteo oficial) u `observada` (derivada a auditoría por irregularidades), generando así un registro completo y trazable del proceso electoral.

---

## Cómo funciona la app

SROV simula el proceso de transcripción de actas electorales. En una elección real, operadores humanos ingresan los datos de cada acta impresa al sistema. Aquí ese proceso lo automatiza **n8n**:

1. Las actas se cargan desde `ActasImpresas.csv` con estado `impresa` y votos en cero.
2. Al presionar **"▶ Simular transcripción (n8n)"**, el frontend dispara un workflow en n8n.
3. n8n lee `Transcripciones.csv` (los resultados reales) y envía cada fila al backend.
4. El backend actualiza cada acta:
   - Sin observaciones → estado `transcrita` (cuenta en resultados oficiales)
   - Con observaciones → estado `observada` (excluida del conteo oficial, va a auditoría)

---

## Cómo ejecutar el proyecto (paso a paso)

### Requisitos
- Docker y Docker Compose instalados

### 1. Clonar el repositorio y preparar variables de entorno

```bash
cp .env.example .env
```

El archivo `.env.example` ya tiene valores funcionales para desarrollo. No es necesario cambiarlo para pruebas locales.

### 2. Levantar todos los servicios

```bash
docker-compose up -d --build
```

Esperar ~30 segundos a que PostgreSQL esté listo. Se puede verificar con:

```bash
docker-compose ps
```

Todos los servicios deben estar en estado `Up`.

### 3. Puertos disponibles

| Servicio   | URL                          | Descripción                     |
|------------|------------------------------|---------------------------------|
| Frontend   | http://localhost:3010        | Interfaz web principal          |
| API        | http://localhost:8090        | REST API (Go)                   |
| n8n        | http://localhost:5678        | Editor de workflows             |
| PostgreSQL | localhost:5434               | Acceso directo a la base de datos |

### 4. Primer acceso

Abrir **http://localhost:3010** e iniciar sesión con:
- **Usuario:** `admin`
- **Contraseña:** `admin123`

Al arrancar por primera vez el sistema carga automáticamente los datos de los CSVs de la carpeta `backend/data/`. Si la tabla ya tiene datos, el seed se omite.

---

## Guía de n8n

### ¿Qué es n8n aquí?

n8n es la herramienta que simula a un operador humano transcribiendo actas. Cuando se presiona el botón "Simular transcripción" en la app, el frontend activa un workflow de n8n que lee todos los datos del CSV y los envía al backend uno por uno.

### Paso 1 — Abrir n8n

Ir a **http://localhost:5678** e ingresar con:
- **Usuario:** `admin`
- **Contraseña:** `admin123`

### Paso 2 — Importar el workflow

En la primera vez que se abre n8n, el workflow no existe. Importarlo desde el archivo incluido en el proyecto:

1. En el menú lateral izquierdo, ir a **Workflows**
2. Hacer click en los tres puntos `...` o en el botón de importar
3. Seleccionar **"Import from file"**
4. Elegir el archivo **`n8n-workflow.json`** que está en la raíz del proyecto
5. El workflow aparece en el editor con 4 nodos conectados

### Paso 3 — Publicar el workflow

Con el workflow abierto, hacer click en el botón **"Publish"** (arriba a la derecha). Esto registra el webhook y lo deja listo para recibir llamadas.

> Si el botón dice "Publish" con un número como "0/1", significa que hay cambios sin publicar — hacer click para publicar.

### Paso 4 — Simular la transcripción

1. Ir al frontend: **http://localhost:3000**
2. Ir a la sección **Actas**
3. Hacer click en **"▶ Simular transcripción (n8n)"**
4. El botón cambia a "⏳ Simulando..."
5. Esperar a que el mensaje de confirmación aparezca con el conteo final

### Paso 5 — Verificar resultados

Después de la simulación, en la sección Actas se deben ver:
- **Transcritas:** ~5305 (actas sin observaciones)
- **Observadas:** 51 (actas con algún problema registrado)

### ¿Qué pasa si el workflow no existía y ya hay datos transcritos?

Si las actas ya están en estado `transcrita` pero sin las observaciones correctas, existe un endpoint para reprocesar todo directamente sin n8n:

```bash
curl -X POST http://localhost:8090/api/v1/actas/procesar-transcripciones
```

Respuesta esperada:
```json
{"actualizadas": 5396, "errores": 0, "observadas": 51}
```

---

## Arquitectura

```
Browser (puerto 3000)
  │
  ▼
Frontend React + Nginx
  ├── /api/*          → proxy → API Go (puerto 8090)
  ├── /webhook/*      → proxy → API Go (puerto 8090)
  └── /n8n/*          → proxy → n8n (puerto 5678)
                                  │
                         Workflow de transcripción
                                  │
                          GET /api/v1/actas/para-transcribir
                          POST /webhook/n8n/transcripcion (×N)
                                  │
                              PostgreSQL HA
                           pg-0 (primary) ← pg-1 (standby)
                                  │
                              pgpool (5434)
```

| Servicio   | Puerto host | Descripción                         |
|-----------|-------------|-------------------------------------|
| frontend  | 3010        | React + Nginx (proxy hacia api/n8n) |
| api       | 8090        | Go REST API                         |
| pgpool    | 5434        | PostgreSQL HA (punto de entrada)    |
| n8n       | 5678        | Automatización de transcripción     |

---

## Flujo de transcripción con n8n (Simulación)

### Qué hace la simulación

Simula el proceso en el que un operador humano transcribe el contenido de cada acta impresa al sistema. En lugar de un humano, **n8n automatiza el proceso** leyendo `Transcripciones.csv` y enviando cada fila al webhook de la API.

```
Acta "impresa" (votos en cero)
        │
        ▼
n8n lee Transcripciones.csv
        │
  Para cada fila:
        ▼
POST /webhook/n8n/transcripcion
        │
        ▼
Acta → "transcrita" (con votos)
   o → "observada" (si tiene observaciones en el Excel)
```

### Configurar el workflow en n8n

1. Abrir **http://localhost:5678**
2. Ingresar con las credenciales `N8N_USER` / `N8N_PASSWORD` del `.env`
3. Crear un nuevo workflow con estos nodos en orden:

#### Nodo 1 — Trigger (Webhook)
- Tipo: **Webhook**
- HTTP Method: `POST`
- Path: `trigger-transcripcion`
- Response Mode: `Immediately`
- Activar el nodo (ícono de rayo)

#### Nodo 2 — Obtener datos del CSV
- Tipo: **HTTP Request**
- Method: `GET`
- URL: `http://api:8080/api/v1/actas/para-transcribir`
- Response Format: `JSON`

#### Nodo 3 — Iterar por acta
- Tipo: **Split Out**
- Field To Split Out: dejar en blanco (itera el array raíz)

#### Nodo 4 — Transcribir cada acta
- Tipo: **HTTP Request**
- Method: `POST`
- URL: `http://api:8080/webhook/n8n/transcripcion`
- Body: `JSON`
- Body Parameters: `{{ $json }}` (pasa todos los campos del item actual)

#### Nodo 5 — (Opcional) Espera entre actas
- Tipo: **Wait**
- Resume: `After Time Interval`
- Amount: `500` milliseconds
> Agrega este nodo entre el 3 y el 4 para simular que un humano tarda en transcribir.

4. **Guardar y activar** el workflow (toggle en la esquina superior derecha)

### Iniciar la simulación desde el frontend

1. Abrir **http://localhost:3000**
2. Iniciar sesión con `admin` / contraseña del `.env`
3. Ir a **Actas**
4. Hacer clic en **"▶ Simular transcripción (n8n)"**
5. El frontend llama al trigger de n8n, que procesa acta por acta
6. La barra de progreso se actualiza en tiempo real (polling cada 2 segundos)
7. Las actas cambian de estado: `Impresa → Transcrita` o `Impresa → Observada`

---

## Endpoints de la API

### Autenticación

| Método | Ruta                      | Auth | Descripción                          |
|--------|---------------------------|------|--------------------------------------|
| POST   | `/api/v1/auth/login`      | —    | Login. Devuelve JWT                  |
| GET    | `/api/v1/auth/me`         | JWT  | Datos del usuario autenticado        |
| POST   | `/api/v1/auth/register`   | Admin| Registrar nuevo usuario              |

### Actas

| Método | Ruta                              | Auth   | Descripción                              |
|--------|-----------------------------------|--------|------------------------------------------|
| GET    | `/api/v1/actas`                   | JWT    | Listar actas (con búsqueda `?search=`)   |
| GET    | `/api/v1/actas/:id`               | JWT    | Obtener acta por ID                      |
| POST   | `/api/v1/actas`                   | JWT    | Crear acta manualmente                   |
| PUT    | `/api/v1/actas/:id`               | JWT    | Actualizar acta                          |
| DELETE | `/api/v1/actas/:id`               | JWT    | Borrado lógico de acta                   |
| GET    | `/api/v1/actas/resumen`           | JWT    | Totales por código de recinto            |
| GET    | `/api/v1/actas/para-transcribir`  | —      | Datos del CSV para n8n (público)         |
| POST   | `/webhook/n8n/transcripcion`      | —      | Webhook que n8n llama para transcribir   |

### Resultados y Auditoría

| Método | Ruta                  | Auth | Descripción                                              |
|--------|-----------------------|------|----------------------------------------------------------|
| GET    | `/api/v1/resultados`  | JWT  | Resultados oficiales: resumen, comparativa, por depto.   |
| GET    | `/api/v1/auditoria`   | JWT  | Actas con observaciones excluidas del conteo oficial     |

### Catálogos (recintos, mesas, distribución, usuarios)

| Método | Ruta                        | Auth   | Descripción                   |
|--------|-----------------------------|--------|-------------------------------|
| GET    | `/api/v1/recintos`          | JWT    | Listar recintos               |
| GET    | `/api/v1/distribuciones`    | JWT    | Listar distribuciones         |
| GET    | `/api/v1/mesas`             | JWT    | Listar mesas                  |
| GET    | `/api/v1/usuarios`          | Admin  | Listar usuarios               |
| GET    | `/health`                   | —      | Estado de la API y BD         |

---

## Respuestas de los endpoints nuevos

### GET /api/v1/resultados

```json
{
  "resumen": {
    "totalActas": 150,
    "actasTranscritas": 140,
    "actasObservadas": 5,
    "actasImpresa": 5
  },
  "comparativa": [
    { "candidato": "P1", "votos": 52340 },
    { "candidato": "P2", "votos": 48210 },
    { "candidato": "P3", "votos": 31500 },
    { "candidato": "P4", "votos": 27800 }
  ],
  "porDepartamento": [
    {
      "ubicacion": "Chuquisaca",
      "p1": 8200, "p2": 7100, "p3": 5300, "p4": 4100,
      "votosValidos": 24700, "votosNulos": 800, "votosBlanco": 300,
      "totalActas": 35
    }
  ]
}
```

### GET /api/v1/auditoria

```json
{
  "total": 5,
  "actas": [
    {
      "id": 23,
      "codigoActa": 1010200001023,
      "codigoRecinto": 1010200001,
      "nroMesa": 23,
      "estado": "observada",
      "observaciones": "El acta presenta tachaduras en el campo P2",
      "fechaCreacion": "2024-11-10T14:32:00Z"
    }
  ]
}
```

### POST /webhook/n8n/transcripcion

**Request:**
```json
{
  "codigoActa": 1010200001001,
  "p1": 140, "p2": 39, "p3": 124, "p4": 345,
  "votosValidos": 648, "votosNulos": 64, "votosBlanco": 76,
  "papeletasAnfora": 788, "papeletasNoUsadas": 89,
  "observaciones": "",
  "aperturaHora": 8, "aperturaMinutos": 0,
  "cierreHora": 16, "cierreMinutos": 30
}
```

**Response (estado = "transcrita" si observaciones vacío):**
```json
{ "ok": true, "id": 1, "estado": "transcrita" }
```

**Response (estado = "observada" si tiene observaciones):**
```json
{ "ok": true, "id": 23, "estado": "observada" }
```

---

## Cómo hacer pruebas

### 1. Login (obtener token)

```bash
TOKEN=$(curl -s -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"usuario":"admin","contrasena":"admin123"}' \
  | jq -r '.token')
echo "Token: $TOKEN"
```

### 2. Ver actas

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8090/api/v1/actas | jq '.[0:3]'
```

### 3. Disparar simulación de n8n manualmente

```bash
curl -s -X POST http://localhost:5678/webhook/trigger-transcripcion \
  -H "Content-Type: application/json" -d '{}'
```

### 4. Reprocesar transcripciones directamente (sin n8n)

```bash
curl -s -X POST http://localhost:8090/api/v1/actas/procesar-transcripciones
```

### 5. Ver resultados oficiales

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8090/api/v1/resultados | jq .
```

### 6. Ver actas con observaciones (auditoría)

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8090/api/v1/auditoria | jq .
```

### 7. Transcribir un acta individual (prueba directa del webhook)

```bash
curl -s -X POST http://localhost:8090/webhook/n8n/transcripcion \
  -H "Content-Type: application/json" \
  -d '{
    "codigoActa": 1010200001001,
    "p1": 140, "p2": 39, "p3": 124, "p4": 345,
    "votosValidos": 648, "votosNulos": 64, "votosBlanco": 76,
    "papeletasAnfora": 788, "papeletasNoUsadas": 89,
    "observaciones": "",
    "aperturaHora": 8, "aperturaMinutos": 0,
    "cierreHora": 16, "cierreMinutos": 30
  }' | jq .
```

### 8. Prueba de failover de PostgreSQL

```bash
# 1. Detener el primary
docker-compose stop pg-0

# 2. Esperar ~15 segundos para que repmgr promueva el standby
sleep 15

# 3. La API debe seguir respondiendo
curl http://localhost:8090/health
# {"database":"up","status":"ok"}

# 4. Volver a levantar pg-0
docker-compose start pg-0
```

---

## Campos de auditoría

El sistema captura automáticamente los siguientes campos de auditoría en cada acta **sin modificar la estructura de la base de datos**:

| Campo del prompt       | Campo en BD         | Descripción                                          |
|------------------------|---------------------|------------------------------------------------------|
| `usuario_transcriptor` | `creado_por`        | ID del usuario que inició la transcripción (admin)   |
| `fecha_transcripcion`  | `fecha_creacion`    | Timestamp automático al crear el registro            |
| `estado_validacion`    | `estado`            | `impresa` / `transcrita` / `observada`               |
| `observaciones`        | `observaciones`     | Texto libre del Excel; si no está vacío → "observada"|

---

## Endpoints de resultados y filtros

Implementados en `handlers/resultados.go` y `handlers/filtros.go`. Todos requieren JWT.

### Resultados por recinto

```http
GET /api/v1/resultados/recintos
```

Devuelve los votos de cada candidato agrupados por recinto electoral.

```sql
SELECT re.recinto AS ubicacion,
  COALESCE(SUM(a.p1),0) AS p1, COALESCE(SUM(a.p2),0) AS p2,
  COALESCE(SUM(a.p3),0) AS p3, COALESCE(SUM(a.p4),0) AS p4,
  COUNT(*) AS total_actas
FROM acta a
JOIN recinto_electoral re ON a.codigo_recinto = re.recinto_id
WHERE a.fecha_eliminado IS NULL AND a.estado = 'transcrita'
GROUP BY re.recinto_id, re.recinto ORDER BY re.recinto;
```

**Respuesta:**
```json
[
  {
    "ubicacion": "Colegio Nacional Bolívar",
    "p1": 342,
    "p2": 198,
    "p3": 87,
    "p4": 210,
    "votosValidos": 0,
    "votosNulos": 0,
    "votosBlanco": 0,
    "totalActas": 5
  },
  {
    "ubicacion": "Unidad Educativa San Martín",
    "p1": 510,
    "p2": 320,
    "p3": 145,
    "p4": 290,
    "votosValidos": 0,
    "votosNulos": 0,
    "votosBlanco": 0,
    "totalActas": 8
  }
]
```

### Resultados por provincia

```http
GET /api/v1/resultados/provincias
```

Devuelve los votos de cada candidato agrupados por provincia.

```sql
SELECT dt.provincia AS ubicacion,
  COALESCE(SUM(a.p1),0) AS p1, COALESCE(SUM(a.p2),0) AS p2,
  COALESCE(SUM(a.p3),0) AS p3, COALESCE(SUM(a.p4),0) AS p4,
  COUNT(*) AS total_actas
FROM acta a
JOIN recinto_electoral re ON a.codigo_recinto = re.recinto_id
JOIN distribucion_territorial dt ON re.id_distribucion_territorial = dt.id
WHERE a.fecha_eliminado IS NULL AND a.estado = 'transcrita'
GROUP BY dt.provincia ORDER BY dt.provincia;
```

**Respuesta:**
```json
[
  {
    "ubicacion": "Aroma",
    "p1": 1540,
    "p2": 890,
    "p3": 320,
    "p4": 760,
    "votosValidos": 0,
    "votosNulos": 0,
    "votosBlanco": 0,
    "totalActas": 22
  },
  {
    "ubicacion": "Murillo",
    "p1": 4210,
    "p2": 3100,
    "p3": 980,
    "p4": 2340,
    "votosValidos": 0,
    "votosNulos": 0,
    "votosBlanco": 0,
    "totalActas": 87
  }
]
```

### Heatmap por departamento

```http
GET /api/v1/heatmap/departamentos
```

Devuelve el candidato ganador y sus votos totales por departamento.

```sql
SELECT dt.departamento AS ubicacion,
  CASE
    WHEN SUM(p1) >= SUM(p2) AND SUM(p1) >= SUM(p3) AND SUM(p1) >= SUM(p4) THEN 'P1'
    WHEN SUM(p2) >= SUM(p3) AND SUM(p2) >= SUM(p4) THEN 'P2'
    WHEN SUM(p3) >= SUM(p4) THEN 'P3'
    ELSE 'P4'
  END AS candidato_ganador,
  GREATEST(SUM(p1), SUM(p2), SUM(p3), SUM(p4)) AS votos
FROM acta a
JOIN recinto_electoral re ON a.codigo_recinto = re.recinto_id
JOIN distribucion_territorial dt ON re.id_distribucion_territorial = dt.id
WHERE a.fecha_eliminado IS NULL AND a.estado = 'transcrita'
GROUP BY dt.departamento ORDER BY dt.departamento;
```

**Respuesta:**
```json
[
  { "ubicacion": "Beni",            "candidatoGanador": "P2", "votos": 5420 },
  { "ubicacion": "Cochabamba",      "candidatoGanador": "P1", "votos": 9800 },
  { "ubicacion": "La Paz",          "candidatoGanador": "P1", "votos": 12340 },
  { "ubicacion": "Oruro",           "candidatoGanador": "P3", "votos": 3100 },
  { "ubicacion": "Pando",           "candidatoGanador": "P4", "votos": 1870 },
  { "ubicacion": "Potosí",          "candidatoGanador": "P1", "votos": 7650 },
  { "ubicacion": "Santa Cruz",      "candidatoGanador": "P2", "votos": 11200 },
  { "ubicacion": "Tarija",          "candidatoGanador": "P3", "votos": 4300 },
  { "ubicacion": "Chuquisaca",      "candidatoGanador": "P1", "votos": 5980 }
]
```

---

## Endpoints de filtros

Devuelven solo id y nombre para poblar selectores en el frontend. Implementados en `handlers/filtros.go`.

### Filtro de recintos

```http
GET /api/v1/filtros/recintos
```

```json
[
  { "id": 1, "nombre": "Colegio Nacional Bolívar" },
  { "id": 2, "nombre": "Unidad Educativa San Martín" }
]
```

### Filtro de mesas

```http
GET /api/v1/filtros/mesas
```

```json
[
  { "id": 1, "nombre": "M-001" },
  { "id": 2, "nombre": "M-002" }
]
```

### Filtro de provincias

```http
GET /api/v1/filtros/provincias
```

```json
[
  { "nombre": "Aroma" },
  { "nombre": "Murillo" },
  { "nombre": "Sud Yungas" }
]
```

### Filtro de departamentos

```http
GET /api/v1/filtros/departamentos
```

```json
[
  { "nombre": "Beni" },
  { "nombre": "Cochabamba" },
  { "nombre": "La Paz" },
  { "nombre": "Oruro" },
  { "nombre": "Pando" },
  { "nombre": "Potosí" },
  { "nombre": "Santa Cruz" },
  { "nombre": "Tarija" },
  { "nombre": "Chuquisaca" }
]
```

---

## Alternativas a n8n para la transcripción

| Herramienta         | Pros                                              | Contras                               |
|---------------------|---------------------------------------------------|---------------------------------------|
| **n8n** (actual)    | Visual, sin código, fácil de depurar              | Requiere configurar el workflow a mano|
| **Script Python**   | Rápido de escribir con pandas, ideal para ETL     | No simula el flujo "humano" en el tiempo |
| **Apache Airflow**  | Robusto para pipelines de datos complejos         | Pesado, mayor curva de aprendizaje    |
| **Temporal.io**     | Workflows resilientes con retry automático        | Más moderno pero menos documentación  |
| **Celery + Redis**  | Colas de tareas, fácil integración con Python/Go  | Requiere dos servicios extra          |

Para este proyecto académico, **n8n es la opción óptima** porque permite visualizar el flujo, simular delays (nodo Wait) y no requiere escribir código extra.

---

## Desarrollo local (sin Docker para la API)

```bash
# 1. Levantar solo la BD
docker compose up -d pg-0 pg-1 pgpool

# 2. Ejecutar la API local
cd backend
export DB_HOST=localhost
export DB_PORT=5434
export DB_USER=votos
export DB_PASSWORD=votos_db_secret
export DB_NAME=votos
export JWT_SECRET=dev-secret
export ADMIN_USER=admin
export ADMIN_PASSWORD=admin123
go run .
```

---

## Licencia

Proyecto académico.

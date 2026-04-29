# README — Integración Cómputo Oficial + n8n

## Resumen de cambios

### Problema detectado
Los CSVs originales (`Mesas.csv`, `RecintosElectorales.csv`, `DistribucionTerritorial.csv`) usaban un sistema de codificación de 11 dígitos para las mesas, incompatible con el Excel de referencia (`Recursos Practica 4.xlsx`) que usa códigos de 13 dígitos (`CodigoActa`).

**Ejemplo del conflicto:**

| Fuente | Código de mesa | Formato |
|--------|---------------|---------|
| CSV original | `10101001001` | 11 dígitos |
| Excel (CodigoActa) | `1010200001001` | 13 dígitos |

La relación en el Excel es:
```
CodigoActa (13) = CodigoRecinto (10) + NroMesa (3 dígitos)
Ejemplo: 1010200001 + 001 = 1010200001001
```

### Solución — Re-sembrado de PostgreSQL desde el Excel

Se regeneraron los 3 CSVs de catálogo usando los datos reales de `Recursos Practica 4.xlsx`:

| Archivo | Hoja Excel usada | Registros |
|---------|-----------------|-----------|
| `backend/data/DistribucionTerritorial.csv` | `DistribucionTerritorial` | 340 territorios |
| `backend/data/RecintosElectorales.csv` | `RecintosElectorales` | 537 recintos |
| `backend/data/Mesas.csv` | `ActasImpresas` | 5396 mesas |

Luego se truncaron las tablas y se rebuildeó la imagen Docker:

```bash
# Limpiar tablas
docker exec srv-pg-0-1 env PGPASSWORD="..." psql -U votos_user -d votos_db -c "
TRUNCATE oficial_evento, oficial_error, oficial_auditoria, acta,
         mesa, recinto_electoral, distribucion_territorial RESTART IDENTITY CASCADE;
"

# Rebuild con los nuevos CSVs
docker-compose build api
docker-compose up -d --no-deps api
```

Al arrancar, la API siembra automáticamente:
```
seed distribuciones: 340 registros insertados
seed recintos:       537 registros insertados
seed mesas:         5396 registros insertados
```

---

## n8n — Cómputo Oficial mediante workflow

### ¿Qué es n8n en este proyecto?

n8n es el orquestador que simula la carga oficial del CSV de transcripciones electorales. En producción este paso lo haría el Tribunal Supremo Electoral; aquí lo automatiza n8n.

### Acceso

```
URL:  http://localhost:5678
```

### Workflow: "Cómputo Oficial — Carga CSV"

El workflow tiene 4 nodos en secuencia:

```
Inicio Manual → Login API → Leer CSV Oficial → Cargar CSV Oficial
```

| Nodo | Tipo | Descripción |
|------|------|-------------|
| Inicio Manual | Manual Trigger | Disparo manual desde la UI |
| Login API | HTTP Request POST | `POST /api/v1/auth/login` → obtiene JWT |
| Leer CSV Oficial | Read Binary File | Lee `/home/node/.n8n-files/oficial_seed.csv` |
| Cargar CSV Oficial | HTTP Request POST | `POST /api/oficial/csv/upload` con Authorization Bearer |

### Importar el workflow

1. Abre http://localhost:5678
2. Ve a **Workflows** → botón `+` → **Import from File**
3. Selecciona `/tmp/workflow_oficial.json`
4. Haz clic en **Save**

### Ejecutar el workflow

1. Abre el workflow importado
2. Haz clic en **Execute Workflow** ▶
3. Verás los 4 nodos ejecutándose en verde
4. El nodo final devuelve el resumen de auditoría

### Archivo de datos

El CSV cargado por n8n es:

```
backend/data/oficial_seed.csv
```

- **5396 filas** — tomadas de la hoja `Transcripciones` del Excel
- **51 actas con observación** — columna `Observaciones` del Excel, texto real
- Las observaciones vienen directamente del Excel (no son sintéticas)

---

## Fuente de datos: oficial_seed.csv

Generado desde la hoja `Transcripciones` de `Recursos Practica 4.xlsx`:

| Columna CSV | Columna Excel | Descripción |
|-------------|---------------|-------------|
| `acta_id` | `CodigoActa` | ID único del acta |
| `departamento` | `Departamento` | Departamento |
| `provincia` | `Provincia` | Provincia |
| `municipio` | `Municipio` | Municipio |
| `recinto` | `RecintoNombre` | Nombre del recinto |
| `mesa` | `CodigoActa` | Código de mesa (13 dígitos) |
| `candidato_1..4` | `P1..P4` | Votos por candidato |
| `votos_nulos` | `VotosNulos` | Votos nulos |
| `votos_blancos` | `VotosBlancos` | Votos en blanco |
| `total_votos` | `PapeletasAnfora` | Total papeletas en ánfora |
| `ciudadanos_habilitados` | `VotantesHabilitados` | Padrón |
| `observacion` | `Observaciones` | Motivo OBSERVADA (Ley N° 026) |

---

## Resultado de la carga (n8n → PostgreSQL)

```
Filas procesadas : 5396
Filas válidas    : 5125  ← actas cargadas en PostgreSQL
Filas rechazadas : 271   ← OBSERVADAS (Ley N° 026 del Régimen Electoral)
Estado           : PROCESADO_CON_ERRORES
```

Las 271 actas rechazadas tienen observación en la columna `Observaciones` del Excel. Según la **Ley N° 026 del Régimen Electoral**, las actas con irregularidades formales se marcan como `OBSERVADA` y **no se computan** en el resultado oficial.

---

## Endpoints para Pablo (dashboard)

### GET /api/rrv/actas — Conteo rápido (MongoDB)

```
GET http://localhost:8080/api/rrv/actas
GET http://localhost:8080/api/rrv/actas?limit=100
GET http://localhost:8080/api/rrv/actas?limit=100&offset=200
```

No requiere autenticación.

**Respuesta:**
```json
[
  {
    "id": "69f1d83c8414f089276e84c6",
    "acta_id": "ACTA-1010200001001",
    "departamento": "Chuquisaca",
    "municipio": "Yotala",
    "recinto": "U.E. Padresama",
    "mesa": "1010200001001",
    "candidatos": [
      { "candidato_id": "CAND-01", "nombre": "Daenerys Targaryen", "votos": 140 },
      { "candidato_id": "CAND-02", "nombre": "Sansa Stark",        "votos": 39  },
      { "candidato_id": "CAND-03", "nombre": "Robert Baratheon",   "votos": 124 },
      { "candidato_id": "CAND-04", "nombre": "Tyrion Lannister",   "votos": 345 }
    ],
    "votos_nulos": 64,
    "votos_blancos": 76,
    "total_votos": 788,
    "estado": "PROCESADA",
    "fuente": "RRV",
    "tipo_entrada": "IMAGEN",
    "fecha_recepcion": "2026-04-29T10:06:52.178Z"
  }
]
```

**Total de actas:** 397 (391 PROCESADA + 6 OBSERVADA)

---

### GET /api/oficial/actas — Cómputo oficial (PostgreSQL)

Requiere `Authorization: Bearer <token>`.

```
GET http://localhost:8080/api/oficial/actas
GET http://localhost:8080/api/oficial/actas?limit=100
GET http://localhost:8080/api/oficial/actas?limit=100&offset=200
```

**Respuesta:**
```json
[
  {
    "acta_id": "ACTA-1010200001001",
    "departamento": "Chuquisaca",
    "provincia": "Yotala",
    "municipio": "Oropeza",
    "recinto": "U.E. Padresama",
    "mesa": "1010200001001",
    "candidatos": [
      { "candidato_id": "CAND-01", "nombre": "Candidato 1", "votos": 140 },
      { "candidato_id": "CAND-02", "nombre": "Candidato 2", "votos": 39  },
      { "candidato_id": "CAND-03", "nombre": "Candidato 3", "votos": 124 },
      { "candidato_id": "CAND-04", "nombre": "Candidato 4", "votos": 345 }
    ],
    "votos_nulos": 64,
    "votos_blancos": 76,
    "total_votos": 788,
    "estado": "VALIDADA",
    "fuente": "OFICIAL"
  }
]
```

**Total de actas:** 5125

---

### GET /api/oficial/actas/:acta_id — Acta individual

```
GET http://localhost:8080/api/oficial/actas/ACTA-1010200001001
Authorization: Bearer <token>
```

---

## Endpoints de auditoría (admin)

Requieren `Authorization: Bearer <token>` con rol administrador.

### GET /api/oficial/auditoria — Resumen de cargas

```
GET http://localhost:8080/api/oficial/auditoria
```

**Respuesta:**
```json
[
  {
    "id": 1,
    "nombre_archivo": "oficial_seed.csv",
    "usuario": "admin",
    "fecha_carga": "2026-04-29T13:59:54.882363Z",
    "filas_procesadas": 5396,
    "filas_validas": 5125,
    "filas_rechazadas": 271,
    "estado": "PROCESADO_CON_ERRORES"
  }
]
```

---

### GET /api/oficial/errores — Actas rechazadas (OBSERVADAS)

```
GET http://localhost:8080/api/oficial/errores
GET http://localhost:8080/api/oficial/errores?limit=50
```

**Respuesta:**
```json
[
  {
    "id": 1,
    "auditoria_id": 1,
    "fila": 6,
    "acta_id": "ACTA-1010200001005",
    "error": "Acta OBSERVADA (Ley N° 026 del Régimen Electoral): Falta de datos de apertura o cierre: Omisión de la hora exacta de inicio o finalización de la jornada electoral en el acta.",
    "created_at": "2026-04-29T13:59:54.938489Z"
  }
]
```

**Total de errores:** 271 (todas son actas OBSERVADAS por Ley N° 026)

---

### GET /api/oficial/eventos — Eventos del proceso

```
GET http://localhost:8080/api/oficial/eventos
```

---

## Obtener token JWT

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"usuario":"admin","contrasena":"admin123"}'
```

Respuesta:
```json
{ "token": "eyJhbGci..." }
```

Usar en headers:
```
Authorization: Bearer eyJhbGci...
```

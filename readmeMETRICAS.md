# Dashboard de Métricas Electorales

Documentación de los endpoints del dashboard OCR para el frontend de Pablo.

---

## Cómo ejecutar el proyecto

### Requisitos
- Docker y Docker Compose instalados
- Archivos PDF de actas en `./data/logs/PdfOutput/`

### Pasos

```bash
# 1. Levantar todos los servicios
docker compose up --build

# Si da error de contenedor duplicado, limpiar primero:
sudo docker rm -f electoral_api electoral_mongo1 electoral_mongo2 electoral_mongo_arbiter electoral_mongo_setup electoral_ocr_api electoral_ocr_worker electoral_redis
docker compose up --build
```

Los servicios que levanta:

| Servicio | Puerto | Descripción |
|---|---|---|
| API Go (métricas) | `8080` | Backend principal con todos los endpoints |
| OCR API (FastAPI) | `8000` | Servicio de procesamiento de imágenes |
| MongoDB Replica Set | `27017` | Base de datos principal |
| Redis | `6379` | Broker de tareas Celery |

---

## PASO OBLIGATORIO: Lanzar el procesamiento OCR

El OCR **no es automático**. Después de levantar los servicios, hay que disparar el procesamiento manualmente:

```
POST http://localhost:8000/api/ocr/scan-all
```

Sin body. Devuelve:
```json
{ "enqueued_files": 597 }
```

Esto encola todos los PDFs en `./data/logs/PdfOutput/` para que el worker de Celery los procese con PaddleOCR y guarde los resultados en MongoDB. El procesamiento puede tardar varios minutos dependiendo de la cantidad de archivos.

---

## Endpoints del Dashboard

Base URL: `http://localhost:8080/api/v1/dashboard`

### KPIs Generales
```
GET /dashboard/kpis
```
Devuelve totales globales: actas procesadas, votos totales, mesas registradas, actas pendientes/completadas/con error.

---

### Votos por Candidato
```
GET /dashboard/votos-candidato
```
Devuelve votos acumulados para cada candidato (P1–P4). Usar para gráfico de barras o torta.

Candidatos:
- `p1` → Daenerys Targaryen
- `p2` → Sansa Stark
- `p3` → Robert Baratheon
- `p4` → Tyrion Lannister

---

### Participación Electoral
```
GET /dashboard/participacion
```
Devuelve total de habilitados, total de votos emitidos y porcentaje de participación.

---

### Distribución Geográfica
```
GET /dashboard/geografico?nivel=departamento
GET /dashboard/geografico?nivel=provincia&departamento=La Paz
GET /dashboard/geografico?nivel=municipio&provincia=Murillo
```
Devuelve votos y mesas agrupados por nivel geográfico. Usar para mapa o gráfico de barras por región.

Parámetros:
- `nivel`: `departamento` | `provincia` | `municipio` (default: `departamento`)
- `departamento`: filtrar por departamento (requerido si nivel=provincia)
- `provincia`: filtrar por provincia (requerido si nivel=municipio)
- `municipio`: filtrar por municipio

---

### Heatmap
```
GET /dashboard/heatmap?metric=participacion&nivel=departamento
```
Devuelve datos para mapa de calor. 

Parámetros:
- `metric`: `participacion` | `votos` (default: `participacion`)
- `nivel`: `departamento` | `provincia` | `municipio` (default: `departamento`)

---

### Transparencia (Actas con imagen)
```
GET /dashboard/transparencia
```
Devuelve listado de actas con su código de mesa, estado OCR, URL de la imagen PDF y timestamp de procesamiento. Usar para la sección de auditoría visual.

`urlImagen` tiene el formato: `acta_{codigoMesa}.pdf`

---

### Trazabilidad de un Acta
```
GET /dashboard/trazabilidad/{codigoActa}
```
Devuelve el historial completo de eventos de una acta específica.

Ejemplo:
```
GET /dashboard/trazabilidad/1060200261005
```

---

### Estado Técnico del Sistema
```
GET /dashboard/tecnico
```
Devuelve métricas de procesamiento: tiempo promedio OCR, tasa de éxito, actas con error, actas pendientes.

---

### Anomalías / Eventos de Alerta
```
GET /dashboard/anomalias
```
Devuelve los últimos eventos que no son `ACTA_RECIBIDA` (rechazos, errores, eventos inusuales), ordenados por timestamp descendente.

---

### Log de Inconsistencias OCR
```
GET /dashboard/logs/inconsistencias
```
Devuelve las actas donde el OCR detectó campos ilegibles, con el listado de qué campos no se pudieron leer.

Respuesta de ejemplo:
```json
{
  "total": 2,
  "registros": [
    {
      "codigoMesa": "1060200261005",
      "archivo": "acta_1060200261005.pdf",
      "camposIlegibles": ["p3"]
    },
    {
      "codigoMesa": "1070400326014",
      "archivo": "acta_1070400326014.pdf",
      "camposIlegibles": ["p3"]
    }
  ]
}
```

---

## Endpoints de Filtros en Cascada

Base URL: `http://localhost:8080/api/v1/filtros`

Usar para poblar los dropdowns de filtro en el dashboard, en este orden:

```
GET /filtros/departamentos
GET /filtros/provincias?departamento=La Paz
GET /filtros/municipios?provincia=Murillo
GET /filtros/recintos?municipio=Mecapaca
GET /filtros/mesas?recinto_id=<ObjectId del recinto>
```

---

## Resumen de todos los endpoints

| Método | URL | Descripción |
|---|---|---|
| `POST` | `http://localhost:8000/api/ocr/scan-all` | Lanzar procesamiento OCR |
| `GET` | `/api/v1/dashboard/kpis` | KPIs generales |
| `GET` | `/api/v1/dashboard/votos-candidato` | Votos por candidato |
| `GET` | `/api/v1/dashboard/participacion` | Participación electoral |
| `GET` | `/api/v1/dashboard/geografico` | Distribución geográfica |
| `GET` | `/api/v1/dashboard/heatmap` | Heatmap |
| `GET` | `/api/v1/dashboard/transparencia` | Actas con imagen |
| `GET` | `/api/v1/dashboard/trazabilidad/:codigoActa` | Historial de un acta |
| `GET` | `/api/v1/dashboard/tecnico` | Estado técnico OCR |
| `GET` | `/api/v1/dashboard/anomalias` | Eventos anómalos |
| `GET` | `/api/v1/dashboard/logs/inconsistencias` | Campos ilegibles por acta |
| `GET` | `/api/v1/filtros/departamentos` | Lista de departamentos |
| `GET` | `/api/v1/filtros/provincias` | Provincias por departamento |
| `GET` | `/api/v1/filtros/municipios` | Municipios por provincia |
| `GET` | `/api/v1/filtros/recintos` | Recintos por municipio |
| `GET` | `/api/v1/filtros/mesas` | Mesas por recinto |

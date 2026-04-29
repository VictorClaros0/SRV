# Módulo de Cómputo Oficial — Documentación de Integración

## Contexto

El sistema SRV tiene dos módulos diferenciados:

- **RRV (Recuento Rápido de Votos):** datos en MongoDB, entrada por imagen o SMS, conteo preliminar.
- **Cómputo Oficial:** datos en PostgreSQL, entrada por CSV validado, resultado definitivo.

Este documento describe qué inconsistencias existían en el proyecto antes de implementar el módulo oficial, qué decisiones se tomaron y por qué, y cómo Pablo (desarrollador del dashboard) debe consumir los datos.

---

## Inconsistencias antes de la implementación

### 1. El modelo `Acta` no tenía identificador de texto

El modelo `Acta` en PostgreSQL solo tenía `ID uint` (número autoincremental). El CSV oficial trae un campo `acta_id` como string (ej: `"ACTA-001"`). No existía ningún lugar donde almacenar ese identificador.

**Consecuencia:** Hubiera sido imposible vincular un acta cargada desde CSV con su identificador original, ni detectar duplicados por código de acta.

### 2. No existía distinción entre actas manuales y oficiales

Todas las actas del sistema caían en la misma tabla `acta` sin ningún campo que indicara su origen. Una acta ingresada manualmente por un operador y una acta cargada desde el CSV oficial eran indistinguibles.

**Consecuencia:** Pablo no podría filtrar "solo actas del cómputo oficial" para el dashboard de resultados definitivos.

### 3. No había auditoría de cargas masivas

No existía ningún registro de quién cargó un CSV, cuándo, cuántas filas procesó, cuántas fueron válidas o rechazadas. En un sistema electoral, esto es un requerimiento crítico para la trazabilidad.

### 4. No había validaciones para el cómputo oficial

El sistema solo tenía validaciones básicas del CRUD manual (votos no negativos, suma correcta). No existían las 13 reglas específicas del cómputo oficial: verificación de acta_id duplicado, mesa no duplicada, total vs ciudadanos habilitados, etc.

### 5. No existía el formato que necesita Pablo

El endpoint `GET /api/v1/actas/resumen` devuelve datos **agregados** por recinto y distribución territorial. Pablo necesita el **detalle por acta** con candidatos en formato array, no campos planos `p1`, `p2`, `p3`, `p4`.

---

## Decisiones de diseño y justificación

### Decisión 1: Reusar la tabla `acta` existente (no crear `oficial_acta`)

**Alternativa descartada:** Crear una tabla `oficial_acta` separada con su propio esquema.

**Decisión tomada:** Agregar solo 2 campos a la tabla `acta` existente: `codigo_acta` y `fuente`.

**Por qué:** La tabla `acta` ya tiene la estructura correcta (P1, P2, P3, P4, votos_nulos, votos_blanco, IDMesa). Duplicar esa estructura en una tabla nueva hubiera creado redundancia, complicado las migraciones y roto el endpoint de resumen (`GET /api/v1/actas/resumen`) que Pablo también puede usar para estadísticas globales.

Los 2 campos nuevos son retrocompatibles: las actas manuales existentes quedan con `fuente = 'MANUAL'` y `codigo_acta` vacío, sin necesidad de modificar los datos actuales.

### Decisión 2: 3 tablas auxiliares independientes

Se crearon `oficial_auditoria`, `oficial_error` y `oficial_evento` como tablas separadas, no como columnas de `acta`.

**Por qué:** Son datos de proceso (quién cargó, qué falló, qué eventos ocurrieron), no datos de resultado. Mezclarlos con la tabla de actas hubiera inflado el modelo principal con columnas nullable que no aplican al flujo manual.

### Decisión 3: La carga CSV requiere rol administrador; la consulta no

El endpoint `POST /api/oficial/csv/upload` exige `RequireAdmin()`. Los endpoints `GET /api/oficial/actas` y `GET /api/oficial/actas/:acta_id` solo requieren JWT válido.

**Por qué:** Cargar datos oficiales es una operación destructiva e irreversible que solo debe hacer personal autorizado. Consultar los resultados es una operación de lectura que Pablo (y cualquier usuario autenticado) debe poder hacer libremente.

### Decisión 4: Los 4 candidatos se exponen como array, no como campos planos

Internamente la base de datos guarda `p1`, `p2`, `p3`, `p4`. El endpoint de Pablo construye un array `candidatos` con `candidato_id`, `nombre` y `votos`.

**Por qué:** El formato con campos planos (`"p1": 100, "p2": 80`) no es consumible por un dashboard sin lógica de mapeo extra. El array con nombres permite iterar directamente en el frontend sin hardcodear los nombres de campo.

---

## Lo que debe usar Pablo

### Endpoint principal

```
GET /api/oficial/actas
Authorization: Bearer <TOKEN>
```

Devuelve todas las actas del cómputo oficial cargadas desde CSV, con el detalle completo geográfico y los 4 candidatos:

```json
[
  {
    "acta_id": "ACTA-001",
    "departamento": "La Paz",
    "provincia": "Murillo",
    "municipio": "El Alto",
    "recinto": "Colegio Ayacucho",
    "mesa": "M-01",
    "candidatos": [
      { "candidato_id": "CAND-01", "nombre": "Candidato 1", "votos": 100 },
      { "candidato_id": "CAND-02", "nombre": "Candidato 2", "votos": 80 },
      { "candidato_id": "CAND-03", "nombre": "Candidato 3", "votos": 60 },
      { "candidato_id": "CAND-04", "nombre": "Candidato 4", "votos": 40 }
    ],
    "votos_nulos": 20,
    "votos_blancos": 15,
    "total_votos": 315,
    "estado": "VALIDADA",
    "fuente": "OFICIAL"
  }
]
```

### Endpoint secundario (acta específica)

```
GET /api/oficial/actas/:acta_id
Authorization: Bearer <TOKEN>
```

Devuelve una sola acta por su código. Útil para búsqueda o detalle.

### Endpoint de resumen global (ya existía)

```
GET /api/v1/actas/resumen
Authorization: Bearer <TOKEN>
```

Devuelve totales agregados por recinto y por distribución territorial. Incluye todas las actas (manuales y oficiales). Útil para gráficos de barra o mapa de resultados a nivel nacional.

---

## Archivos modificados o creados

| Archivo | Tipo | Descripción |
|---|---|---|
| `backend/models/acta.go` | Modificado | +`codigo_acta`, +`fuente` |
| `backend/models/oficial.go` | Nuevo | Modelos `OficialAuditoria`, `OficialError`, `OficialEvento` |
| `backend/database/database.go` | Modificado | +3 modelos en AutoMigrate |
| `backend/handlers/oficial.go` | Nuevo | Upload CSV + 5 GET endpoints |
| `backend/routes/routes.go` | Modificado | +6 rutas bajo `/api/oficial/` |

---

## Error encontrado al levantar el proyecto por primera vez

### Problema: `go.sum` desactualizado

Al ejecutar `docker-compose up --build` por primera vez después de integrar el módulo RRV de Sebastian, el build falló con el siguiente error:

```
models/rrv_acta.go:6:2: missing go.sum entry for module providing package
go.mongodb.org/mongo-driver/v2/bson (imported by github.com/srvof/votos-backend/models); to add:
  go get github.com/srvof/votos-backend/models

database/mongo.go:8:2: missing go.sum entry for module providing package
go.mongodb.org/mongo-driver/v2/mongo (imported by github.com/srvof/votos-backend/database); to add:
  go get github.com/srvof/votos-backend/database
```

### Por qué ocurrió

El archivo `go.mod` tenía declarada la dependencia `go.mongodb.org/mongo-driver/v2 v2.5.0`, pero el archivo `go.sum` (que contiene los hashes de verificación de cada módulo) **no tenía las entradas correspondientes** para ese driver ni sus dependencias transitivas.

Esto ocurrió porque el `go.sum` se genera automáticamente al ejecutar `go mod tidy` o `go get` desde la máquina de desarrollo, y nunca se actualizó al agregar el driver de MongoDB al proyecto.

### Solución aplicada

Como Go no estaba instalado localmente en la máquina de desarrollo, se utilizó un contenedor temporal de Go para ejecutar `go mod tidy` y regenerar el `go.sum` correctamente:

```bash
docker run --rm \
  -v /ruta/al/proyecto/backend:/app \
  -w /app \
  golang:1.23-alpine \
  go mod tidy
```

Este comando descargó todas las dependencias declaradas en `go.mod`, verificó sus hashes y actualizó `go.sum` con las entradas faltantes del driver de MongoDB y sus dependencias transitivas (`xdg-go/scram`, `klauspost/compress`, `youmark/pkcs8`, etc.).

Después de ejecutarlo, el `docker-compose up --build` compiló correctamente.

### Lección

Cada vez que se agrega una dependencia nueva a `go.mod` (manualmente o copiando código de otro módulo), se debe ejecutar `go mod tidy` antes de hacer el build. Si Go no está instalado localmente, usar el contenedor `golang:1.23-alpine` como se mostró arriba.

---

## Formato del CSV de entrada

```
acta_id,departamento,provincia,municipio,recinto,mesa,candidato_1,candidato_2,candidato_3,candidato_4,votos_nulos,votos_blancos,total_votos,ciudadanos_habilitados
ACTA-001,La Paz,Murillo,El Alto,Colegio Ayacucho,M-01,100,80,60,40,20,15,315,400
```

El campo `mesa` debe coincidir exactamente con el campo `codigo` de la tabla `mesa` en PostgreSQL. Si la mesa no existe, la fila es rechazada y queda registrada en `oficial_error`.

---

## Errores encontrados al levantar el proyecto (sesión de pruebas)

### Error 1: `REPMGR_USERNAME` faltante en pg-0 y pg-1

**Síntoma:**
```
pg-0-1 | ERROR ==> The repmgr credentials are mandatory.
Set the environment variables REPMGR_USERNAME and REPMGR_PASSWORD
```

**Por qué ocurrió:** La imagen Bitnami `postgresql-repmgr:17.6.0` requiere que `REPMGR_USERNAME` se declare explícitamente. El `docker-compose.yml` solo tenía `REPMGR_PASSWORD` pero no el usuario.

**Solución:** Se agregó `REPMGR_USERNAME=repmgr` en la sección `environment` de los servicios `pg-0` y `pg-1` en `docker-compose.yml`.

---

### Error 2: Puerto del host ya en uso (5433 y luego 5434)

**Síntoma:**
```
failed to bind host port 0.0.0.0:5433/tcp: address already in use
failed to bind host port 0.0.0.0:5434/tcp: address already in use
```

**Por qué ocurrió:** El proyecto no tenía archivo `.env` (solo `.env.example`). Al no existir el `.env`, Docker Compose no tomaba la variable `PGPOOL_HOST_PORT` y usaba el valor por defecto `5433`, que ya estaba ocupado por otro proceso del sistema. Al cambiarlo a `5434`, ese puerto también estaba ocupado por un contenedor anterior que no se había limpiado.

**Solución:**
1. Se creó el archivo `.env` copiando `.env.example`: `cp .env.example .env`
2. Se cambió `PGPOOL_HOST_PORT=5435` (puerto libre verificado con `ss -tlnp`)

> **Nota:** El puerto del host solo afecta a herramientas externas como DBeaver. La API interna se conecta a pgpool por la red Docker directamente y no usa este puerto.

---

### Error 3: Contenedores bloqueados — "permission denied" al detener

**Síntoma:**
```
Error response from daemon: cannot stop container: <id>: permission denied
sudo kill -9 <pid>: Permission denied
```

**Por qué ocurrió:** En algún momento se ejecutó `sudo docker-compose up`, lo que creó contenedores bajo el contexto de root. Al volver a ejecutar sin sudo, el daemon de Docker (que sí corre como root) tenía un conflicto de namespace con esos contenedores. Los procesos aparecían en el host corriendo como usuario `dnsmasq` — indicativo de un mapeo de namespaces de usuario que impidió enviar señales incluso con sudo.

**Solución:** Se detuvo el daemon de Docker, se eliminaron manualmente los archivos de estado de los contenedores bloqueados, y se reinició el daemon:
```bash
sudo systemctl stop docker docker.socket
sudo rm -rf /var/lib/docker/containers/<id1>* /var/lib/docker/containers/<id2>* ...
sudo systemctl start docker
docker-compose down -v
docker-compose up --build
```

> **Lección:** Nunca mezclar `sudo docker-compose` con `docker-compose` sin sudo en la misma máquina. Elegir uno y usarlo siempre.

---

### Error 4: `mongo-setup` falla — Replica Set en estado corrupto

**Síntoma:**
```
Container srv-mongo-setup-1 Error
service "mongo-setup" didn't complete successfully: exit 1
```

**Por qué ocurrió:** Los volúmenes de MongoDB (`mongo1_data`, `mongo2_data`) tenían datos de corridas anteriores que terminaron de forma abrupta. El Replica Set quedó en un estado parcialmente inicializado que `mongo-setup` no podía resolver.

**Solución:** Bajar el stack eliminando los volúmenes para empezar desde cero:
```bash
docker-compose down -v
docker-compose up --build
```

---

### Error 5: Seed inicial tardaba más de 17 minutos

**Síntoma:** El proyecto arrancaba pero la API tardaba más de 17 minutos en estar lista porque el seed de datos iniciales insertaba 3508 mesas una por una con ~300ms por INSERT.

**Por qué ocurrió:** El código en `csv_seed.go` llamaba `db.Create(&item)` dentro de un bucle, generando una transacción de red por cada fila.

**Solución:** Se reemplazó el bucle de `db.Create()` individual por `db.CreateInBatches(items, 100)` en las tres funciones de seed (`seedDistribuciones`, `seedRecintos`, `seedMesas`). Esto agrupa las inserciones en lotes de 100 registros por query, reduciendo el tiempo del seed de ~17 minutos a segundos.

> **Nota:** El seed solo corre la primera vez (cuando las tablas están vacías). En corridas posteriores se omite automáticamente.

---

## Validaciones aplicadas (FASE 5)

| # | Regla |
|---|---|
| 1 | `acta_id` obligatorio |
| 2 | `departamento` obligatorio |
| 3 | `municipio` obligatorio |
| 4 | `recinto` obligatorio |
| 5 | `mesa` obligatoria |
| 6 | Votos no negativos |
| 7-8 | `candidato_1+candidato_2+candidato_3+candidato_4+votos_nulos+votos_blancos = total_votos` |
| 9 | `total_votos <= ciudadanos_habilitados` |
| 10 | `acta_id` no duplicado (en CSV ni en BD) |
| 11 | Mesa sin acta oficial previa (en CSV ni en BD) |
| 12 | Mesa debe existir en la base de datos |
| 13 | Todos los campos numéricos deben ser enteros válidos |

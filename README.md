# Conteo de votos — API Go + PostgreSQL HA (Docker)

Backend REST en **Go (Gin + GORM + JWT)** para registro manual de actas de votación. La base de datos es **PostgreSQL** en alta disponibilidad: **dos nodos** con **repmgr** (streaming replication) y **Pgpool-II** como punto único de conexión con balanceo y tolerancia a fallos.

## Requisitos

- Docker y Docker Compose v2
- (Opcional) Go 1.22+ para ejecutar la API fuera de Docker

## Puesta en marcha

1. Copiar variables de entorno:

   ```bash
   cp .env.example .env
   ```

   Editar `.env` y definir contraseñas y `JWT_SECRET`.

2. Levantar el stack:

   ```bash
   docker compose up -d --build
   ```

3. Comprobar salud:

   - API: `GET http://localhost:8080/health`
   - Login (usuario inicial `admin` / contraseña `ADMIN_PASSWORD` del `.env`):

     ```http
     POST http://localhost:8080/api/v1/auth/login
     Content-Type: application/json

     {"usuario":"admin","contrasena":"admin123"}
     ```

4. Con el token JWT (`Authorization: Bearer <token>`), usar los demás endpoints bajo `/api/v1/`.

## Arquitectura

| Servicio     | Rol |
|-------------|-----|
| `pg-primary` | PostgreSQL primary (repmgr) |
| `pg-standby` | Réplica en caliente |
| `pgpool`     | Pool + detección de primary; la API solo conecta aquí (`DB_HOST=pgpool`) |
| `api`        | Backend Go |

Desde el host, Pgpool expone el puerto `PGPOOL_HOST_PORT` (por defecto **5433**) para `psql`/clientes; dentro de Compose la API usa `pgpool:5432`.

## Modelo de datos

Entidades alineadas al diagrama ERD: `usuario`, `distribucion_territorial`, `recinto_electoral`, `mesa`, `acta` (con auditoría y borrado lógico en tablas con campos `fecha_eliminado`).

Validación de **actas**: `p1+p2+p3+p4` debe igualar `votosValidos`; la suma de votos emitidos no puede superar `cantidadHabilitada` de la mesa.

## Endpoints principales

| Método | Ruta | Notas |
|--------|------|--------|
| POST | `/api/v1/auth/login` | Público |
| GET | `/api/v1/auth/me` | JWT |
| POST | `/api/v1/auth/register` | JWT + admin |
| CRUD | `/api/v1/usuarios` | List/delete: admin; get/update: propio usuario o admin |
| CRUD | `/api/v1/distribuciones` | Lectura: cualquier usuario autenticado; escritura: admin |
| CRUD | `/api/v1/recintos` | Idem |
| CRUD | `/api/v1/mesas` | Idem |
| CRUD | `/api/v1/actas` | Lectura/escritura: cualquier usuario autenticado |
| GET | `/api/v1/actas/resumen` | Totales por recinto y por distribución territorial |
| GET | `/health` | Ping a la base vía GORM |

## Prueba de failover (manual)

1. Con el stack en marcha, insertar datos (por API o conectando a `localhost:5433` con el usuario `votos`).
2. Detener el primary:

   ```bash
   docker compose stop pg-primary
   ```

3. Tras unos segundos, **repmgr** promueve el standby. La API sigue respondiendo (reintentos de conexión y mismo DSN hacia Pgpool).
4. Comprobar `GET /health` y operaciones de lectura/escritura.
5. Para volver a levantar el nodo antiguo como réplica (según estado del clúster Bitnami), consultar la documentación de [Bitnami postgresql-repmgr](https://github.com/bitnami/containers/tree/main/bitnami/postgresql-repmgr).

> **Nota:** En entornos reales conviene supervisión (Prometheus, logs de repmgr/pgpool) y backups (pg_basebackup, WAL).

## Desarrollo local (sin Docker para la API)

```bash
cd backend
set DB_HOST=localhost   # o IP del host donde corre Pgpool
set DB_PORT=5433        # PGPOOL_HOST_PORT
go run .
```

## Licencia

Proyecto académico / uso interno según corresponda.

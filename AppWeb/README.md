# Sistema Nacional de Cómputo Electoral - Frontend

Frontend del dashboard administrativo para el Sistema Nacional de Cómputo Electoral. Esta aplicación permite la visualización en tiempo real de métricas electorales, comparación de flujos de datos (RRV preliminar vs Cómputo Oficial), monitoreo de infraestructura y gestión de inconsistencias en las actas.

## 🚀 Tecnologías Principales

- **Framework**: React 18+ (Vite)
- **Lenguaje**: TypeScript
- **Estilos**: Tailwind CSS v4 (con soporte avanzado para Dark/Light Mode)
- **Estado Global**: Zustand (Persistencia en sesión para Autenticación)
- **Data Fetching**: Axios + TanStack React Query v5
- **Enrutamiento**: React Router v6
- **Gráficos**: Recharts
- **Iconos**: Lucide React

## 📁 Arquitectura del Proyecto

El proyecto sigue una estructura **Screaming Architecture (Feature-based)** para facilitar la escalabilidad:

```
src/
├── config/        # Configuraciones globales (Axios, React Query, LoadAbort)
├── features/      # Módulos del dominio
│   ├── auth/      # Lógica de autenticación, hooks (Zustand) y vistas
│   ├── dashboard/ # Componentes core, vistas y servicios del dashboard
│   └── scanner/   # Módulo de escaneo de actas
├── hooks/         # Hooks globales (ej. useDarkMode)
├── layouts/       # Componentes estructurales (MainLayout, Sidebar)
└── router/        # Configuración de enrutamiento y guardias (PrivateRoutes)
```

## 🛠️ Instalación y Configuración

1. **Instalar dependencias**:
   ```bash
   npm install
   ```

2. **Levantar el servidor de desarrollo**:
   ```bash
   npm run dev
   ```

3. **Compilar para producción**:
   ```bash
   npm run build
   ```

## 🔐 Autenticación (Mock)

Actualmente la autenticación está simulada mediante la capa `useAuthStore`. 
Al ingresar cualquier usuario y contraseña en la vista de `/auth/login`, la aplicación generará una sesión simulada con rol `ADMIN` que redirige al Dashboard.

## 📊 Endpoints Analíticos

Los servicios de datos (`src/features/dashboard/services/dashboardService.ts`) actualmente devuelven información simulada (Mocks) estructurada en las siguientes operaciones clave:

- `getKPIs()`: Métricas principales (actas procesadas, diferencias, confiabilidad).
- `getVotosCandidato()`: Comparativa de votos por partido (RRV vs Oficial).
- `getRRVvsOficial()`: Flujo temporal de recepción de actas.
- `getGeografico()`: Avance de computo por departamento.
- `getTecnico()`: Salud del sistema (latencia, errores API, uptime).
- `getInconsistencias()`: Auditoría de actas que no cuadran entre pipelines.

*Nota: Estos servicios están listos para ser reemplazados por los endpoints reales del equipo de Comparación (Pablo).*

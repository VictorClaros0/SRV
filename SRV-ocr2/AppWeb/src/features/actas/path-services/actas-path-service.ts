const ACTAS_BASE = "actas";
const TRANSCRIPCION_BASE = "transcripcion";
const RECINTOS_BASE = "recintos";

export const ActasPath = {
    LIST: `${ACTAS_BASE}`,
    DETAIL: (id: number | string) => `${ACTAS_BASE}/${id}`,
    CREATE: `${ACTAS_BASE}`,
    UPDATE: (id: number | string) => `${ACTAS_BASE}/${id}`,
    DELETE: (id: number | string) => `${ACTAS_BASE}/${id}`,
    RESUMEN: `${ACTAS_BASE}/resumen`,
    PARA_TRANSCRIBIR: `${ACTAS_BASE}/para-transcribir`,
    SIMULAR: `${TRANSCRIPCION_BASE}/simular`,
} as const;

export const RecintosPath = {
    LIST: `${RECINTOS_BASE}`,
} as const;

// Path: http://localhost:8080/api/v1/actas
// Path: http://localhost:8080/api/v1/actas/:id
// Path: http://localhost:8080/api/v1/actas/resumen
// Path: http://localhost:8080/api/v1/actas/para-transcribir
// Path: http://localhost:8080/api/v1/transcripcion/simular
// Path: http://localhost:8080/api/v1/recintos

export interface TecnicoResponse {
    timestamp: string;
    latenciaMs: { p50: number; p95: number; p99: number };
    throughputActasMin: number;
    disponibilidad: string;
    uptimeSegundos: number;
    seguridad: { jwtRequired: boolean; httpsRequerido: boolean; tlsVersion: string };
    fuente: { estado: string; totalActas: number; tabla: string };
}

export interface ResultadosOficialesResponse {
    resumen?: {
        totalActas?: number;
        actasTranscritas?: number;
        actasObservadas?: number;
    };
    comparativa?: { candidato: string; votos: number }[];
    porDepartamento?: { ubicacion: string; totalActas: number; votosValidos: number }[];
}

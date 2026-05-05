export interface ResumenResultados {
    totalActas: number;
    actasTranscritas: number;
    actasObservadas: number;
    actasImpresa: number;
    pctProcesadas: number;
}

export interface ComparativaItem {
    candidato: string;
    votos: number;
    pct: number;
}

export interface DesgloseUbicacion {
    ubicacion: string;
    p1: number;
    p1_pct: number;
    p2: number;
    p2_pct: number;
    p3: number;
    p3_pct: number;
    p4: number;
    p4_pct: number;
    votosValidos: number;
    votosNulos: number;
    votosBlanco: number;
    totalActas: number;
}

export interface VelocidadStats {
    actasUltimaHora: number;
    actasUltimas24h: number;
}

export interface ResultadosResponse {
    resumen: ResumenResultados;
    comparativa: ComparativaItem[];
    margen_victoria: number;
    desglose: DesgloseUbicacion[];
    velocidad: VelocidadStats;
}

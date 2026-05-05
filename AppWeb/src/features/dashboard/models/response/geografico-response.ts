export type GeoNivel = "departamento" | "municipio" | "provincia" | "recinto";

export interface GeoItem {
    nombre: string;
    totalActas: number;
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
    ganador: string;
    margen: number;
}

export interface GeograficoResponse {
    nivel: GeoNivel;
    items: GeoItem[];
}

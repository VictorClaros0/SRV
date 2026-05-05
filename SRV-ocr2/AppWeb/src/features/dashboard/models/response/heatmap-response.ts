export type HeatmapMetric = "ganador" | "inconsistencias" | "procesamiento";

export interface HeatmapItem {
    nombre: string;
    valor: number;
    ganador: string;
    votos: number;
}

export interface HeatmapResponse {
    metric: HeatmapMetric;
    nivel: string;
    items: HeatmapItem[];
}

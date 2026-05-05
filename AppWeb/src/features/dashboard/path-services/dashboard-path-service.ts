const BASE = "dashboard";
const FILTROS = "filtros";

export const DashboardPath = {
    KPIS: `${BASE}/kpis`,
    GEOGRAFICO: `${BASE}/geografico`,
    HEATMAP: `${BASE}/heatmap`,
    ANOMALIAS: `${BASE}/anomalias`,

    RESULTADOS: `resultados`,
    AUDITORIA: `auditoria`,

    FILTRO_DEPARTAMENTOS: `${FILTROS}/departamentos`,
    FILTRO_MUNICIPIOS: `${FILTROS}/municipios`,
    FILTRO_PROVINCIAS: `${FILTROS}/provincias`,
    FILTRO_RECINTOS: `${FILTROS}/recintos`,
    FILTRO_MESAS: `${FILTROS}/mesas`,
} as const;

export type DashboardPath = (typeof DashboardPath)[keyof typeof DashboardPath];

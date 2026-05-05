import { httpClient } from "../../../config/axios";
import { loadAbort } from "../../../config/load-abort";
import type { UseApiCall } from "../../../config/useApicall";
import { DashboardPath } from "../path-services/dashboard-path-service";

import type { KpisResponse } from "../models/response/kpis-response";
import type { ResultadosResponse } from "../models/response/resultados-response";
import type { GeograficoResponse } from "../models/response/geografico-response";
import type { HeatmapResponse, HeatmapMetric } from "../models/response/heatmap-response";
import type { AnomaliasResponse } from "../models/response/anomalias-response";
import type { GeoFilters, GeoNivel } from "../models/request/geografico-request";
import type { FiltroItem } from "../models/response/filtros-response";

export type { GeoFilters };

const get = <T>(url: string, params?: Record<string, unknown>): UseApiCall<T> => {
    const controller = loadAbort();
    return {
        call: httpClient.get<T>(url, { params, signal: controller.signal }),
        controller,
    };
};

const cleanGeo = (f: GeoFilters = {}) => ({
    departamento: f.departamento || undefined,
    municipio: f.municipio || undefined,
    provincia: f.provincia || undefined,
    recinto: f.recinto || undefined,
    mesa: f.mesa || undefined,
});

export const dashboardService = {
    getKPIs: () => get<KpisResponse>(DashboardPath.KPIS),

    getResultados: (f?: GeoFilters) =>
        get<ResultadosResponse>(DashboardPath.RESULTADOS, cleanGeo(f)),

    getGeografico: (nivel: GeoNivel, f?: GeoFilters) =>
        get<GeograficoResponse>(DashboardPath.GEOGRAFICO, {
            nivel,
            departamento: f?.departamento || undefined,
            municipio: f?.municipio || undefined,
            provincia: f?.provincia || undefined,
        }),

    getHeatmap: (metric: HeatmapMetric = "ganador", nivel: string = "departamento") =>
        get<HeatmapResponse>(DashboardPath.HEATMAP, { metric, nivel }),

    getAnomalias: () => get<AnomaliasResponse>(DashboardPath.ANOMALIAS),

    getAuditoria: () => get<{ total: number; actas: unknown[] }>(DashboardPath.AUDITORIA),

    getDepartamentos: () => get<FiltroItem[]>(DashboardPath.FILTRO_DEPARTAMENTOS),

    getMunicipios: (departamento?: string) =>
        get<FiltroItem[]>(DashboardPath.FILTRO_MUNICIPIOS, {
            departamento: departamento || undefined,
        }),

    getProvincias: (departamento?: string, municipio?: string) =>
        get<FiltroItem[]>(DashboardPath.FILTRO_PROVINCIAS, {
            departamento: departamento || undefined,
            municipio: municipio || undefined,
        }),

    getRecintos: (params: { departamento?: string; municipio?: string; provincia?: string } = {}) =>
        get<FiltroItem[]>(DashboardPath.FILTRO_RECINTOS, {
            departamento: params.departamento || undefined,
            municipio: params.municipio || undefined,
            provincia: params.provincia || undefined,
        }),

    getMesas: (recinto?: string) =>
        get<FiltroItem[]>(DashboardPath.FILTRO_MESAS, {
            recinto: recinto || undefined,
        }),
};

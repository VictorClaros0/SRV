import { httpClientRrv } from "../../../config/axios-rrv";
import { loadAbort } from "../../../config/load-abort";
import type { UseApiCall } from "../../../config/useApicall";
import type {
    RrvKpis,
    RrvVotosCandidatoResponse,
    RrvParticipacion,
    RrvGeograficoResponse,
    RrvHeatmapResponse,
    RrvTransparenciaResponse,
    RrvTecnico,
    RrvAnomaliasResponse,
    RrvLogInconsistenciasResponse,
} from "../models/response/rrv-responses";

const get = <T>(url: string, params?: Record<string, unknown>): UseApiCall<T> => {
    const controller = loadAbort();
    return {
        call: httpClientRrv.get<T>(url, { params, signal: controller.signal }),
        controller,
    };
};

export interface RrvGeoFilters {
    departamento?: string;
    provincia?: string;
    municipio?: string;
}

export const rrvService = {
    getKPIs: () => get<RrvKpis>("dashboard/kpis"),

    getVotosCandidato: () => get<RrvVotosCandidatoResponse>("dashboard/votos-candidato"),

    getParticipacion: () => get<RrvParticipacion>("dashboard/participacion"),

    getGeografico: (nivel: string = "departamento", f: RrvGeoFilters = {}) =>
        get<RrvGeograficoResponse>("dashboard/geografico", {
            nivel,
            departamento: f.departamento || undefined,
            provincia: f.provincia || undefined,
            municipio: f.municipio || undefined,
        }),

    getHeatmap: (metric: string = "participacion", nivel: string = "departamento") =>
        get<RrvHeatmapResponse>("dashboard/heatmap", { metric, nivel }),

    getTransparencia: () => get<RrvTransparenciaResponse>("dashboard/transparencia"),

    getTecnico: () => get<RrvTecnico>("dashboard/tecnico"),

    getAnomalias: () => get<RrvAnomaliasResponse>("dashboard/anomalias"),

    getLogInconsistencias: () =>
        get<RrvLogInconsistenciasResponse>("dashboard/logs/inconsistencias"),
};

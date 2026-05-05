import type { CandidatoTotal } from "./kpis-response";

export interface VotosCandidatoResponse {
    totalValidos: number;
    items: CandidatoTotal[];
}

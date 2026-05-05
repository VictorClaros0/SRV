export interface CandidatoTotal {
    candidato: string;
    votos: number;
    porcentaje: number;
}

export interface KpisResponse {
    totalActas: number;
    actasTranscritas: number;
    actasObservadas: number;
    actasPendientes: number;
    porcentajePublicadas: number;
    votosValidos: number;
    votosNulos: number;
    votosBlanco: number;
    ganador: CandidatoTotal;
    segundo: CandidatoTotal;
    margenVictoria: number;
}

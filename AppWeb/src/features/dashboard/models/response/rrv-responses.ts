export interface RrvKpis {
    totalActasEnBD: number;
    totalActasProcesadas: number;
    totalActasInconsistentes: number;
    totalMesas: number;
    votosP1: number;
    votosP2: number;
    votosP3: number;
    votosP4: number;
    votosNulos: number;
    votosBlancos: number;
    votosValidos: number;
    totalHabilitados: number;
    porcentajeParticipacion: number;
}

export interface RrvCandidato {
    nombre: string;
    campo: string;
    votos: number;
    porcentaje: number;
}

export interface RrvVotosCandidatoResponse {
    candidatos: RrvCandidato[];
}

export interface RrvParticipacion {
    totalHabilitados: number;
    totalVotantes: number;
    porcentajeParticipacion: number;
}

export interface RrvGeoItem {
    nombre: string;
    votosP1: number;
    votosP2: number;
    votosP3: number;
    votosP4: number;
    votosNulos: number;
    votosValidos: number;
    habilitados: number;
    participacion: number;
}

export interface RrvGeograficoResponse {
    nivel: string;
    datos: RrvGeoItem[];
}

export interface RrvHeatItem {
    nombre: string;
    valor: number;
}

export interface RrvHeatmapResponse {
    nivel: string;
    metric: string;
    datos: RrvHeatItem[];
}

export interface RrvTransparenciaActa {
    codigoMesa: string;
    departamento: string;
    provincia: string;
    municipio: string;
    recinto: string;
    urlImagen: string;
    p1: number;
    p2: number;
    p3: number;
    p4: number;
    votosNulos: number;
    votosBlanco: number;
    votosValidos: number;
}

export interface RrvTransparenciaResponse {
    actas: RrvTransparenciaActa[];
}

export interface RrvTecnico {
    totalRRVActas: number;
    totalActasProcesadas: number;
    totalActasInconsistentes: number;
    totalEventos: number;
    eventosPorTipo: Record<string, number>;
    actasUltimas24h: number;
}

export interface RrvEvento {
    _id?: string;
    tipo?: string;
    actaId?: string;
    fecha?: string;
    payload?: unknown;
    [k: string]: unknown;
}

export interface RrvAnomaliasResponse {
    total: number;
    anomalias: RrvEvento[];
}

export interface RrvLogInconsistencia {
    codigoMesa: string;
    archivo: string;
    camposIlegibles: string[];
}

export interface RrvLogInconsistenciasResponse {
    total: number;
    registros: RrvLogInconsistencia[];
}

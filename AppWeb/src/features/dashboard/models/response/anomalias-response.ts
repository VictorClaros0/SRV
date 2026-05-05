export interface Anomalia {
    codigoActa: number;
    nroMesa: number;
    tipo: string;
    severidad: "alta" | "media" | "baja";
    detalle: string;
    valoresEsperados?: number;
    valoresReales?: number;
}

export interface AnomaliasResponse {
    total: number;
    anomalias: Anomalia[];
}

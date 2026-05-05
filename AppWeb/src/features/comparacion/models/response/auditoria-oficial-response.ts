export interface AuditoriaOficialResponse {
    total?: number;
    actas?: DataActaAudit[]
};

export interface DataActaAudit {
    id?: string | number;
    codigoActa: number;
    codigoRecinto: number;
    nroMesa: number;
    estado: string;
    observaciones?: string;
    fechaCreacion: string;
}

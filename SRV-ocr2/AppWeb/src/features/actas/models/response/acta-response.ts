export type ActaEstado = "impresa" | "transcrita" | "observada";

export interface Acta {
    id: number;
    codigoActa: string;
    codigoRecinto: string;
    nroMesa: number | string;
    estado: ActaEstado;
    p1: number;
    p2: number;
    p3: number;
    p4: number;
    votosValidos: number;
    votosNulos: number;
    votosBlanco: number;
    observaciones?: string;
}

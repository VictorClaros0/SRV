import type { ActaEstado } from "../response/acta-response";

export interface ActaCreateRequest {
    codigoActa: number;
    codigoRecinto: number;
    nroMesa: number;
    estado: ActaEstado;
    papeletasAnfora: number;
    papeletasNoUsadas: number;
    p1: number;
    p2: number;
    p3: number;
    p4: number;
    votosValidos: number;
    votosNulos: number;
    votosBlanco: number;
    observaciones: string;
    aperturaHora: number;
    aperturaMinutos: number;
    cierreHora: number;
    cierreMinutos: number;
}

export const emptyActaCreate = (): ActaCreateRequest => ({
    codigoActa: 0,
    codigoRecinto: 0,
    nroMesa: 1,
    estado: "impresa",
    papeletasAnfora: 0,
    papeletasNoUsadas: 0,
    p1: 0,
    p2: 0,
    p3: 0,
    p4: 0,
    votosValidos: 0,
    votosNulos: 0,
    votosBlanco: 0,
    observaciones: "",
    aperturaHora: 8,
    aperturaMinutos: 0,
    cierreHora: 18,
    cierreMinutos: 0,
});

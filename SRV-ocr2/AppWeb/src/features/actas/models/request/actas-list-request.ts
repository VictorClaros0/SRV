import type { ActaEstado } from "../response/acta-response";

export interface ActasListParams {
    estado?: ActaEstado;
    codigoRecinto?: number | string;
    nroMesa?: number | string;
    pagina?: number;
    por_pagina?: number;
}

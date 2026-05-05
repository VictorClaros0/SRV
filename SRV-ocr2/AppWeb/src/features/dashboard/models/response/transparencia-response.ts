export interface ActaPublicada {
    codigoActa: number;
    fechaPublicacion: string;
    urlImagen: string;
}

export interface TransparenciaResponse {
    porcentajeActasPublicadas: number;
    actasConImagen: number;
    actasSinImagen: number;
    ultimasPublicadas: ActaPublicada[];
}

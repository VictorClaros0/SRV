export interface ParticipacionResponse {
    habilitados: number;
    votantes: number;
    abstencion: number;
    tasaParticipacion: number;
    desglose: {
        validos: number;
        nulos: number;
        blanco: number;
    };
}

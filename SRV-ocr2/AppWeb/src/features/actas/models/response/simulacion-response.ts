export interface SimulacionResponse {
    success: boolean;
    estado?: "COMPLETADO" | "EN_PROGRESO" | string;
    source?: string;
    actualizadas?: number;
    observadas?: number;
    message?: string;
    hint?: string;
}

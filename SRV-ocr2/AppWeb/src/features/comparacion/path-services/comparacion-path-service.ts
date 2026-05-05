const BASE_PATH = "auditoria";

export const ComparacionPath = {
    LIST: `${BASE_PATH}`,
} as const;

export type ComparacionPath = (typeof ComparacionPath)[keyof typeof ComparacionPath];

// Path: http://localhost:8080/api/v1/comparacion

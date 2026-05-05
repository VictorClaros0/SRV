const BASE_PATH = "auth";

export const AuthPath = {
    LOGIN: `${BASE_PATH}/login`,
    ME: `${BASE_PATH}/me`,
    REGISTER: `${BASE_PATH}/register`,
} as const;

export type AuthPath = (typeof AuthPath)[keyof typeof AuthPath];

// Path: http://localhost:3000/api/auth/login
// Path: http://localhost:3000/api/auth/me
// Path: http://localhost:3000/api/auth/register

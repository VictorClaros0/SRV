export interface LoginResponse {
    token: string;
    expiresIn: number;
    esAdmin: boolean;
}

export interface SuccessPostAuth {
    response: LoginResponse;
    username: string;
}

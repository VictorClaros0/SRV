import { jwtDecode, type JwtPayload } from "jwt-decode";

export interface DecodedJWT extends JwtPayload {
    userId: number;
    esAdmin: boolean;
}

export function jwtDecoder(token: string): DecodedJWT {
    try {
        return jwtDecode<DecodedJWT>(token);
    } catch (error) {
        console.error("Failed to decode JWT:", error);
        throw new Error("Invalid JWT token");
    }
}

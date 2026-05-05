import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";
import { jwtDecoder } from "../../../config/jwtDecode";
import type { SuccessPostAuth } from "../models/response/success-post-auth";

type AuthStore = {
    jwt: string | null;
    userId: number | null;
    username: string | null;
    role: string | null;
    esAdmin: boolean;
    expiresJwt: number | null;
    isAuthenticated: boolean;
    setAuthUser: (data: SuccessPostAuth) => void;
    clearAuthUser: () => void;
};

const initialState = {
    jwt: null,
    userId: null,
    username: null,
    role: null,
    esAdmin: false,
    expiresJwt: null,
    isAuthenticated: false,
};

export const useAuthStore = create<AuthStore>()(
    persist(
        (set) => ({
            ...initialState,
            setAuthUser: ({ response, username }: SuccessPostAuth) => {
                const decoded = jwtDecoder(response.token);
                const expirationTime = Date.now() + response.expiresIn * 1000;

                set({
                    jwt: response.token,
                    userId: decoded.userId,
                    username,
                    role: response.esAdmin ? "ADMIN" : "USER",
                    esAdmin: response.esAdmin,
                    expiresJwt: expirationTime,
                    isAuthenticated: true,
                });
            },
            clearAuthUser: () => set({ ...initialState }),
        }),
        {
            name: "auth-storage",
            storage: createJSONStorage(() => sessionStorage),
        },
    ),
);

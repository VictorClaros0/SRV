import type { AxiosError, AxiosInstance, InternalAxiosRequestConfig } from "axios";
import axios from "axios";
import { useAuthStore } from "../features/auth/hooks/useAuthStore";

const baseURL: string =
    (import.meta.env.VITE_API_URL as string | undefined) ??
    "http://localhost:8090/api/v1";

const createAxiosInstance = (): AxiosInstance => {
    return axios.create({ baseURL });
};

const setupInterceptors = (httpClient: AxiosInstance) => {
    httpClient.interceptors.request.use(
        (config: InternalAxiosRequestConfig) => {
            if (!config.headers["Content-Type"]) {
                config.headers["Content-Type"] = "application/json";
            }
            const { jwt } = useAuthStore.getState();
            if (jwt) {
                config.headers["Authorization"] = `Bearer ${jwt}`;
            }
            return config;
        },
        (error: AxiosError) => Promise.reject(error),
    );

    httpClient.interceptors.response.use(
        (response) => response,
        (error: AxiosError) => {
            if (error.response?.status === 401) {
                useAuthStore.getState().clearAuthUser();
                if (typeof window !== "undefined" && !window.location.pathname.startsWith("/auth")) {
                    window.location.href = "/auth/login";
                }
            }
            return Promise.reject(error);
        },
    );
};

const initAxios = (): AxiosInstance => {
    const httpClient = createAxiosInstance();
    setupInterceptors(httpClient);
    return httpClient;
};

export const httpClient: AxiosInstance = initAxios();

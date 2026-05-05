import axios, { type AxiosInstance, type InternalAxiosRequestConfig } from "axios";

const baseURL: string =
    (import.meta.env.VITE_RRV_API_URL as string | undefined) ??
    "http://localhost:8081/api/v1";

const httpClientRrv: AxiosInstance = axios.create({ baseURL });

httpClientRrv.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    if (!config.headers["Content-Type"]) {
        config.headers["Content-Type"] = "application/json";
    }
    return config;
});

export { httpClientRrv };

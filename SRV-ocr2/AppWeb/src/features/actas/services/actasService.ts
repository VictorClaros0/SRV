import { httpClient } from "../../../config/axios";
import { loadAbort } from "../../../config/load-abort";
import type { UseApiCall } from "../../../config/useApicall";
import { ActasPath } from "../path-services/actas-path-service";
import type { Acta } from "../models/response/acta-response";
import type { ResumenActasResponse } from "../models/response/resumen-actas-response";
import type { ActasListParams } from "../models/request/actas-list-request";
import type { ActaCreateRequest } from "../models/request/acta-create-request";
import type { ActaUpdateRequest } from "../models/request/acta-update-request";

export const actasService = {
    list: (params: ActasListParams = {}): UseApiCall<Acta[]> => {
        const controller = loadAbort();
        return {
            call: httpClient.get<Acta[]>(ActasPath.LIST, {
                params,
                signal: controller.signal,
            }),
            controller,
        };
    },
    get: (id: number | string): UseApiCall<Acta> => {
        const controller = loadAbort();
        return {
            call: httpClient.get<Acta>(ActasPath.DETAIL(id), { signal: controller.signal }),
            controller,
        };
    },
    create: (payload: ActaCreateRequest): UseApiCall<Acta> => {
        const controller = loadAbort();
        return {
            call: httpClient.post<Acta>(ActasPath.CREATE, payload, { signal: controller.signal }),
            controller,
        };
    },
    update: (id: number | string, payload: ActaUpdateRequest): UseApiCall<Acta> => {
        const controller = loadAbort();
        return {
            call: httpClient.put<Acta>(ActasPath.UPDATE(id), payload, { signal: controller.signal }),
            controller,
        };
    },
    delete: (id: number | string): UseApiCall<void> => {
        const controller = loadAbort();
        return {
            call: httpClient.delete<void>(ActasPath.DELETE(id), { signal: controller.signal }),
            controller,
        };
    },
    resumen: (): UseApiCall<ResumenActasResponse> => {
        const controller = loadAbort();
        return {
            call: httpClient.get<ResumenActasResponse>(ActasPath.RESUMEN, { signal: controller.signal }),
            controller,
        };
    },
    paraTranscribir: (): UseApiCall<Acta[]> => {
        const controller = loadAbort();
        return {
            call: httpClient.get<Acta[]>(ActasPath.PARA_TRANSCRIBIR, { signal: controller.signal }),
            controller,
        };
    },
    // en actasService.ts

    simular: () => {
        const controller = new AbortController();

        // Hacemos el fetch directo al webhook de n8n
        const call = fetch('http://localhost:5678/webhook/trigger-transcripcion', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({}),
            signal: controller.signal,
        }).then(async (r) => {
            if (!r.ok) {
                const txt = await r.text().catch(() => r.statusText);
                throw new Error(txt || 'n8n no respondió. ¿El workflow está activo?');
            }
            // n8n a veces no devuelve nada (204) o devuelve texto plano.
            // Aseguramos que siempre devuelva un objeto data para no romper el frontend.
            const text = await r.text();
            const data = text ? JSON.parse(text) : { success: true };
            return { data };
        });

        return { call, controller };
    },
};

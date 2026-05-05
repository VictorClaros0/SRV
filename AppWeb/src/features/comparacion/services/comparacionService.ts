import { httpClient } from "../../../config/axios";
import { loadAbort } from "../../../config/load-abort";
import type { UseApiCall } from "../../../config/useApicall";
import type { AuditoriaOficialResponse } from "../models/response/auditoria-oficial-response";
import { ComparacionPath } from "../path-services/comparacion-path-service";

export const comparacionService = {
    list: (): UseApiCall<AuditoriaOficialResponse> => {
        const controller = loadAbort();
        return {
            call: httpClient.get<AuditoriaOficialResponse>(ComparacionPath.LIST, {

                signal: controller.signal,
            }),
            controller,
        };
    },
};

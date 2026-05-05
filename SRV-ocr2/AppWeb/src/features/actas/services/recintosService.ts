import { httpClient } from "../../../config/axios";
import { loadAbort } from "../../../config/load-abort";
import type { UseApiCall } from "../../../config/useApicall";
import { RecintosPath } from "../path-services/actas-path-service";
import type { Recinto } from "../models/response/recinto-response";

export const recintosService = {
    list: (): UseApiCall<Recinto[]> => {
        const controller = loadAbort();
        return {
            call: httpClient.get<Recinto[]>(RecintosPath.LIST, { signal: controller.signal }),
            controller,
        };
    },
};

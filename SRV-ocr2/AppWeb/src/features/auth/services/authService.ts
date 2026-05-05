import { httpClient } from "../../../config/axios";
import type { CredentialRequest } from "../models/request/credential-request";
import type { LoginResponse } from "../models/response/success-post-auth";
import { AuthPath } from "../path-services/auth-path-service";
import { loadAbort } from "../../../config/load-abort";
import type { UseApiCall } from "../../../config/useApicall";
import type { UserResponse } from "../models/response/success-user";
import type { NewUserDto } from "../models/request/new-user-dto";


export const authService = {
    login: (payload: CredentialRequest): UseApiCall<LoginResponse> => {
        const controller = loadAbort();
        return {
            call: httpClient.post<LoginResponse>(AuthPath.LOGIN, payload, { signal: controller.signal }),
            controller
        }
    },
    me: (): UseApiCall<UserResponse> => {
        const controller = loadAbort();
        return {
            call: httpClient.get<UserResponse>(AuthPath.ME, { signal: controller.signal }),
            controller
        }
    },
    register: (payload: NewUserDto): UseApiCall<UserResponse> => {
        const controller = loadAbort();
        return {
            call: httpClient.post<UserResponse>(AuthPath.REGISTER, payload, { signal: controller.signal }),
            controller
        }
    },
};

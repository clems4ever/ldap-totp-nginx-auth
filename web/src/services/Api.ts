import { AxiosResponse } from "axios";

import { getBasePath } from "@utils/BasePath";

const basePath = getBasePath();

// Note: If you change this const you must also do so in the backend at internal/handlers/cost.go.
export const ConsentPath = basePath + "/api/oidc/consent";

export const FirstFactorPath = basePath + "/api/firstfactor";
export const InitiateTOTPRegistrationPath = basePath + "/api/secondfactor/totp/identity/start";
export const CompleteTOTPRegistrationPath = basePath + "/api/secondfactor/totp/identity/finish";

export const InitiateU2FRegistrationPath = basePath + "/api/secondfactor/u2f/identity/start";
export const CompleteU2FRegistrationStep1Path = basePath + "/api/secondfactor/u2f/identity/finish";
export const CompleteU2FRegistrationStep2Path = basePath + "/api/secondfactor/u2f/register";

export const InitiateU2FSignInPath = basePath + "/api/secondfactor/u2f/sign_request";
export const CompleteU2FSignInPath = basePath + "/api/secondfactor/u2f/sign";

export const InitiateDuoDeviceSelectionPath = basePath + "/api/secondfactor/duo_devices";
export const CompleteDuoDeviceSelectionPath = basePath + "/api/secondfactor/duo_device";

export const CompletePushNotificationSignInPath = basePath + "/api/secondfactor/duo";
export const CompleteTOTPSignInPath = basePath + "/api/secondfactor/totp";

export const InitiateResetPasswordPath = basePath + "/api/reset-password/identity/start";
export const CompleteResetPasswordPath = basePath + "/api/reset-password/identity/finish";
// Do the password reset during completion.
export const ResetPasswordPath = basePath + "/api/reset-password";

export const LogoutPath = basePath + "/api/logout";
export const StatePath = basePath + "/api/state";
export const UserInfoPath = basePath + "/api/user/info";
export const UserInfo2FAMethodPath = basePath + "/api/user/info/2fa_method";

export const ConfigurationPath = basePath + "/api/configuration";

export interface ErrorResponse {
    status: "KO";
    message: string;
}

export interface Response<T> {
    status: "OK";
    data: T;
}

export type ServiceResponse<T> = Response<T> | ErrorResponse;

function toErrorResponse<T>(resp: AxiosResponse<ServiceResponse<T>>): ErrorResponse | undefined {
    if (resp.data && "status" in resp.data && resp.data["status"] === "KO") {
        return resp.data as ErrorResponse;
    }
    return undefined;
}

export function toData<T>(resp: AxiosResponse<ServiceResponse<T>>): T | undefined {
    if (resp.data && "status" in resp.data && resp.data["status"] === "OK") {
        return resp.data.data as T;
    }
    return undefined;
}

export function hasServiceError<T>(resp: AxiosResponse<ServiceResponse<T>>) {
    const errResp = toErrorResponse(resp);
    if (errResp && errResp.status === "KO") {
        return { errored: true, message: errResp.message };
    }
    return { errored: false, message: null };
}

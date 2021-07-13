import { SecondFactorMethod } from "@models/Methods";
import { UserInfo } from "@models/UserInfo";
import { UserInfoPath, UserInfo2FAMethodPath } from "@services/Api";
import { Get, PostWithOptionalResponse } from "@services/Client";

export type Method2FA = "u2f" | "totp" | "mobile_push";

export interface UserInfoPayload {
    display_name: string;
    method: Method2FA;
    has_u2f: boolean;
    has_totp: boolean;
    has_duo: boolean;
}

export interface MethodPreferencePayload {
    method: Method2FA;
}

export function toEnum(method: Method2FA): SecondFactorMethod {
    switch (method) {
        case "u2f":
            return SecondFactorMethod.U2F;
        case "totp":
            return SecondFactorMethod.TOTP;
        case "mobile_push":
            return SecondFactorMethod.MobilePush;
    }
}

export function toString(method: SecondFactorMethod): Method2FA {
    switch (method) {
        case SecondFactorMethod.U2F:
            return "u2f";
        case SecondFactorMethod.TOTP:
            return "totp";
        case SecondFactorMethod.MobilePush:
            return "mobile_push";
    }
}

export async function getUserPreferences(): Promise<UserInfo> {
    const res = await Get<UserInfoPayload>(UserInfoPath);
    return { ...res, method: toEnum(res.method) };
}

export function setPreferred2FAMethod(method: SecondFactorMethod) {
    return PostWithOptionalResponse(UserInfo2FAMethodPath, { method: toString(method) } as MethodPreferencePayload);
}

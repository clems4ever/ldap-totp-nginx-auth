import {
    CompletePushNotificationSignInPath,
    InitiateDuoDeviceSelectionPath,
    CompleteDuoDeviceSelectionPath,
} from "@services/Api";
import { Get, PostWithOptionalResponse } from "@services/Client";

interface CompleteU2FSigninBody {
    targetURL?: string;
}

export function completePushNotificationSignIn(targetURL: string | undefined) {
    const body: CompleteU2FSigninBody = {};
    if (targetURL) {
        body.targetURL = targetURL;
    }
    return PostWithOptionalResponse<DuoSignInResponse>(CompletePushNotificationSignInPath, body);
}

export interface DuoSignInResponse {
    result: string;
    devices: DuoDevice[];
    redirect: string;
    enroll_url: string;
}

export interface DuoDevicesGetResponse {
    result: string;
    devices: DuoDevice[];
    enroll_url: string;
}

export interface DuoDevice {
    device: string;
    display_name: string;
    capabilities: string[];
}
export async function initiateDuoDeviceSelectionProcess() {
    return Get<DuoDevicesGetResponse>(InitiateDuoDeviceSelectionPath);
}

export interface DuoDevicePostRequest {
    device: string;
    method: string;
}
export async function completeDuoDeviceSelectionProcess(device: DuoDevicePostRequest) {
    return PostWithOptionalResponse(CompleteDuoDeviceSelectionPath, { device: device.device, method: device.method });
}

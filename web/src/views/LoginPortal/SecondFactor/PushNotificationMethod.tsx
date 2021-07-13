import React, { useEffect, useCallback, useState, ReactNode } from "react";

import { Button, makeStyles } from "@material-ui/core";

import FailureIcon from "@components/FailureIcon";
import PushNotificationIcon from "@components/PushNotificationIcon";
import SuccessIcon from "@components/SuccessIcon";
import { useIsMountedRef } from "@hooks/Mounted";
import { useRedirectionURL } from "@hooks/RedirectionURL";
import {
    completePushNotificationSignIn,
    completeDuoDeviceSelectionProcess,
    DuoDevicePostRequest,
    initiateDuoDeviceSelectionProcess,
} from "@services/PushNotification";
import { AuthenticationLevel } from "@services/State";
import DeviceSelectionContainer, {
    SelectedDevice,
    SelectableDevice,
} from "@views/LoginPortal/SecondFactor/DeviceSelectionContainer";
import MethodContainer, { State as MethodContainerState } from "@views/LoginPortal/SecondFactor/MethodContainer";

export enum State {
    SignInInProgress = 1,
    Success = 2,
    Failure = 3,
    Selection = 4,
    Enroll = 5,
}

export interface Props {
    id: string;
    authenticationLevel: AuthenticationLevel;
    selected: boolean;

    onSignInError: (err: Error) => void;
    onSelectionClick: () => void;
    onSignInSuccess: (redirectURL: string | undefined) => void;
}

const PushNotificationMethod = function (props: Props) {
    const style = useStyles();
    const [state, setState] = useState(State.SignInInProgress);
    const redirectionURL = useRedirectionURL();
    const mounted = useIsMountedRef();
    const [enroll_url, setEnrollUrl] = useState("https://duo.com/");
    const [devices, setDevices] = useState([] as SelectableDevice[]);

    const { onSignInSuccess, onSignInError } = props;
    /* eslint-disable react-hooks/exhaustive-deps */
    const onSignInSuccessCallback = useCallback(onSignInSuccess, []);
    const onSignInErrorCallback = useCallback(onSignInError, []);
    /* eslint-enable react-hooks/exhaustive-deps */

    const fetchDuoDevicesFunc = useCallback(async () => {
        try {
            const res = await initiateDuoDeviceSelectionProcess();
            if (!mounted.current) return;
            switch (res.result) {
                case "auth":
                    var devices_temp = [] as SelectableDevice[];
                    res.devices.forEach((d: { device: any; display_name: any; capabilities: any }) =>
                        devices_temp.push({ id: d.device, name: d.display_name, methods: d.capabilities }),
                    );
                    setDevices(devices_temp);
                    setState(State.Selection);
                    break;
                case "allow":
                    onSignInErrorCallback(new Error("Device Selection is being bypassed by Duo Policy"));
                    break;
                case "deny":
                    onSignInErrorCallback(new Error("Device Selection was denied by Duo Policy"));
                    break;
                case "enroll":
                    onSignInErrorCallback(new Error("No compatible device found"));
                    if (res.enroll_url) setEnrollUrl(res.enroll_url);
                    setState(State.Enroll);
                    break;
            }
        } catch (err) {
            if (!mounted.current) return;
            console.error(err);
            onSignInErrorCallback(new Error("There was an issue fetching Duo device(s)"));
        }
    }, [mounted, onSignInErrorCallback]);

    const signInFunc = useCallback(async () => {
        if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
            setState(State.Success);
            return;
        }

        try {
            setState(State.SignInInProgress);
            const res = await completePushNotificationSignIn(redirectionURL);
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;
            if (res && res.result === "auth") {
                if (res.devices.length < 1) {
                    fetchDuoDevicesFunc();
                    return;
                }
                var devices_temp = [] as SelectableDevice[];
                res.devices.forEach((d) =>
                    devices_temp.push({ id: d.device, name: d.display_name, methods: d.capabilities }),
                );
                setDevices(devices_temp);
                setState(State.Selection);
                return;
            }
            if (res && res.result === "enroll") {
                onSignInErrorCallback(new Error("No compatible device found"));
                if (res.enroll_url) setEnrollUrl(res.enroll_url);
                setState(State.Enroll);
                return;
            }
            setState(State.Success);
            setTimeout(() => {
                if (!mounted.current) return;
                onSignInSuccessCallback(res ? res.redirect : undefined);
            }, 1500);
        } catch (err) {
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current || state !== State.SignInInProgress) return;

            console.error(err);
            onSignInErrorCallback(new Error("There was an issue completing sign in process"));
            setState(State.Failure);
        }
    }, [
        props.authenticationLevel,
        redirectionURL,
        mounted,
        fetchDuoDevicesFunc,
        onSignInErrorCallback,
        onSignInSuccessCallback,
        state,
    ]);

    const updateDuoDevice = useCallback(
        async function (device: DuoDevicePostRequest) {
            try {
                await completeDuoDeviceSelectionProcess(device);
                if (!props.selected) {
                    setState(State.SignInInProgress);
                    props.onSelectionClick();
                } else {
                    setState(State.SignInInProgress);
                }
            } catch (err) {
                console.error(err);
                onSignInErrorCallback(new Error("There was an issue updating preferred Duo device"));
            }
        },
        [onSignInErrorCallback, props],
    );

    const handleDuoDeviceSelected = useCallback(
        (device: SelectedDevice) => {
            updateDuoDevice({ device: device.id, method: device.method });
        },
        [updateDuoDevice],
    );

    // Set successful state if user is already authenticated.
    useEffect(() => {
        if (props.authenticationLevel >= AuthenticationLevel.TwoFactor) {
            setState(State.Success);
        }
    }, [props.authenticationLevel, setState]);

    useEffect(() => {
        if (state === State.SignInInProgress) signInFunc();
    }, [signInFunc, state]);

    if (state === State.Selection)
        return (
            <DeviceSelectionContainer
                devices={devices}
                onBack={() => setState(State.SignInInProgress)}
                onSelect={handleDuoDeviceSelected}
            />
        );

    let icon: ReactNode;
    switch (state) {
        case State.SignInInProgress:
            icon = <PushNotificationIcon width={64} height={64} animated />;
            break;
        case State.Success:
            icon = <SuccessIcon />;
            break;
        case State.Failure:
            icon = <FailureIcon />;
    }

    let methodState = MethodContainerState.METHOD;
    if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
        methodState = MethodContainerState.ALREADY_AUTHENTICATED;
    } else if (state === State.Enroll) {
        // TODO: add Duo Enrollment
        methodState = MethodContainerState.NOT_REGISTERED;
    } else if (!props.selected) {
        methodState = MethodContainerState.NOT_SELECTED;
    }
    return (
        <MethodContainer
            id={props.id}
            title="Push Notification"
            explanation="A notification has been sent to your smartphone"
            registered={props.selected}
            state={methodState}
            onSelectClick={fetchDuoDevicesFunc}
            // TODO: add Duo Enrollment
            onRegisterClick={() => window.open(enroll_url, "_blank")}
        >
            <div className={style.icon}>{icon}</div>
            <div className={state !== State.Failure ? "hidden" : ""}>
                <Button color="secondary" onClick={signInFunc}>
                    Retry
                </Button>
            </div>
        </MethodContainer>
    );
};

export default PushNotificationMethod;

const useStyles = makeStyles(() => ({
    icon: {
        width: "64px",
        height: "64px",
        display: "inline-block",
    },
}));

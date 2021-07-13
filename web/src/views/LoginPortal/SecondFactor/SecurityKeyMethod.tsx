import React, { useCallback, useEffect, useState, Fragment } from "react";

import { makeStyles, Button, useTheme } from "@material-ui/core";
import { CSSProperties } from "@material-ui/styles";
import u2fApi from "u2f-api";

import FailureIcon from "@components/FailureIcon";
import FingerTouchIcon from "@components/FingerTouchIcon";
import LinearProgressBar from "@components/LinearProgressBar";
import { useIsMountedRef } from "@hooks/Mounted";
import { useRedirectionURL } from "@hooks/RedirectionURL";
import { useTimer } from "@hooks/Timer";
import { initiateU2FSignin, completeU2FSignin } from "@services/SecurityKey";
import { AuthenticationLevel } from "@services/State";
import IconWithContext from "@views/LoginPortal/SecondFactor/IconWithContext";
import MethodContainer, { State as MethodContainerState } from "@views/LoginPortal/SecondFactor/MethodContainer";

export enum State {
    WaitTouch = 1,
    SigninInProgress = 2,
    Failure = 3,
}

export interface Props {
    id: string;
    authenticationLevel: AuthenticationLevel;
    registered: boolean;

    onRegisterClick: () => void;
    onSignInError: (err: Error) => void;
    onSignInSuccess: (redirectURL: string | undefined) => void;
}

const SecurityKeyMethod = function (props: Props) {
    const signInTimeout = 30;
    const [state, setState] = useState(State.WaitTouch);
    const style = useStyles();
    const redirectionURL = useRedirectionURL();
    const mounted = useIsMountedRef();
    const [timerPercent, triggerTimer] = useTimer(signInTimeout * 1000 - 500);

    const { onSignInSuccess, onSignInError } = props;
    /* eslint-disable react-hooks/exhaustive-deps */
    const onSignInErrorCallback = useCallback(onSignInError, []);
    const onSignInSuccessCallback = useCallback(onSignInSuccess, []);
    /* eslint-enable react-hooks/exhaustive-deps */

    const doInitiateSignIn = useCallback(async () => {
        // If user is already authenticated, we don't initiate sign in process.
        if (!props.registered || props.authenticationLevel === AuthenticationLevel.TwoFactor) {
            return;
        }

        try {
            triggerTimer();
            setState(State.WaitTouch);
            const signRequest = await initiateU2FSignin();
            const signRequests: u2fApi.SignRequest[] = [];
            for (var i in signRequest.registeredKeys) {
                const r = signRequest.registeredKeys[i];
                signRequests.push({
                    appId: signRequest.appId,
                    challenge: signRequest.challenge,
                    keyHandle: r.keyHandle,
                    version: r.version,
                });
            }
            const signResponse = await u2fApi.sign(signRequests, signInTimeout);
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;

            setState(State.SigninInProgress);
            const res = await completeU2FSignin(signResponse, redirectionURL);
            onSignInSuccessCallback(res ? res.redirect : undefined);
        } catch (err) {
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;
            console.error(err);
            onSignInErrorCallback(new Error("Failed to initiate security key sign in process"));
            setState(State.Failure);
        }
    }, [
        onSignInSuccessCallback,
        onSignInErrorCallback,
        redirectionURL,
        mounted,
        triggerTimer,
        props.authenticationLevel,
        props.registered,
    ]);

    useEffect(() => {
        doInitiateSignIn();
    }, [doInitiateSignIn]);

    let methodState = MethodContainerState.METHOD;
    if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
        methodState = MethodContainerState.ALREADY_AUTHENTICATED;
    } else if (!props.registered) {
        methodState = MethodContainerState.NOT_REGISTERED;
    }

    return (
        <MethodContainer
            id={props.id}
            title="Security Key"
            explanation="Touch the token of your security key"
            registered={props.registered}
            state={methodState}
            onRegisterClick={props.onRegisterClick}
        >
            <div className={style.icon}>
                <Icon state={state} timer={timerPercent} onRetryClick={doInitiateSignIn} />
            </div>
        </MethodContainer>
    );
};

export default SecurityKeyMethod;

const useStyles = makeStyles((theme) => ({
    icon: {
        display: "inline-block",
    },
}));

interface IconProps {
    state: State;

    timer: number;
    onRetryClick: () => void;
}

function Icon(props: IconProps) {
    const state = props.state as State;
    const theme = useTheme();

    const progressBarStyle: CSSProperties = {
        marginTop: theme.spacing(),
    };

    const touch = (
        <IconWithContext
            icon={<FingerTouchIcon size={64} animated strong />}
            context={<LinearProgressBar value={props.timer} style={progressBarStyle} height={theme.spacing(2)} />}
            className={state === State.WaitTouch ? undefined : "hidden"}
        />
    );

    const failure = (
        <IconWithContext
            icon={<FailureIcon />}
            context={
                <Button color="secondary" onClick={props.onRetryClick}>
                    Retry
                </Button>
            }
            className={state === State.Failure ? undefined : "hidden"}
        />
    );

    return (
        <Fragment>
            {touch}
            {failure}
        </Fragment>
    );
}

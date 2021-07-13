import React, { Fragment, ReactNode, useCallback, useEffect, useState } from "react";

import { Redirect, Route, Switch, useHistory, useLocation } from "react-router";

import {
    AuthenticatedRoute,
    FirstFactorRoute,
    SecondFactorPushRoute,
    SecondFactorRoute,
    SecondFactorTOTPRoute,
    SecondFactorU2FRoute,
} from "@constants/Routes";
import { useConfiguration } from "@hooks/Configuration";
import { useNotifications } from "@hooks/NotificationsContext";
import { useRedirectionURL } from "@hooks/RedirectionURL";
import { useRedirector } from "@hooks/Redirector";
import { useRequestMethod } from "@hooks/RequestMethod";
import { useAutheliaState } from "@hooks/State";
import { useUserPreferences as userUserInfo } from "@hooks/UserInfo";
import { SecondFactorMethod } from "@models/Methods";
import { checkSafeRedirection } from "@services/SafeRedirection";
import { AuthenticationLevel } from "@services/State";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import AuthenticatedView from "@views/LoginPortal/AuthenticatedView/AuthenticatedView";
import FirstFactorForm from "@views/LoginPortal/FirstFactor/FirstFactorForm";
import SecondFactorForm from "@views/LoginPortal/SecondFactor/SecondFactorForm";

export interface Props {
    rememberMe: boolean;
    resetPassword: boolean;
}

const RedirectionErrorMessage =
    "There was an issue redirecting the user. Check that the redirection URI matches the domain.";

const LoginPortal = function (props: Props) {
    const history = useHistory();
    const location = useLocation();
    const redirectionURL = useRedirectionURL();
    const requestMethod = useRequestMethod();
    const { createErrorNotification } = useNotifications();
    const [firstFactorDisabled, setFirstFactorDisabled] = useState(true);
    const redirector = useRedirector();

    const [state, fetchState, , fetchStateError] = useAutheliaState();
    const [userInfo, fetchUserInfo, , fetchUserInfoError] = userUserInfo();
    const [configuration, fetchConfiguration, , fetchConfigurationError] = useConfiguration();

    const redirect = useCallback((url: string) => history.push(url), [history]);

    // Fetch the state when portal is mounted.
    useEffect(() => {
        fetchState();
    }, [fetchState]);

    // Fetch preferences and configuration when user is authenticated.
    useEffect(() => {
        if (state && state.authentication_level >= AuthenticationLevel.OneFactor) {
            fetchUserInfo();
            fetchConfiguration();
        }
    }, [state, fetchUserInfo, fetchConfiguration]);

    // Enable first factor when user is unauthenticated.
    useEffect(() => {
        if (state && state.authentication_level > AuthenticationLevel.Unauthenticated) {
            setFirstFactorDisabled(true);
        }
    }, [state, setFirstFactorDisabled]);

    // Display an error when state fetching fails
    useEffect(() => {
        if (fetchStateError) {
            createErrorNotification("There was an issue fetching the current user state");
        }
    }, [fetchStateError, createErrorNotification]);

    // Display an error when configuration fetching fails
    useEffect(() => {
        if (fetchConfigurationError) {
            createErrorNotification("There was an issue retrieving global configuration");
        }
    }, [fetchConfigurationError, createErrorNotification]);

    // Display an error when preferences fetching fails
    useEffect(() => {
        if (fetchUserInfoError) {
            createErrorNotification("There was an issue retrieving user preferences");
        }
    }, [fetchUserInfoError, createErrorNotification]);

    // Redirect to the correct stage if not enough authenticated
    useEffect(() => {
        (async function () {
            if (!state) {
                return;
            }

            if (
                redirectionURL &&
                ((configuration &&
                    !configuration.second_factor_enabled &&
                    state.authentication_level >= AuthenticationLevel.OneFactor) ||
                    state.authentication_level === AuthenticationLevel.TwoFactor)
            ) {
                try {
                    const res = await checkSafeRedirection(redirectionURL);
                    if (res && res.ok) {
                        redirector(redirectionURL);
                    } else {
                        createErrorNotification(RedirectionErrorMessage);
                    }
                } catch (err) {
                    createErrorNotification(RedirectionErrorMessage);
                }
                return;
            }

            const redirectionSuffix = redirectionURL
                ? `?rd=${encodeURIComponent(redirectionURL)}${requestMethod ? `&rm=${requestMethod}` : ""}`
                : "";

            if (state.authentication_level === AuthenticationLevel.Unauthenticated) {
                setFirstFactorDisabled(false);
                redirect(`${FirstFactorRoute}${redirectionSuffix}`);
            } else if (state.authentication_level >= AuthenticationLevel.OneFactor && userInfo && configuration) {
                if (!configuration.second_factor_enabled) {
                    redirect(AuthenticatedRoute);
                } else {
                    if (userInfo.method === SecondFactorMethod.U2F) {
                        redirect(`${SecondFactorU2FRoute}${redirectionSuffix}`);
                    } else if (userInfo.method === SecondFactorMethod.MobilePush) {
                        redirect(`${SecondFactorPushRoute}${redirectionSuffix}`);
                    } else {
                        redirect(`${SecondFactorTOTPRoute}${redirectionSuffix}`);
                    }
                }
            }
        })();
    }, [
        state,
        redirectionURL,
        requestMethod,
        redirect,
        userInfo,
        setFirstFactorDisabled,
        configuration,
        createErrorNotification,
        redirector,
    ]);

    const handleAuthSuccess = async (redirectionURL: string | undefined) => {
        if (redirectionURL) {
            // Do an external redirection pushed by the server.
            redirector(redirectionURL);
        } else {
            // Refresh state
            fetchState();
        }
    };

    const firstFactorReady =
        state !== undefined &&
        state.authentication_level === AuthenticationLevel.Unauthenticated &&
        location.pathname === FirstFactorRoute;

    return (
        <Switch>
            <Route path={FirstFactorRoute} exact>
                <ComponentOrLoading ready={firstFactorReady}>
                    <FirstFactorForm
                        disabled={firstFactorDisabled}
                        rememberMe={props.rememberMe}
                        resetPassword={props.resetPassword}
                        onAuthenticationStart={() => setFirstFactorDisabled(true)}
                        onAuthenticationFailure={() => setFirstFactorDisabled(false)}
                        onAuthenticationSuccess={handleAuthSuccess}
                    />
                </ComponentOrLoading>
            </Route>
            <Route path={SecondFactorRoute}>
                {state && userInfo && configuration ? (
                    <SecondFactorForm
                        authenticationLevel={state.authentication_level}
                        userInfo={userInfo}
                        configuration={configuration}
                        onMethodChanged={() => fetchUserInfo()}
                        onAuthenticationSuccess={handleAuthSuccess}
                    />
                ) : null}
            </Route>
            <Route path={AuthenticatedRoute} exact>
                {userInfo ? <AuthenticatedView name={userInfo.display_name} /> : null}
            </Route>
            {/* By default we route to first factor page */}
            <Route path="/">
                <Redirect to={FirstFactorRoute} />
            </Route>
        </Switch>
    );
};

export default LoginPortal;

interface ComponentOrLoadingProps {
    ready: boolean;

    children: ReactNode;
}

function ComponentOrLoading(props: ComponentOrLoadingProps) {
    return (
        <Fragment>
            <div className={props.ready ? "hidden" : ""}>
                <LoadingPage />
            </div>
            {props.ready ? props.children : null}
        </Fragment>
    );
}

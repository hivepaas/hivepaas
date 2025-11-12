import { useState } from "react";

import { useSearchParams } from "react-router";
import { useCookie, useMount } from "react-use";

import { WarningAuthenticationIcon } from "@assets/icons";

import { AppLoader, AppNavigate } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";

import { useAuthContext } from "@application/authentication/context";
import { AuthCommands } from "@application/authentication/data/commands";
import { AuthQueries } from "@application/authentication/data/queries";
import { AuthenticationLayout } from "@application/authentication/layouts";

import { Button, Card, CardContent, CardDescription, CardTitle } from "@/components/ui";

import { SignInForm } from "../form";

interface State {
    type: "initial" | "lockout";
}

function ErrorMessage({ message }: { message: string }) {
    return <div className="flex items-center justify-center h-screen text-red-500">{message}</div>;
}

function View() {
    const [state, setState] = useState<State>({
        type: "initial",
    });

    // const { navigate } = useAppNavigate();

    const { setProfile } = useProfileContext();
    const { enable2FA, enableMfaSetup } = useAuthContext();

    const { data: { data: loginOptions } = { data: [] } } = AuthQueries.useGetLoginOptions();

    const { mutate: signIn, isPending } = AuthCommands.useSignIn({
        onSuccess: setProfile,
        on2FARequired: enable2FA,
        on2FASetupRequired: enableMfaSetup,
        onTooManyAttempts: () => {
            setState({
                type: "lockout",
            });
        },
    });

    let child = null;

    if (state.type === "initial") {
        child = (
            <SignInForm
                loginOptions={loginOptions}
                isPending={isPending}
                onSubmit={signIn}
            />
        );
    }

    if (state.type === "lockout") {
        child = (
            <Card>
                <CardContent className="flex flex-col items-center gap-4 text-center">
                    <div className="rounded-full bg-yellow-50 p-3 text-yellow-600">
                        <WarningAuthenticationIcon className="h-6 w-6" />
                    </div>
                    <CardTitle className="text-base">Your account has been temporarily blocked</CardTitle>
                    <CardDescription>
                        Due to several failed login attempts, we have blocked your account for 15 minutes. Please try
                        again later.
                    </CardDescription>
                    <Button
                        variant="outline"
                        type="button"
                        onClick={() => {
                            setState({
                                type: "initial",
                            });
                        }}
                    >
                        Try again
                    </Button>
                </CardContent>
            </Card>
        );
    }

    return (
        <AuthenticationLayout>
            <div>{child}</div>
        </AuthenticationLayout>
    );
}

export function SignInRoute() {
    const [token, , deleteToken] = useCookie("access_token");

    const { setProfile } = useProfileContext();

    const {
        mutate: signInSSO,
        isPending,
        error,
    } = AuthCommands.useSignInSSO({
        onSuccess: profile => {
            deleteToken();

            setProfile(profile);
        },
        onError: () => {
            deleteToken();
        },
    });

    useMount(() => {
        if (token) {
            signInSSO({
                token,
            });
        }
    });

    const { data } = useAuthContext();

    const [params] = useSearchParams();

    if ("required2FA" in data && data.required2FA) {
        return (
            <AppNavigate
                to={{
                    pathname: ROUTE.auth.twoFA.$route,
                    search: params.toString(),
                }}
                replace
            />
        );
    }

    if (token && isPending) {
        return <AppLoader />;
    }

    if (token && error) {
        return <ErrorMessage message={error.message} />;
    }

    return <View />;
}

import { useState } from "react";

import { useSearchParams } from "react-router";

import { ErrorAuthenticationIcon, SuccessAuthenticationIcon } from "@assets/icons";

import { AppNavigate } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";
import type { Profile } from "@application/shared/entities";

import { useAuthContext } from "@application/authentication/context";
import { AuthCommands } from "@application/authentication/data/commands";
import type { SignIn2FA } from "@application/authentication/domain";
import { AuthenticationLayout } from "@application/authentication/layouts";

import { Button, Card, CardContent, CardDescription, CardTitle } from "@/components/ui";

import { TwoFaForm } from "../form";

type State =
    | {
          type: "initial";
      }
    | {
          type: "success";
          profile: Profile;
      }
    | {
          type: "error";
      };

function View({ email, mfaToken }: ViewProps) {
    void email;
    void mfaToken;

    const [state, setState] = useState<State>({
        type: "initial",
    });

    const { setProfile } = useProfileContext();
    const { clear: clearAuthContext } = useAuthContext();

    const { mutate: signIn2FA, isPending: isSignIn2FAPending } = AuthCommands.useSignIn2FA({
        onSuccess: profile => {
            setProfile(profile);
        },
        onError: () => {
            setState({
                type: "error",
            });
        },
        onTooManyAttempts: clearAuthContext,
    });

    function handleSignIn2FA(values: Pick<SignIn2FA, "code">) {
        signIn2FA({
            code: values.code,
            mfaToken,
        });
    }

    let child = null;

    if (state.type === "initial") {
        child = (
            <TwoFaForm
                isPending={isSignIn2FAPending}
                onSubmit={handleSignIn2FA}
            />
        );
    }

    if (state.type === "error") {
        child = (
            <Card>
                <CardContent className="flex flex-col items-center gap-4 text-center">
                    <div className="rounded-full bg-red-50 p-3 text-red-600">
                        <ErrorAuthenticationIcon className="h-6 w-6" />
                    </div>
                    <CardTitle className="text-base">Your Authentication Failed</CardTitle>
                    <CardDescription>Check your code and try again.</CardDescription>
                    <Button
                        variant="outline"
                        type="button"
                        onClick={() => {
                            setState({
                                type: "initial",
                            });
                        }}
                    >
                        Try Again
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

interface ViewProps {
    email: string;
    mfaToken: string;
}

export function TwoFaRoute() {
    const { data } = useAuthContext();

    const [params] = useSearchParams();

    if (!("required2FA" in data) || !data.required2FA) {
        return (
            <AppNavigate.Basic
                to={{
                    pathname: ROUTE.auth.signIn.$route,
                    search: params.toString(),
                }}
                replace
            />
        );
    }

    return (
        <View
            email={data.email}
            mfaToken={data.mfaToken}
        />
    );
}

import { useState } from "react";

import { Button, Card, CardContent, CardDescription, CardTitle } from "@/components/ui";
import { useSearchParams } from "react-router";

import { ErrorAuthenticationIcon, SuccessAuthenticationIcon } from "@assets/icons";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import { AuthCommands } from "@application/authentication/data/commands";
import { AuthenticationLayout } from "@application/authentication/layouts";

import { ForgotPasswordForm } from "../form/forgot-password.form.com";

type State =
    | {
          type: "initial";
      }
    | {
          type: "success";
          linkExpirationMins: number;
      }
    | {
          type: "error";
      };

function View() {
    const [state, setState] = useState<State>({
        type: "initial",
    });

    const { navigate } = useAppNavigate();

    const { mutate: forgotPassword, isPending } = AuthCommands.useForgotPassword({
        onSuccess: res => {
            setState({
                type: "success",
                linkExpirationMins: res.data.linkExpirationMins,
            });
        },
        onError: () => {
            setState({
                type: "error",
            });
        },
    });

    let child = null;

    if (state.type === "initial") {
        child = (
            <ForgotPasswordForm
                isPending={isPending}
                onSubmit={values => {
                    forgotPassword(values);
                }}
            />
        );
    }

    if (state.type === "success") {
        child = (
            <Card>
                <CardContent className="flex flex-col items-center gap-4 text-center">
                    <div className="rounded-full bg-emerald-50 p-3 text-emerald-600">
                        <SuccessAuthenticationIcon className="h-6 w-6" />
                    </div>
                    <CardTitle className="text-base">Reset link sent</CardTitle>
                    <CardDescription>
                        We emailed you a reset link. It expires in {state.linkExpirationMins} minutes.
                    </CardDescription>
                    <Button
                        variant="outline"
                        type="button"
                        onClick={() => navigate(ROUTE.auth.signIn.$route)}
                    >
                        Back to sign in
                    </Button>
                </CardContent>
            </Card>
        );
    }

    if (state.type === "error") {
        child = (
            <Card>
                <CardContent className="flex flex-col items-center gap-4 text-center">
                    <div className="rounded-full bg-red-50 p-3 text-red-600">
                        <ErrorAuthenticationIcon className="h-6 w-6" />
                    </div>
                    <CardTitle className="text-base">Something went wrong</CardTitle>
                    <CardDescription>Please try again.</CardDescription>
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

export const ForgotPasswordRoute = () => {
    void useSearchParams;

    return <View />;
};



import { useState } from "react";

import { Card, CardContent, CardDescription, CardTitle } from "@/components/ui";
import { useSearchParams } from "react-router";

import { SuccessAuthenticationIcon } from "@assets/icons";

import { BackToSignIn } from "@application/authentication/components";
import { AuthCommands } from "@application/authentication/data/commands";
import type { ForgotPassword } from "@application/authentication/domain";
import { AuthenticationLayout } from "@application/authentication/layouts";

import { ForgotPasswordForm } from "../form/forgot-password.form.com";

interface State {
    type: "initial" | "success";
}
function View() {
    const [state, setState] = useState<State>({
        type: "initial",
    });

    const { mutate: forgotPassword, isPending } = AuthCommands.useForgotPassword({
        onSuccess: () => {
            setState({
                type: "success",
            });
        },
    });

    function handleSubmit(values: ForgotPassword) {
        forgotPassword(values);
    }

    let child = null;

    if (state.type === "initial") {
        child = (
            <ForgotPasswordForm
                isPending={isPending}
                onSubmit={handleSubmit}
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
                    <CardDescription>We have sent password reset instructions to your email</CardDescription>
                    <BackToSignIn />
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

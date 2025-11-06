import { useState } from "react";

import { useSearchParams } from "react-router";

import { AppNavigate } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";

import { useAuthContext } from "@application/authentication/context";
import { AuthQueries } from "@application/authentication/data/queries";

import { SignInForm } from "../form";
import type { SignInSchemaOutput } from "../schemas";

function View() {
    const [isPending, setIsPending] = useState(false);
    const profileContext = useProfileContext();

    const { data, isLoading } = AuthQueries.useGetLoginOptions();

    console.log(isLoading);

    console.log(data);

    async function onSubmit(values: SignInSchemaOutput) {
        setIsPending(true);
        await new Promise(resolve => setTimeout(resolve, 700));

        const VALID_EMAIL = "test@example.com";
        const VALID_PASSWORD = "123456";

        if (values.email === VALID_EMAIL && values.password === VALID_PASSWORD) {
            profileContext.setToken("fake-jwt-token-abc123");
            profileContext.setProfile({
                id: "user_1",
                firstName: "Test",
                lastName: "User",
                fullName: "Test User",
                photo: null,
                email: "test@example.com",
            });
            setIsPending(false);
            return;
        }

        // Simple feedback for invalid credentials
        alert("Email or password is incorrect.");
        setIsPending(false);
    }
    return (
        <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
            <div className="w-full max-w-sm">
                <SignInForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                />
            </div>
        </div>
    );
}

export const SignInRoute = () => {
    const { data } = useAuthContext();

    const [params] = useSearchParams();

    if (data.required2FA) {
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

    return <View />;
};

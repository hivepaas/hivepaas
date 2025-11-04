import { useState } from "react";

import { useProfileContext } from "@application/shared/context";
import { Profile } from "@application/shared/entities";

import { SignInForm } from "../form";
import type { SignInSchemaOutput } from "../schemas";

function View() {
    const [isPending, setIsPending] = useState(false);
    const profileContext = useProfileContext();

    async function onSubmit(values: SignInSchemaOutput) {
        setIsPending(true);
        await new Promise(resolve => setTimeout(resolve, 700));

        const VALID_EMAIL = "test@example.com";
        const VALID_PASSWORD = "123456";

        if (values.email === VALID_EMAIL && values.password === VALID_PASSWORD) {
            profileContext.setToken("fake-jwt-token-abc123");
            profileContext.setProfile(new Profile({ id: "user_1", username: "Test User" }));
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
    return <View />;
};

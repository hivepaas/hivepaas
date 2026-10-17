import { useCookie, useMount } from "react-use";

import { AppLoader, AppNavigate } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";

import { useAuthContext } from "@application/authentication/context";
import { AuthCommands } from "@application/authentication/data/commands";

function ErrorMessage({ message }: { message: string }) {
    return <div className="flex items-center justify-center h-screen text-red-500">{message}</div>;
}

export function SsoRoute() {
    const [token, , deleteToken] = useCookie("access_token");
    // A user with an authenticator app still owes the passcode: the server hands
    // over the token to sign in with it instead of a session.
    const [mfaToken, , deleteMfaToken] = useCookie("mfa_token");

    const { setProfile } = useProfileContext();
    const { data, enable2FA } = useAuthContext();

    const {
        mutate: signInSSO,
        isPending,
        error,
    } = AuthCommands.useSignInSSO({
        onSuccess: profile => {
            deleteToken();

            setProfile(profile);
        },
    });

    useMount(() => {
        if (mfaToken) {
            deleteMfaToken();
            enable2FA({ email: "", mfaToken });
            return;
        }

        if (!token) return;

        signInSSO({
            token,
        });
    });

    if ("required2FA" in data && data.required2FA) {
        return (
            <AppNavigate.Basic
                to={ROUTE.auth.twoFA.$route}
                replace
            />
        );
    }

    if (isPending) {
        return <AppLoader />;
    }

    if (!token) {
        return <ErrorMessage message="Access token not found" />;
    }

    if (error) {
        return <ErrorMessage message={error.message} />;
    }

    return null;
}

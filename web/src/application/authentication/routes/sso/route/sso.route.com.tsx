import { useCookie, useMount } from "react-use";

import { AppLoader } from "@application/shared/components";
import { useProfileContext } from "@application/shared/context";

import { AuthCommands } from "@application/authentication/data/commands";

function ErrorMessage({ message }: { message: string }) {
    return <div className="flex items-center justify-center h-screen text-red-500">{message}</div>;
}

export function SsoRoute() {
    const [token, , deleteToken] = useCookie("access_token");

    console.log("token", token);

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
    });

    useMount(() => {
        if (!token) return;

        signInSSO({
            token,
        });
    });

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

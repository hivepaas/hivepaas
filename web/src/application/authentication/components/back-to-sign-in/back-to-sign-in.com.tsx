import { useSearchParams } from "react-router";

import { ArrowBackIcon } from "@assets/icons";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { useAuthContext } from "@application/authentication/context";

export function BackToSignIn() {
    const { clear } = useAuthContext();

    const [params] = useSearchParams();

    return (
        <AppLink
            to={{
                pathname: ROUTE.auth.signIn.$route,
                search: params.toString(),
            }}
            onClick={clear}
            replace
            className="flex items-center gap-2 text-sm text-primary hover:text-gray-700"
        >
            <ArrowBackIcon />
            Back to Sign In
        </AppLink>
    );
}

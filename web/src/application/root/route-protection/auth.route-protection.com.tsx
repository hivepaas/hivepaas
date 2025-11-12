import { type PropsWithChildren } from "react";

import { createSearchParams, useLocation, useMatch, useSearchParams } from "react-router";
import { useCookie } from "react-use";

import { AppNavigate } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";

import { AuthCommands } from "@application/authentication/data/commands";
import { SsoRoute } from "@application/authentication/routes";

export function AuthRouteProtection({ children }: PropsWithChildren) {
    const { profile } = useProfileContext();
    const [token, , deleteToken] = useCookie("access_token");

    const { setProfile } = useProfileContext();

    const { mutate: signInSSO } = AuthCommands.useSignInSSO({
        onSuccess: profileData => {
            deleteToken();

            setProfile(profileData);
        },
    });

    const location = useLocation();

    const [params] = useSearchParams();

    const isMain = useMatch("/") !== null;
    const isAuthGroup = useMatch("auth/*") !== null;

    const nextPath = params.get("next");

    /**
     * If the next path is "auth/sso/success", redirect to the sso success page
     */
    if (nextPath === "auth/sso/success") {
        if (token) {
            signInSSO({
                token,
            });
        }
        return <SsoRoute />;
    }

    if (profile && isAuthGroup && params.has("next")) {
        // TODO get page to redirect from query, "?next=/modules/projects/dashboard/" for example

        return (
            <AppNavigate
                to={params.get("next")!}
                replace
            />
        );
    }

    if (profile && (isMain || isAuthGroup)) {
        return (
            <AppNavigate
                to={ROUTE.modules.usersAndRoles.users.$route}
                replace
            />
        );
    }

    if (!profile && !isAuthGroup) {
        return (
            <AppNavigate
                to={{
                    pathname: ROUTE.auth.signIn.$route,
                    search: createSearchParams({ next: location.pathname }).toString(),
                }}
                replace
            />
        );
    }

    return children;
}

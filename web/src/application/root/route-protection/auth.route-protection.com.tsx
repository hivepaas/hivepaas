import { type PropsWithChildren } from "react";

import { useLocation, useMatch, useSearchParams } from "react-router";
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
            <AppNavigate.Basic
                to={params.get("next")!}
                replace
            />
        );
    }

    if (profile && (isMain || isAuthGroup)) {
        if (params.has("next")) {
            return (
                <AppNavigate.Basic
                    to={params.get("next")!}
                    replace
                />
            );
        } else {
            return (
                <AppNavigate.Basic
                    to={ROUTE.userManagement.users.$route}
                    replace
                />
            );
        }
    }

    if (!profile && !isAuthGroup) {
        const currentFullPath = nextPath ?? `${location.pathname}${location.search}`;
        const [pathPart = "", searchPart = ""] = currentFullPath.split("?");

        // Map route patterns to their destinations
        const authRouteMap: Record<string, string> = {
            "sign-up": ROUTE.auth.signUp.$route,
            "reset-password": ROUTE.auth.resetPassword.$route,
            "forgot-password": ROUTE.auth.forgotPassword.$route,
        };

        const redirectTo =
            Object.entries(authRouteMap).find(([pattern]) => pathPart.includes(pattern))?.[1] ??
            ROUTE.auth.signIn.$route;

        return (
            <AppNavigate.Basic
                to={{
                    pathname: redirectTo,
                    search: searchPart,
                }}
                replace
            />
        );
    }

    return children;
}

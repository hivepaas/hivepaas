import { type PropsWithChildren } from "react";

import { createSearchParams, useLocation, useMatch, useSearchParams } from "react-router";

import { AppNavigate } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";

export function AuthRouteProtection({ children }: PropsWithChildren) {
    const { profile } = useProfileContext();

    const location = useLocation();

    const [params] = useSearchParams();

    const isMain = useMatch("/") !== null;
    const isAuthGroup = useMatch("auth/*") !== null;

    if (params.has("next")) {
        return (
            <AppNavigate
                to={params.get("next")!}
                replace
            />
        );
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

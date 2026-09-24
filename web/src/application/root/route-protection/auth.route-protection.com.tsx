import { type PropsWithChildren } from "react";

import { useLocation, useMatch, useSearchParams } from "react-router";

import { AppNavigate } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";

function normalizePathname(pathname: string) {
    const withLeadingSlash = pathname.startsWith("/") ? pathname : `/${pathname}`;

    return withLeadingSlash.replace(/\/+$/, "") || "/";
}

function isAuthPath(pathname: string) {
    const normalizedPathname = normalizePathname(pathname);

    return normalizedPathname === "/auth" || normalizedPathname.startsWith("/auth/");
}

function parseSameOriginUrl(path: string | null) {
    const trimmedPath = path?.trim();

    if (!trimmedPath) {
        return null;
    }

    try {
        const url = new URL(trimmedPath, `${window.location.origin}/`);

        if (url.origin !== window.location.origin) {
            return null;
        }

        return url;
    } catch {
        return null;
    }
}

function getSafeNextPath(path: string | null) {
    const url = parseSameOriginUrl(path);

    if (url === null || isAuthPath(url.pathname)) {
        return null;
    }

    return `${url.pathname}${url.search}${url.hash}`;
}

function isSsoSuccessPath(path: string | null) {
    const url = parseSameOriginUrl(path);

    return normalizePathname(url?.pathname ?? "") === normalizePathname(ROUTE.auth.sso.success.$route);
}

function getAuthHandoff(path: string | null, pageParams: URLSearchParams) {
    const url = parseSameOriginUrl(path);

    if (url === null) {
        return null;
    }

    const pathname = normalizePathname(url.pathname);

    const authRouteMap: Record<string, string> = {
        "auth/sign-up": ROUTE.auth.signUp.$route,
        "auth/reset-password": ROUTE.auth.resetPassword.$route,
        "auth/forgot-password": ROUTE.auth.forgotPassword.$route,
    };

    const redirectTo = Object.entries(authRouteMap).find(([pattern]) => pathname.includes(pattern))?.[1];

    if (redirectTo === undefined) {
        return null;
    }

    // Merge sibling page-level params (e.g. `token`) that belong to the handoff
    // route but ended up at the top level because the `next` value was not
    // URL-encoded by the server (e.g. /?next=auth/reset-password?userID=X&token=Y).
    const mergedSearch = new URLSearchParams(url.searchParams);

    for (const [key, value] of pageParams.entries()) {
        if (key !== "next" && !mergedSearch.has(key)) {
            mergedSearch.set(key, value);
        }
    }

    return {
        pathname: redirectTo,
        search: mergedSearch.toString(),
    };
}

function getCurrentPath(location: ReturnType<typeof useLocation>) {
    return `${location.pathname}${location.search}${location.hash}`;
}

export function AuthRouteProtection({ children }: PropsWithChildren) {
    const { profile } = useProfileContext();

    const location = useLocation();

    const [params] = useSearchParams();

    const isMain = useMatch("/") !== null;
    const isAuthGroup = useMatch("auth/*") !== null;

    const nextPath = params.get("next");
    const safeNextPath = getSafeNextPath(nextPath);

    if (profile && (isMain || isAuthGroup)) {
        return (
            <AppNavigate.Basic
                // Every signed-in user has a home page, showing each only their part.
                to={safeNextPath ?? ROUTE.home.$route}
                replace
            />
        );
    }

    if (!profile && !isAuthGroup) {
        if (isMain && isSsoSuccessPath(nextPath)) {
            return (
                <AppNavigate.Basic
                    to={ROUTE.auth.sso.success.$route}
                    replace
                />
            );
        }

        const authHandoff = isMain ? getAuthHandoff(nextPath, params) : null;

        return (
            <AppNavigate.Basic
                to={
                    authHandoff ?? {
                        pathname: ROUTE.auth.signIn.$route,
                        search: isMain ? "" : new URLSearchParams({ next: getCurrentPath(location) }).toString(),
                    }
                }
                replace
            />
        );
    }

    return children;
}

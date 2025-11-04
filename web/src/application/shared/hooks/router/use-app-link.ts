import { useCallback } from "react";

import { type To, matchPath } from "react-router-dom";

// import { useI18n } from "@i18n/hooks";

function createHook() {
    function isNotTranslatable(pathname?: string) {
        if (!pathname) {
            return false;
        }

        return matchPath("auth/*", pathname) !== null;
    }

    return function useAppLink() {
        // const { language } = useI18n();

        const linkTo = useCallback((to: To): To => {
            const prefix = "";

            if (typeof to === "string") {
                if (isNotTranslatable(to)) {
                    return to;
                }

                return prefix + to;
            }

            if (isNotTranslatable(to.pathname)) {
                return to;
            }

            return {
                ...to,
                pathname: prefix + (to.pathname ?? ""),
            };
        }, []);

        return {
            linkTo,
        };
    };
}

export const useAppLink = createHook();

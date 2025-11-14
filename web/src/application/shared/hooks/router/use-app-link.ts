import { useCallback, useMemo } from "react";

import { type Path, type To } from "react-router";

function createHook() {
    function isString(to: To): to is string {
        return typeof to === "string";
    }

    function isPath(to: To): to is Partial<Path> & object {
        return typeof to === "object";
    }

    return function useAppLink() {
        /**
         * Generate link to modules.
         */
        const modules = useCallback(<T extends To>(to: T): T => {
            const prefix = "/modules";

            if (isString(to)) {
                return (prefix + to) as T;
            }

            if (isPath(to)) {
                return {
                    ...to,
                    pathname: prefix + (to.pathname ?? ""),
                };
            }

            throw new Error("Invalid to argument");
        }, []);

        const memoizedLink = useMemo(
            () => ({
                modules,
            }),
            [modules],
        );

        return {
            link: memoizedLink,
        };
    };
}

export const useAppLink = createHook();

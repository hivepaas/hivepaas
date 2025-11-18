import { useCallback, useMemo } from "react";

import { matchPath, useLocation } from "react-router";

interface MatchParams {
    pattern: string;
}

function createHook() {
    return function useAppMatchPath() {
        const location = useLocation();

        /**
         * Match all paths.
         */
        const all = useCallback(
            (params: MatchParams) => {
                return matchPath(
                    {
                        path: params.pattern,
                        caseSensitive: false,
                        end: false,
                    },
                    location.pathname,
                );
            },
            [location],
        );

        /**
         * Match the path with modules.
         */
        const modules = useCallback(
            (params: MatchParams) => {
                return matchPath(
                    {
                        path: `/${params.pattern}`,
                        caseSensitive: false,
                        end: false,
                    },
                    location.pathname,
                );
            },
            [location],
        );

        const memoizedMatch = useMemo(
            () => ({
                all,
                modules,
            }),
            [all, modules],
        );

        return {
            match: memoizedMatch,
        };
    };
}

export const useAppMatchPath = createHook();

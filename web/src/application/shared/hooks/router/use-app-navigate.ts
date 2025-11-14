import { useCallback, useMemo } from "react";

import { type NavigateOptions, type To, useLocation, useNavigate } from "react-router";

import { useAppLink } from "./use-app-link";

type Options = NavigateOptions & {
    ignorePrevPath?: boolean;
};

function createHook() {
    return function useAppNavigate() {
        const { link } = useAppLink();

        const location = useLocation();
        const navigate = useNavigate();

        /**
         * Navigate to a modules route.
         */
        const modules = useCallback(
            (to: To, { ignorePrevPath, ...options }: Options = {}) => {
                void navigate(link.modules(to), {
                    ...options,
                    state: ignorePrevPath
                        ? (location.state as unknown)
                        : {
                              from: location.pathname + location.search,
                          },
                });
            },
            [link, navigate, location],
        );

        const memoizedNavigate = useMemo(
            () => ({
                basic: async (to: To, { ignorePrevPath, ...options }: Options = {}) => {
                    return navigate(to, {
                        ...options,
                        state: ignorePrevPath
                            ? (location.state as unknown)
                            : {
                                  from: location.pathname + location.search,
                              },
                    });
                },
                modules,
            }),
            [navigate, modules, location],
        );

        return {
            navigate: memoizedNavigate,
        };
    };
}

export const useAppNavigate = createHook();

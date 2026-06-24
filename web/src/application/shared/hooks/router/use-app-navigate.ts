import { useCallback, useMemo } from "react";

import { type NavigateOptions, type To, useLocation, useNavigate } from "react-router";

import { useAppLink } from "./use-app-link";

type Options = Omit<NavigateOptions, "state"> & {
    ignorePrevPath?: boolean;
    state?: unknown;
};

function isPlainStateRecord(value: unknown): value is Record<string, unknown> {
    return typeof value === "object" && value !== null && !Array.isArray(value);
}

function getNavigateState(
    options: Pick<Options, "state">,
    locationState: unknown,
    locationPath: string,
    ignorePrevPath?: boolean,
): unknown {
    const nextState = options.state;

    if (ignorePrevPath) {
        return nextState ?? locationState;
    }

    const fromState = {
        from: locationPath,
    };

    if (nextState === undefined) {
        return fromState;
    }

    if (isPlainStateRecord(nextState)) {
        return {
            ...fromState,
            ...nextState,
        };
    }

    return {
        ...fromState,
        value: nextState,
    };
}

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
                    state: getNavigateState(
                        options,
                        location.state as unknown,
                        location.pathname + location.search,
                        ignorePrevPath,
                    ),
                });
            },
            [link, navigate, location],
        );

        const memoizedNavigate = useMemo(
            () => ({
                basic: async (to: To, { ignorePrevPath, ...options }: Options = {}) => {
                    return navigate(to, {
                        ...options,
                        state: getNavigateState(
                            options,
                            location.state as unknown,
                            location.pathname + location.search,
                            ignorePrevPath,
                        ),
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

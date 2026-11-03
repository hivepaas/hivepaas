import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { HomeApiContext } from "~/home/api/api-context";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useHomeAttentionApi() {
        const { api } = use(HomeApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findAll: async (signal?: AbortSignal) => {
                    const result = await api.home.attention.findAll({ data: {} }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get what needs attention", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return { queries };
    };
}

export const useHomeAttentionApi = createHook();

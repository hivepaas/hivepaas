import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { HomeApiContext } from "~/home/api/api-context";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useGetStartedApi() {
        const { api } = use(HomeApiContext);
        const { notifyError } = useApiErrorNotifications();

        const mutations = useMemo(
            () => ({
                requestDashboardCert: async () => {
                    const result = await api.home.getStarted.requestDashboardCert({ data: {} });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to ask for the dashboard's certificate", error });
                            throw error;
                        },
                    });
                },
                dismiss: async () => {
                    const result = await api.home.getStarted.dismiss({ data: {} });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to close Get started", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return { mutations };
    };
}

export const useGetStartedApi = createHook();

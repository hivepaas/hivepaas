import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { ApplicationApiContext } from "@application/shared/api/api-context";
import type { SupportFeedbacks_CreateOne_Req } from "@application/shared/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSupportFeedbacksApi() {
        const { api } = use(ApplicationApiContext);
        const { notifyError } = useApiErrorNotifications();

        const mutations = useMemo(
            () => ({
                createOne: async (data: SupportFeedbacks_CreateOne_Req["data"]) => {
                    const result = await api.support.feedbacks.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to submit feedback",
                                error,
                            });

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

export const useSupportFeedbacksApi = createHook();

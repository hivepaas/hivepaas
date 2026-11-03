import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useGetStartedApi } from "~/home/api";
import type { GetStarted_Dismiss_Res, GetStarted_RequestDashboardCert_Res } from "~/home/api/services";
import { QK as HOME_QK } from "~/home/data/constants";

import { QK } from "@application/shared/data/constants";

/**
 * Asking answers with where the certificate stands now - being obtained - which
 * is put in place of the last reading, so the card polls from there.
 */
function useRequestDashboardCert({
    onSuccess,
    ...options
}: Omit<UseMutationOptions<GetStarted_RequestDashboardCert_Res>, "mutationFn"> = {}) {
    const { mutations } = useGetStartedApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.requestDashboardCert,
        onSuccess: (response, ...rest) => {
            queryClient.setQueryData([HOME_QK["home.get-started.dashboard-cert"]], response);
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

/** Closing clears the step, which the profile says: it is read again. */
function useDismiss({ onSuccess, ...options }: Omit<UseMutationOptions<GetStarted_Dismiss_Res>, "mutationFn"> = {}) {
    const { mutations } = useGetStartedApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.dismiss,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["session.get-profile"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const GetStartedCommands = Object.freeze({
    useRequestDashboardCert,
    useDismiss,
});

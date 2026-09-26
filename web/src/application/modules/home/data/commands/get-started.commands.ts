import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useGetStartedApi } from "~/home/api";
import type { GetStarted_Dismiss_Res, GetStarted_RequestDashboardCert_Res } from "~/home/api/services";

import { QK } from "@application/shared/data/constants";

/**
 * The checklist comes with the profile, so after either call the profile is
 * read again: it has the certificate being obtained, or no checklist at all.
 */
function useRefreshProfile() {
    const queryClient = useQueryClient();

    return () => queryClient.invalidateQueries({ queryKey: [QK["session.get-profile"]] });
}

function useRequestDashboardCert({
    onSuccess,
    ...options
}: Omit<UseMutationOptions<GetStarted_RequestDashboardCert_Res>, "mutationFn"> = {}) {
    const { mutations } = useGetStartedApi();
    const refreshProfile = useRefreshProfile();

    return useMutation({
        mutationFn: mutations.requestDashboardCert,
        onSuccess: (response, ...rest) => {
            void refreshProfile();
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

function useDismiss({ onSuccess, ...options }: Omit<UseMutationOptions<GetStarted_Dismiss_Res>, "mutationFn"> = {}) {
    const { mutations } = useGetStartedApi();
    const refreshProfile = useRefreshProfile();

    return useMutation({
        mutationFn: mutations.dismiss,
        onSuccess: (response, ...rest) => {
            void refreshProfile();
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const GetStartedCommands = Object.freeze({
    useRequestDashboardCert,
    useDismiss,
});

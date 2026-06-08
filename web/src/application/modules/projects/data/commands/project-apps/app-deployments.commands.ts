import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAppDeploymentsApi } from "~/projects/api/hooks/project-apps";
import type { AppDeployments_Cancel_Req, AppDeployments_Cancel_Res } from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

import { invalidateSingleAppSummaryQueries } from "./app-configuration-cache.helpers";

type CancelReq = AppDeployments_Cancel_Req["data"];
type CancelRes = AppDeployments_Cancel_Res;
type CancelOptions = Omit<UseMutationOptions<CancelRes, Error, CancelReq>, "mutationFn">;

function useCancel({ onSuccess, ...options }: CancelOptions = {}) {
    const { mutations } = useAppDeploymentsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.cancel,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.deployments.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.deployments.$.find-one-by-id"]],
            });
            invalidateSingleAppSummaryQueries(queryClient, {
                projectID: request.projectID,
                appID: request.appID,
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppDeploymentsCommands = Object.freeze({
    useCancel,
});

import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppAutoscaleApi } from "../../../api/hooks/project-apps";
import type { AppAutoscale_UpdateOne_Req, AppAutoscale_UpdateOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

import { invalidateSingleAppConfigurationQueries } from "./app-configuration-cache.helpers";

type UpdateOneReq = AppAutoscale_UpdateOne_Req["data"];
type UpdateOneRes = AppAutoscale_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppAutoscaleApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            // Turning it on can scale the function, which the service settings show.
            void queryClient.invalidateQueries({ queryKey: [QK["projects.apps.autoscale.$.find-one"]] });
            invalidateSingleAppConfigurationQueries(queryClient, {
                projectID: request.projectID,
                appID: request.appID,
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppAutoscaleCommands = Object.freeze({
    useUpdateOne,
});

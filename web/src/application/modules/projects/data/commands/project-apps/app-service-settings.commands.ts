import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppServiceSettingsApi } from "../../../api/hooks/project-apps";
import { type AppServiceSettings_UpdateOne_Req, type AppServiceSettings_UpdateOne_Res } from "../../../api/services";

import { invalidateSingleAppConfigurationQueries } from "./app-configuration-cache.helpers";

type UpdateOneReq = AppServiceSettings_UpdateOne_Req["data"];
type UpdateOneRes = AppServiceSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppServiceSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            invalidateSingleAppConfigurationQueries(queryClient, {
                projectID: request.projectID,
                appID: request.appID,
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppServiceSettingsCommands = Object.freeze({
    useUpdateOne,
});

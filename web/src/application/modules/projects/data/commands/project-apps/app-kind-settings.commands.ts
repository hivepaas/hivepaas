import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppKindSettingsApi } from "../../../api/hooks/project-apps";
import { type AppKindSettings_UpdateOne_Req, type AppKindSettings_UpdateOne_Res } from "../../../api/services";

import { invalidateSingleAppConfigurationQueries } from "./app-configuration-cache.helpers";

type UpdateOneReq = AppKindSettings_UpdateOne_Req["data"];
type UpdateOneRes = AppKindSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppKindSettingsApi();
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

export const AppKindSettingsCommands = Object.freeze({
    useUpdateOne,
});

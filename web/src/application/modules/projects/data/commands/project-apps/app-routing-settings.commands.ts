import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppRoutingSettingsApi } from "../../../api/hooks/project-apps";
import { type AppRoutingSettings_UpdateOne_Req, type AppRoutingSettings_UpdateOne_Res } from "../../../api/services";

import { invalidateSingleAppConfigurationQueries } from "./app-configuration-cache.helpers";

type UpdateOneReq = AppRoutingSettings_UpdateOne_Req["data"];
type UpdateOneRes = AppRoutingSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppRoutingSettingsApi();
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

export const AppRoutingSettingsCommands = Object.freeze({
    useUpdateOne,
});

export { AppRoutingSettingsCommands as AppHttpSettingsCommands };

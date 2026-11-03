import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppFeatureSettingsApi } from "../../../api/hooks/project-apps";
import type { AppFeatureSettings_UpdateOne_Req, AppFeatureSettings_UpdateOne_Res } from "../../../api/services";

import { invalidateSingleAppConfigurationQueries } from "./app-configuration-cache.helpers";

type UpdateOneReq = AppFeatureSettings_UpdateOne_Req["data"];
type UpdateOneRes = AppFeatureSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppFeatureSettingsApi();
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

export const AppFeatureSettingsCommands = Object.freeze({
    useUpdateOne,
});

import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppDockerApiSettingsApi } from "../../../api/hooks/project-apps";
import type { AppDockerApiSettings_UpdateOne_Req, AppDockerApiSettings_UpdateOne_Res } from "../../../api/services";

import { invalidateSingleAppConfigurationQueries } from "./app-configuration-cache.helpers";

type UpdateOneReq = AppDockerApiSettings_UpdateOne_Req["data"];
type UpdateOneRes = AppDockerApiSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppDockerApiSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            // Access changes the service's mounts, networks and environment, which
            // the other screens show.
            invalidateSingleAppConfigurationQueries(queryClient, {
                projectID: request.projectID,
                appID: request.appID,
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppDockerApiSettingsCommands = Object.freeze({
    useUpdateOne,
});

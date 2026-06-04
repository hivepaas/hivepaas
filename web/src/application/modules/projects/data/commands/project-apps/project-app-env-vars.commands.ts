import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectAppEnvVarsApi } from "~/projects/api/hooks/project-apps";
import type { ProjectAppEnvVars_UpdateOne_Req, ProjectAppEnvVars_UpdateOne_Res } from "~/projects/api/services";

import { invalidateSingleAppConfigurationQueries } from "./app-configuration-cache.helpers";

/**
 * Update a project app env vars command
 */
type UpdateOneReq = ProjectAppEnvVars_UpdateOne_Req["data"];
type UpdateOneRes = ProjectAppEnvVars_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectAppEnvVarsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            invalidateSingleAppConfigurationQueries(queryClient, {
                projectID: request.projectID,
                appID: request.appID,
            });

            if (onSuccess) {
                onSuccess(response, request, ...rest);
            }
        },
        ...options,
    });
}

export const ProjectAppEnvVarsCommands = Object.freeze({
    useUpdateOne,
});

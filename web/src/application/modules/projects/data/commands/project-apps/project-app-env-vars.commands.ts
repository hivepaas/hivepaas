import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectAppEnvVarsApi } from "~/projects/api/hooks/project-apps";
import type { ProjectAppEnvVars_UpdateOne_Req, ProjectAppEnvVars_UpdateOne_Res } from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

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
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.env-vars.$.find-one"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

export const ProjectAppEnvVarsCommands = Object.freeze({
    useUpdateOne,
});

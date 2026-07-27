import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectEnvVarsApi } from "~/projects/api/hooks";
import type {
    ProjectEnvVars_Compute_Req,
    ProjectEnvVars_Compute_Res,
    ProjectEnvVars_UpdateOne_Req,
    ProjectEnvVars_UpdateOne_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

/**
 * Update project env vars command
 */
type UpdateOneReq = ProjectEnvVars_UpdateOne_Req["data"];
type UpdateOneRes = ProjectEnvVars_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectEnvVarsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.env-vars.$.find-one"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

/**
 * Compute project env vars command
 */
type ComputeReq = ProjectEnvVars_Compute_Req["data"];
type ComputeRes = ProjectEnvVars_Compute_Res;
type ComputeOptions = Omit<UseMutationOptions<ComputeRes, Error, ComputeReq>, "mutationFn">;

function useCompute({ onSuccess, ...options }: ComputeOptions = {}) {
    const { mutations } = useProjectEnvVarsApi();

    return useMutation({
        mutationFn: mutations.compute,
        onSuccess: (response, request, ...rest) => {
            if (onSuccess) {
                onSuccess(response, request, ...rest);
            }
        },
        ...options,
    });
}

export const ProjectEnvVarsCommands = Object.freeze({
    useUpdateOne,
    useCompute,
});

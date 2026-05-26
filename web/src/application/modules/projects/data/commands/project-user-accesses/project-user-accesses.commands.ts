import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectUserAccessesApi } from "~/projects/api/hooks";
import type { ProjectUserAccesses_UpdateOne_Req, ProjectUserAccesses_UpdateOne_Res } from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

/**
 * Update project user accesses command
 */
type UpdateOneReq = ProjectUserAccesses_UpdateOne_Req["data"];
type UpdateOneRes = ProjectUserAccesses_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectUserAccessesApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.user-accesses.$.find-one"]],
            });

            void queryClient.invalidateQueries({
                queryKey: [QK["projects.$.find-one-by-id"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

export const ProjectUserAccessesCommands = Object.freeze({
    useUpdateOne,
});

import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectsApi } from "~/projects/api/hooks";
import type {
    Projects_CreateOne_Req,
    Projects_CreateOne_Res,
    Projects_DeleteOne_Req,
    Projects_DeleteOne_Res,
} from "~/projects/api/services";

import { QK } from "~/projects/data/constants";

/**
 * Create a project command
 */
type CreateOneReq = Projects_CreateOne_Req["data"];
type CreateOneRes = Projects_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

/**
 * Delete a project command
 */
type DeleteOneReq = Projects_DeleteOne_Req["data"];
type DeleteOneRes = Projects_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

export const ProjectsCommands = Object.freeze({
    useCreateOne,
    useDeleteOne,
});


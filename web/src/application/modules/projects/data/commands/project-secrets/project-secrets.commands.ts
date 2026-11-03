import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectSecretsApi } from "~/projects/api/hooks";
import type {
    ProjectSecrets_CreateOne_Req,
    ProjectSecrets_CreateOne_Res,
    ProjectSecrets_DeleteOne_Req,
    ProjectSecrets_DeleteOne_Res,
    ProjectSecrets_UpdateOne_Req,
    ProjectSecrets_UpdateOne_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

/**
 * Create a project secret command
 */
type CreateOneReq = ProjectSecrets_CreateOne_Req["data"];
type CreateOneRes = ProjectSecrets_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectSecretsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.secrets.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

/**
 * Delete a project secret command
 */
type DeleteOneReq = ProjectSecrets_DeleteOne_Req["data"];
type DeleteOneRes = ProjectSecrets_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectSecretsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.secrets.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

/**
 * Update a project secret command
 */
type UpdateOneReq = ProjectSecrets_UpdateOne_Req["data"];
type UpdateOneRes = ProjectSecrets_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectSecretsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.secrets.$.find-many-paginated"]],
            });

            void queryClient.invalidateQueries({
                queryKey: [QK["projects.secrets.$.find-one-by-id"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

export const ProjectSecretsCommands = Object.freeze({
    useCreateOne,
    useDeleteOne,
    useUpdateOne,
});

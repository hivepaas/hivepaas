import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectConfigFilesApi } from "~/projects/api/hooks";
import type {
    ProjectConfigFiles_CreateOne_Req,
    ProjectConfigFiles_CreateOne_Res,
    ProjectConfigFiles_DeleteOne_Req,
    ProjectConfigFiles_DeleteOne_Res,
    ProjectConfigFiles_UpdateOne_Req,
    ProjectConfigFiles_UpdateOne_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

/**
 * Create a project config file command
 */
type CreateOneReq = ProjectConfigFiles_CreateOne_Req["data"];
type CreateOneRes = ProjectConfigFiles_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectConfigFilesApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.config-files.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

/**
 * Delete a project config file command
 */
type DeleteOneReq = ProjectConfigFiles_DeleteOne_Req["data"];
type DeleteOneRes = ProjectConfigFiles_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectConfigFilesApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.config-files.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

/**
 * Update a project config file command
 */
type UpdateOneReq = ProjectConfigFiles_UpdateOne_Req["data"];
type UpdateOneRes = ProjectConfigFiles_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectConfigFilesApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.config-files.$.find-many-paginated"]],
            });

            void queryClient.invalidateQueries({
                queryKey: [QK["projects.config-files.$.find-one-by-id"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

export const ProjectConfigFilesCommands = Object.freeze({
    useCreateOne,
    useDeleteOne,
    useUpdateOne,
});

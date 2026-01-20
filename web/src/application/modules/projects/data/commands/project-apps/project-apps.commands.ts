import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectAppsApi } from "~/projects/api/hooks";
import type {
    ProjectApps_CreateOne_Req,
    ProjectApps_CreateOne_Res,
    ProjectApps_DeleteOne_Req,
    ProjectApps_DeleteOne_Res,
    ProjectApps_UpdateOne_Req,
    ProjectApps_UpdateOne_Res,
} from "~/projects/api/services";

import { QK } from "~/projects/data/constants";

/**
 * Create a project app command
 */
type CreateOneReq = ProjectApps_CreateOne_Req["data"];
type CreateOneRes = ProjectApps_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectAppsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

/**
 * Delete a project app command
 */
type DeleteOneReq = ProjectApps_DeleteOne_Req["data"];
type DeleteOneRes = ProjectApps_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectAppsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

/**
 * Update a project app command
 */
type UpdateOneReq = ProjectApps_UpdateOne_Req["data"];
type UpdateOneRes = ProjectApps_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectAppsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.$.find-many-paginated"]],
            });

            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.$.find-one-by-id"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

export const ProjectAppsCommands = Object.freeze({
    useCreateOne,
    useDeleteOne,
    useUpdateOne,
});

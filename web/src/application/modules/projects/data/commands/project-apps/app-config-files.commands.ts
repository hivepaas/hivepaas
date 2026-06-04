import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAppConfigFilesApi } from "~/projects/api/hooks/project-apps";
import type {
    AppConfigFiles_CreateOne_Req,
    AppConfigFiles_CreateOne_Res,
    AppConfigFiles_DeleteOne_Req,
    AppConfigFiles_DeleteOne_Res,
    AppConfigFiles_UpdateOne_Req,
    AppConfigFiles_UpdateOne_Res,
} from "~/projects/api/services";

import { invalidateSingleAppConfigurationQueries } from "./app-configuration-cache.helpers";

/**
 * Create an app config file command
 */
type CreateOneReq = AppConfigFiles_CreateOne_Req["data"];
type CreateOneRes = AppConfigFiles_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useAppConfigFilesApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
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

/**
 * Delete an app config file command
 */
type DeleteOneReq = AppConfigFiles_DeleteOne_Req["data"];
type DeleteOneRes = AppConfigFiles_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useAppConfigFilesApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
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

/**
 * Update an app config file command
 */
type UpdateOneReq = AppConfigFiles_UpdateOne_Req["data"];
type UpdateOneRes = AppConfigFiles_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppConfigFilesApi();

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

export const AppConfigFilesCommands = Object.freeze({
    useCreateOne,
    useDeleteOne,
    useUpdateOne,
});

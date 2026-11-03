import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectAppSecretsApi } from "~/projects/api/hooks/project-apps";
import type {
    AppSecrets_CreateOne_Req,
    AppSecrets_CreateOne_Res,
    AppSecrets_DeleteOne_Req,
    AppSecrets_DeleteOne_Res,
    AppSecrets_UpdateOne_Req,
    AppSecrets_UpdateOne_Res,
} from "~/projects/api/services";

import { invalidateSingleAppConfigurationQueries } from "./app-configuration-cache.helpers";

/**
 * Create an app secret command
 */
type CreateOneReq = AppSecrets_CreateOne_Req["data"];
type CreateOneRes = AppSecrets_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectAppSecretsApi();

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
 * Delete an app secret command
 */
type DeleteOneReq = AppSecrets_DeleteOne_Req["data"];
type DeleteOneRes = AppSecrets_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectAppSecretsApi();

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
 * Update an app secret command
 */
type UpdateOneReq = AppSecrets_UpdateOne_Req["data"];
type UpdateOneRes = AppSecrets_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectAppSecretsApi();

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

export const ProjectAppSecretsCommands = Object.freeze({
    useCreateOne,
    useDeleteOne,
    useUpdateOne,
});

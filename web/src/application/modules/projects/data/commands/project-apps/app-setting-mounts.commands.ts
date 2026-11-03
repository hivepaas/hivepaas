import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAppSettingMountsApi } from "~/projects/api/hooks/project-apps";
import type {
    AppSettingMounts_CreateOne_Req,
    AppSettingMounts_CreateOne_Res,
    AppSettingMounts_DeleteOne_Req,
    AppSettingMounts_DeleteOne_Res,
    AppSettingMounts_UpdateOne_Req,
    AppSettingMounts_UpdateOne_Res,
    AppSettingMounts_UpdateStatus_Req,
    AppSettingMounts_UpdateStatus_Res,
} from "~/projects/api/services";

import { invalidateSingleAppConfigurationQueries } from "./app-configuration-cache.helpers";

/**
 * Create an app setting mount command
 */
type CreateOneReq = AppSettingMounts_CreateOne_Req["data"];
type CreateOneRes = AppSettingMounts_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useAppSettingMountsApi();

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
 * Update an app setting mount command
 */
type UpdateOneReq = AppSettingMounts_UpdateOne_Req["data"];
type UpdateOneRes = AppSettingMounts_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppSettingMountsApi();

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

/**
 * Enable or disable an app setting mount command
 */
type UpdateStatusReq = AppSettingMounts_UpdateStatus_Req["data"];
type UpdateStatusRes = AppSettingMounts_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useAppSettingMountsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
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
 * Delete an app setting mount command
 */
type DeleteOneReq = AppSettingMounts_DeleteOne_Req["data"];
type DeleteOneRes = AppSettingMounts_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useAppSettingMountsApi();

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

export const AppSettingMountsCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});

import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useNotificationsApi } from "~/settings/api/hooks";
import {
    type Notifications_CreateOne_Req,
    type Notifications_CreateOne_Res,
    type Notifications_DeleteOne_Req,
    type Notifications_DeleteOne_Res,
    type Notifications_UpdateOne_Req,
    type Notifications_UpdateOne_Res,
    type Notifications_UpdateStatus_Req,
    type Notifications_UpdateStatus_Res,
} from "~/settings/api/services/notifications-services";
import { QK } from "~/settings/data/constants";

/**
 * Create a node command
 */
type CreateOneReq = Notifications_CreateOne_Req["data"];
type CreateOneRes = Notifications_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useNotificationsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.notifications.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.notifications.find-one-by-id"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

/**
 * Update a node command
 */
type UpdateOneReq = Notifications_UpdateOne_Req["data"];
type UpdateOneRes = Notifications_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useNotificationsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.notifications.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.notifications.find-one-by-id"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

type UpdateStatusReq = Notifications_UpdateStatus_Req["data"];
type UpdateStatusRes = Notifications_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useNotificationsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.notifications.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.notifications.find-one-by-id"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

/**
 * Delete a node command
 */
type DeleteOneReq = Notifications_DeleteOne_Req["data"];
type DeleteOneRes = Notifications_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useNotificationsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.notifications.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.notifications.find-one-by-id"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

export const NotificationCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});

import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectNotificationsApi } from "~/projects/api/hooks";
import type {
    ProjectNotification_CreateOne_Req,
    ProjectNotification_CreateOne_Res,
    ProjectNotification_DeleteOne_Req,
    ProjectNotification_DeleteOne_Res,
    ProjectNotification_UpdateOne_Req,
    ProjectNotification_UpdateOne_Res,
    ProjectNotification_UpdateStatus_Req,
    ProjectNotification_UpdateStatus_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type CreateOneReq = ProjectNotification_CreateOne_Req["data"];
type CreateOneRes = ProjectNotification_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectNotificationsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.notifications.$.find-many-paginated"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = ProjectNotification_UpdateOne_Req["data"];
type UpdateOneRes = ProjectNotification_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectNotificationsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.notifications.$.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.notifications.$.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = ProjectNotification_UpdateStatus_Req["data"];
type UpdateStatusRes = ProjectNotification_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useProjectNotificationsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.notifications.$.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.notifications.$.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = ProjectNotification_DeleteOne_Req["data"];
type DeleteOneRes = ProjectNotification_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectNotificationsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.notifications.$.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.notifications.$.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const ProjectNotificationCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});

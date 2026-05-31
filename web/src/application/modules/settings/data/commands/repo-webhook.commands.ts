import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { QK as PROJECT_QK } from "~/projects/data/constants";
import { useRepoWebhookApi } from "~/settings/api/hooks";
import type {
    RepoWebhook_CreateOne_Req,
    RepoWebhook_CreateOne_Res,
    RepoWebhook_DeleteOne_Req,
    RepoWebhook_DeleteOne_Res,
    RepoWebhook_UpdateOne_Req,
    RepoWebhook_UpdateOne_Res,
    RepoWebhook_UpdateStatus_Req,
    RepoWebhook_UpdateStatus_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type CreateOneReq = RepoWebhook_CreateOne_Req["data"];
type CreateOneRes = RepoWebhook_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useRepoWebhookApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.repo-webhook.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [PROJECT_QK["projects.repo-webhook.$.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = RepoWebhook_UpdateOne_Req["data"];
type UpdateOneRes = RepoWebhook_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useRepoWebhookApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.repo-webhook.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.repo-webhook.find-one-by-id"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [PROJECT_QK["projects.repo-webhook.$.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = RepoWebhook_UpdateStatus_Req["data"];
type UpdateStatusRes = RepoWebhook_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useRepoWebhookApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.repo-webhook.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.repo-webhook.find-one-by-id"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [PROJECT_QK["projects.repo-webhook.$.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = RepoWebhook_DeleteOne_Req["data"];
type DeleteOneRes = RepoWebhook_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useRepoWebhookApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.repo-webhook.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.repo-webhook.find-one-by-id"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [PROJECT_QK["projects.repo-webhook.$.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const RepoWebhookCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});

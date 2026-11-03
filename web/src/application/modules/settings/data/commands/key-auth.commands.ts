import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useKeyAuthApi } from "~/settings/api/hooks";
import type {
    KeyAuth_CreateOne_Req,
    KeyAuth_CreateOne_Res,
    KeyAuth_DeleteOne_Req,
    KeyAuth_DeleteOne_Res,
    KeyAuth_UpdateOne_Req,
    KeyAuth_UpdateOne_Res,
    KeyAuth_UpdateStatus_Req,
    KeyAuth_UpdateStatus_Res,
} from "~/settings/api/services/key-auth-services";
import { QK } from "~/settings/data/constants";

type CreateOneReq = KeyAuth_CreateOne_Req["data"];
type CreateOneRes = KeyAuth_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useKeyAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.key-auth.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = KeyAuth_UpdateOne_Req["data"];
type UpdateOneRes = KeyAuth_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useKeyAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.key-auth.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.key-auth.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = KeyAuth_UpdateStatus_Req["data"];
type UpdateStatusRes = KeyAuth_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useKeyAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.key-auth.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.key-auth.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = KeyAuth_DeleteOne_Req["data"];
type DeleteOneRes = KeyAuth_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useKeyAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.key-auth.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.key-auth.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const KeyAuthCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});

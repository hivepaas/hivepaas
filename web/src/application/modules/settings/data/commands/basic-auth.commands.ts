import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useBasicAuthApi } from "~/settings/api/hooks";
import type {
    BasicAuth_CreateOne_Req,
    BasicAuth_CreateOne_Res,
    BasicAuth_DeleteOne_Req,
    BasicAuth_DeleteOne_Res,
    BasicAuth_UpdateOne_Req,
    BasicAuth_UpdateOne_Res,
    BasicAuth_UpdateStatus_Req,
    BasicAuth_UpdateStatus_Res,
} from "~/settings/api/services/basic-auth-services";
import { QK } from "~/settings/data/constants";

type CreateOneReq = BasicAuth_CreateOne_Req["data"];
type CreateOneRes = BasicAuth_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useBasicAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.basic-auth.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = BasicAuth_UpdateOne_Req["data"];
type UpdateOneRes = BasicAuth_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useBasicAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.basic-auth.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.basic-auth.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = BasicAuth_UpdateStatus_Req["data"];
type UpdateStatusRes = BasicAuth_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useBasicAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.basic-auth.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.basic-auth.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = BasicAuth_DeleteOne_Req["data"];
type DeleteOneRes = BasicAuth_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useBasicAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.basic-auth.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.basic-auth.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const BasicAuthCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});

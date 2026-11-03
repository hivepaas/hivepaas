import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSslProviderApi } from "~/settings/api/hooks";
import type {
    SslProvider_CreateOne_Req,
    SslProvider_CreateOne_Res,
    SslProvider_DeleteOne_Req,
    SslProvider_DeleteOne_Res,
    SslProvider_UpdateOne_Req,
    SslProvider_UpdateOne_Res,
    SslProvider_UpdateStatus_Req,
    SslProvider_UpdateStatus_Res,
} from "~/settings/api/services/ssl-provider-services";
import { QK } from "~/settings/data/constants";

type CreateOneReq = SslProvider_CreateOne_Req["data"];
type CreateOneRes = SslProvider_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useSslProviderApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.ssl-provider.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = SslProvider_UpdateOne_Req["data"];
type UpdateOneRes = SslProvider_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useSslProviderApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.ssl-provider.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.ssl-provider.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = SslProvider_UpdateStatus_Req["data"];
type UpdateStatusRes = SslProvider_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useSslProviderApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.ssl-provider.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.ssl-provider.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = SslProvider_DeleteOne_Req["data"];
type DeleteOneRes = SslProvider_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useSslProviderApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.ssl-provider.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.ssl-provider.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const SslProviderCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});

import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSslCertApi } from "~/settings/api/hooks";
import type {
    SslCert_CreateOne_Req,
    SslCert_CreateOne_Res,
    SslCert_DeleteOne_Req,
    SslCert_DeleteOne_Res,
    SslCert_UpdateOne_Req,
    SslCert_UpdateOne_Res,
    SslCert_UpdateStatus_Req,
    SslCert_UpdateStatus_Res,
} from "~/settings/api/services/ssl-cert-services";
import { QK } from "~/settings/data/constants";

type CreateOneReq = SslCert_CreateOne_Req["data"];
type CreateOneRes = SslCert_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useSslCertApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.ssl-cert.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = SslCert_UpdateOne_Req["data"];
type UpdateOneRes = SslCert_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useSslCertApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.ssl-cert.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.ssl-cert.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = SslCert_UpdateStatus_Req["data"];
type UpdateStatusRes = SslCert_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useSslCertApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.ssl-cert.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.ssl-cert.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = SslCert_DeleteOne_Req["data"];
type DeleteOneRes = SslCert_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useSslCertApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.ssl-cert.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.ssl-cert.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const SslCertCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});

import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAcmeDnsProviderApi } from "~/settings/api/hooks";
import type {
    AcmeDnsProvider_CreateOne_Req,
    AcmeDnsProvider_CreateOne_Res,
    AcmeDnsProvider_DeleteOne_Req,
    AcmeDnsProvider_DeleteOne_Res,
    AcmeDnsProvider_TestAccess_Req,
    AcmeDnsProvider_TestAccess_Res,
    AcmeDnsProvider_UpdateOne_Req,
    AcmeDnsProvider_UpdateOne_Res,
    AcmeDnsProvider_UpdateStatus_Req,
    AcmeDnsProvider_UpdateStatus_Res,
} from "~/settings/api/services/acme-dns-provider-services";
import { QK } from "~/settings/data/constants";

type CreateOneReq = AcmeDnsProvider_CreateOne_Req["data"];
type CreateOneRes = AcmeDnsProvider_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useAcmeDnsProviderApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.acme-dns-provider.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = AcmeDnsProvider_UpdateOne_Req["data"];
type UpdateOneRes = AcmeDnsProvider_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAcmeDnsProviderApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.acme-dns-provider.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.acme-dns-provider.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = AcmeDnsProvider_UpdateStatus_Req["data"];
type UpdateStatusRes = AcmeDnsProvider_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useAcmeDnsProviderApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.acme-dns-provider.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.acme-dns-provider.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = AcmeDnsProvider_DeleteOne_Req["data"];
type DeleteOneRes = AcmeDnsProvider_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useAcmeDnsProviderApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.acme-dns-provider.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.acme-dns-provider.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type TestAccessReq = AcmeDnsProvider_TestAccess_Req["data"];
type TestAccessRes = AcmeDnsProvider_TestAccess_Res;
type TestAccessOptions = Omit<UseMutationOptions<TestAccessRes, Error, TestAccessReq>, "mutationFn">;

function useTestAccess(options: TestAccessOptions = {}) {
    const { mutations } = useAcmeDnsProviderApi();

    return useMutation({
        mutationFn: mutations.testAccess,
        ...options,
    });
}

export const AcmeDnsProviderCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
    useTestAccess,
});

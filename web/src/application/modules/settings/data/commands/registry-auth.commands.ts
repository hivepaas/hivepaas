import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useRegistryAuthApi } from "~/settings/api/hooks";
import type {
    RegistryAuth_CreateOne_Req,
    RegistryAuth_CreateOne_Res,
    RegistryAuth_DeleteOne_Req,
    RegistryAuth_DeleteOne_Res,
    RegistryAuth_TestConn_Req,
    RegistryAuth_TestConn_Res,
    RegistryAuth_UpdateMeta_Req,
    RegistryAuth_UpdateMeta_Res,
    RegistryAuth_UpdateOne_Req,
    RegistryAuth_UpdateOne_Res,
} from "~/settings/api/services/registry-auth-services";
import { QK } from "~/settings/data/constants";

type CreateOneReq = RegistryAuth_CreateOne_Req["data"];
type CreateOneRes = RegistryAuth_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useRegistryAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.registry-auth.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = RegistryAuth_UpdateOne_Req["data"];
type UpdateOneRes = RegistryAuth_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useRegistryAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.registry-auth.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.registry-auth.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateMetaReq = RegistryAuth_UpdateMeta_Req["data"];
type UpdateMetaRes = RegistryAuth_UpdateMeta_Res;
type UpdateMetaOptions = Omit<UseMutationOptions<UpdateMetaRes, Error, UpdateMetaReq>, "mutationFn">;

function useUpdateMeta({ onSuccess, ...options }: UpdateMetaOptions = {}) {
    const { mutations } = useRegistryAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateMeta,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.registry-auth.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.registry-auth.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = RegistryAuth_DeleteOne_Req["data"];
type DeleteOneRes = RegistryAuth_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useRegistryAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.registry-auth.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.registry-auth.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type TestConnReq = RegistryAuth_TestConn_Req["data"];
type TestConnRes = RegistryAuth_TestConn_Res;
type TestConnOptions = Omit<UseMutationOptions<TestConnRes, Error, TestConnReq>, "mutationFn">;

function useTestConn(options: TestConnOptions = {}) {
    const { mutations } = useRegistryAuthApi();

    return useMutation({
        mutationFn: mutations.testConn,
        ...options,
    });
}

export const RegistryAuthCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateMeta,
    useDeleteOne,
    useTestConn,
});

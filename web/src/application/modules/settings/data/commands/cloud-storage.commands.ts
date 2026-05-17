import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useCloudStorageApi } from "~/settings/api/hooks";
import type {
    CloudStorage_CreateOne_Req,
    CloudStorage_CreateOne_Res,
    CloudStorage_DeleteOne_Req,
    CloudStorage_DeleteOne_Res,
    CloudStorage_TestConn_Req,
    CloudStorage_TestConn_Res,
    CloudStorage_UpdateMeta_Req,
    CloudStorage_UpdateMeta_Res,
    CloudStorage_UpdateOne_Req,
    CloudStorage_UpdateOne_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type CreateOneReq = CloudStorage_CreateOne_Req["data"];
type CreateOneRes = CloudStorage_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useCloudStorageApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.cloud-storage.find-many-paginated"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = CloudStorage_UpdateOne_Req["data"];
type UpdateOneRes = CloudStorage_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useCloudStorageApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.cloud-storage.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.cloud-storage.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateMetaReq = CloudStorage_UpdateMeta_Req["data"];
type UpdateMetaRes = CloudStorage_UpdateMeta_Res;
type UpdateMetaOptions = Omit<UseMutationOptions<UpdateMetaRes, Error, UpdateMetaReq>, "mutationFn">;

function useUpdateMeta({ onSuccess, ...options }: UpdateMetaOptions = {}) {
    const { mutations } = useCloudStorageApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateMeta,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.cloud-storage.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.cloud-storage.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = CloudStorage_DeleteOne_Req["data"];
type DeleteOneRes = CloudStorage_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useCloudStorageApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.cloud-storage.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.cloud-storage.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type TestConnReq = CloudStorage_TestConn_Req["data"];
type TestConnRes = CloudStorage_TestConn_Res;
type TestConnOptions = Omit<UseMutationOptions<TestConnRes, Error, TestConnReq>, "mutationFn">;

function useTestConn(options: TestConnOptions = {}) {
    const { mutations } = useCloudStorageApi();

    return useMutation({
        mutationFn: mutations.testConn,
        ...options,
    });
}

export const CloudStorageCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateMeta,
    useDeleteOne,
    useTestConn,
});

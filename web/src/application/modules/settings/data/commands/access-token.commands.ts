import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAccessTokenApi } from "~/settings/api/hooks";
import type {
    AccessToken_CreateOne_Req,
    AccessToken_CreateOne_Res,
    AccessToken_DeleteOne_Req,
    AccessToken_DeleteOne_Res,
    AccessToken_TestConn_Req,
    AccessToken_TestConn_Res,
    AccessToken_UpdateMeta_Req,
    AccessToken_UpdateMeta_Res,
    AccessToken_UpdateOne_Req,
    AccessToken_UpdateOne_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type CreateOneReq = AccessToken_CreateOne_Req["data"];
type CreateOneRes = AccessToken_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useAccessTokenApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.access-token.find-many-paginated"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = AccessToken_UpdateOne_Req["data"];
type UpdateOneRes = AccessToken_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAccessTokenApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.access-token.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.access-token.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateMetaReq = AccessToken_UpdateMeta_Req["data"];
type UpdateMetaRes = AccessToken_UpdateMeta_Res;
type UpdateMetaOptions = Omit<UseMutationOptions<UpdateMetaRes, Error, UpdateMetaReq>, "mutationFn">;

function useUpdateMeta({ onSuccess, ...options }: UpdateMetaOptions = {}) {
    const { mutations } = useAccessTokenApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateMeta,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.access-token.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.access-token.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = AccessToken_DeleteOne_Req["data"];
type DeleteOneRes = AccessToken_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useAccessTokenApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.access-token.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.access-token.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type TestConnReq = AccessToken_TestConn_Req["data"];
type TestConnRes = AccessToken_TestConn_Res;
type TestConnOptions = Omit<UseMutationOptions<TestConnRes, Error, TestConnReq>, "mutationFn">;

function useTestConn(options: TestConnOptions = {}) {
    const { mutations } = useAccessTokenApi();

    return useMutation({
        mutationFn: mutations.testConn,
        ...options,
    });
}

export const AccessTokenCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateMeta,
    useDeleteOne,
    useTestConn,
});

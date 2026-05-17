import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useOAuthApi } from "~/settings/api/hooks";
import type {
    OAuth_CreateOne_Req,
    OAuth_CreateOne_Res,
    OAuth_DeleteOne_Req,
    OAuth_DeleteOne_Res,
    OAuth_UpdateMeta_Req,
    OAuth_UpdateMeta_Res,
    OAuth_UpdateOne_Req,
    OAuth_UpdateOne_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type CreateOneReq = OAuth_CreateOne_Req["data"];
type CreateOneRes = OAuth_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useOAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.oauth.find-many-paginated"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = OAuth_UpdateOne_Req["data"];
type UpdateOneRes = OAuth_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useOAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.oauth.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.oauth.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateMetaReq = OAuth_UpdateMeta_Req["data"];
type UpdateMetaRes = OAuth_UpdateMeta_Res;
type UpdateMetaOptions = Omit<UseMutationOptions<UpdateMetaRes, Error, UpdateMetaReq>, "mutationFn">;

function useUpdateMeta({ onSuccess, ...options }: UpdateMetaOptions = {}) {
    const { mutations } = useOAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateMeta,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.oauth.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.oauth.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = OAuth_DeleteOne_Req["data"];
type DeleteOneRes = OAuth_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useOAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.oauth.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.oauth.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const OAuthCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateMeta,
    useDeleteOne,
});

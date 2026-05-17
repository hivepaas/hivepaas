import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSSHKeyApi } from "~/settings/api/hooks";
import type {
    SSHKey_CreateOne_Req,
    SSHKey_CreateOne_Res,
    SSHKey_DeleteOne_Req,
    SSHKey_DeleteOne_Res,
    SSHKey_UpdateMeta_Req,
    SSHKey_UpdateMeta_Res,
    SSHKey_UpdateOne_Req,
    SSHKey_UpdateOne_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type CreateOneReq = SSHKey_CreateOne_Req["data"];
type CreateOneRes = SSHKey_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useSSHKeyApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.ssh-key.find-many-paginated"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = SSHKey_UpdateOne_Req["data"];
type UpdateOneRes = SSHKey_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useSSHKeyApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.ssh-key.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.ssh-key.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateMetaReq = SSHKey_UpdateMeta_Req["data"];
type UpdateMetaRes = SSHKey_UpdateMeta_Res;
type UpdateMetaOptions = Omit<UseMutationOptions<UpdateMetaRes, Error, UpdateMetaReq>, "mutationFn">;

function useUpdateMeta({ onSuccess, ...options }: UpdateMetaOptions = {}) {
    const { mutations } = useSSHKeyApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateMeta,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.ssh-key.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.ssh-key.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = SSHKey_DeleteOne_Req["data"];
type DeleteOneRes = SSHKey_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useSSHKeyApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.ssh-key.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.ssh-key.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const SSHKeyCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateMeta,
    useDeleteOne,
});

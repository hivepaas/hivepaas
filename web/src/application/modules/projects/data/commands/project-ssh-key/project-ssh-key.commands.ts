import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectSSHKeyApi } from "~/projects/api/hooks";
import type {
    ProjectSSHKey_CreateOne_Req,
    ProjectSSHKey_CreateOne_Res,
    ProjectSSHKey_DeleteOne_Req,
    ProjectSSHKey_DeleteOne_Res,
    ProjectSSHKey_UpdateOne_Req,
    ProjectSSHKey_UpdateOne_Res,
    ProjectSSHKey_UpdateStatus_Req,
    ProjectSSHKey_UpdateStatus_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type CreateOneReq = ProjectSSHKey_CreateOne_Req["data"];
type CreateOneRes = ProjectSSHKey_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectSSHKeyApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.ssh-key.$.find-many-paginated"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = ProjectSSHKey_UpdateOne_Req["data"];
type UpdateOneRes = ProjectSSHKey_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectSSHKeyApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.ssh-key.$.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.ssh-key.$.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = ProjectSSHKey_UpdateStatus_Req["data"];
type UpdateStatusRes = ProjectSSHKey_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useProjectSSHKeyApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.ssh-key.$.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.ssh-key.$.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = ProjectSSHKey_DeleteOne_Req["data"];
type DeleteOneRes = ProjectSSHKey_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectSSHKeyApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.ssh-key.$.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.ssh-key.$.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const ProjectSSHKeyCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});

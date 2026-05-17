import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectCloudStorageApi } from "~/projects/api/hooks";
import type {
    ProjectCloudStorage_CreateOne_Req,
    ProjectCloudStorage_CreateOne_Res,
    ProjectCloudStorage_DeleteOne_Req,
    ProjectCloudStorage_DeleteOne_Res,
    ProjectCloudStorage_UpdateMeta_Req,
    ProjectCloudStorage_UpdateMeta_Res,
    ProjectCloudStorage_UpdateOne_Req,
    ProjectCloudStorage_UpdateOne_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type CreateOneReq = ProjectCloudStorage_CreateOne_Req["data"];
type CreateOneRes = ProjectCloudStorage_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectCloudStorageApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.cloud-storage.$.find-many-paginated"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = ProjectCloudStorage_UpdateOne_Req["data"];
type UpdateOneRes = ProjectCloudStorage_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectCloudStorageApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.cloud-storage.$.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.cloud-storage.$.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateMetaReq = ProjectCloudStorage_UpdateMeta_Req["data"];
type UpdateMetaRes = ProjectCloudStorage_UpdateMeta_Res;
type UpdateMetaOptions = Omit<UseMutationOptions<UpdateMetaRes, Error, UpdateMetaReq>, "mutationFn">;

function useUpdateMeta({ onSuccess, ...options }: UpdateMetaOptions = {}) {
    const { mutations } = useProjectCloudStorageApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateMeta,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.cloud-storage.$.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.cloud-storage.$.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = ProjectCloudStorage_DeleteOne_Req["data"];
type DeleteOneRes = ProjectCloudStorage_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectCloudStorageApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.cloud-storage.$.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.cloud-storage.$.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const ProjectCloudStorageCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateMeta,
    useDeleteOne,
});

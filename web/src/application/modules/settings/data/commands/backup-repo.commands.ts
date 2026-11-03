import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useBackupRepoApi } from "~/settings/api/hooks";
import type {
    BackupRepo_Cleanup_Req,
    BackupRepo_Cleanup_Res,
    BackupRepo_CreateOne_Req,
    BackupRepo_CreateOne_Res,
    BackupRepo_DeleteOne_Req,
    BackupRepo_DeleteOne_Res,
    BackupRepo_Sync_Req,
    BackupRepo_Sync_Res,
    BackupRepo_UpdateOne_Req,
    BackupRepo_UpdateOne_Res,
    BackupRepo_UpdatePassword_Req,
    BackupRepo_UpdatePassword_Res,
    BackupRepo_UpdateStatus_Req,
    BackupRepo_UpdateStatus_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type CreateOneReq = BackupRepo_CreateOne_Req["data"];
type CreateOneRes = BackupRepo_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useBackupRepoApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.backup-repos.find-many-paginated"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = BackupRepo_UpdateOne_Req["data"];
type UpdateOneRes = BackupRepo_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useBackupRepoApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.backup-repos.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.backup-repos.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = BackupRepo_UpdateStatus_Req["data"];
type UpdateStatusRes = BackupRepo_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useBackupRepoApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.backup-repos.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.backup-repos.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = BackupRepo_DeleteOne_Req["data"];
type DeleteOneRes = BackupRepo_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useBackupRepoApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.backup-repos.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.backup-repos.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdatePasswordReq = BackupRepo_UpdatePassword_Req["data"];
type UpdatePasswordRes = BackupRepo_UpdatePassword_Res;
type UpdatePasswordOptions = Omit<UseMutationOptions<UpdatePasswordRes, Error, UpdatePasswordReq>, "mutationFn">;

function useUpdatePassword({ onSuccess, ...options }: UpdatePasswordOptions = {}) {
    const { mutations } = useBackupRepoApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updatePassword,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.backup-repos.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.backup-repos.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type CleanupReq = BackupRepo_Cleanup_Req["data"];
type CleanupRes = BackupRepo_Cleanup_Res;
type CleanupOptions = Omit<UseMutationOptions<CleanupRes, Error, CleanupReq>, "mutationFn">;

function useCleanup({ onSuccess, ...options }: CleanupOptions = {}) {
    const { mutations } = useBackupRepoApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.cleanup,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.backup-repos.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.backup-repos.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type SyncReq = BackupRepo_Sync_Req["data"];
type SyncRes = BackupRepo_Sync_Res;
type SyncOptions = Omit<UseMutationOptions<SyncRes, Error, SyncReq>, "mutationFn">;

function useSync({ onSuccess, ...options }: SyncOptions = {}) {
    const { mutations } = useBackupRepoApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.sync,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.backup-repos.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["settings.backup-repos.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const BackupRepoCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useUpdatePassword,
    useCleanup,
    useSync,
    useDeleteOne,
});

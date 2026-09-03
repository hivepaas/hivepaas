import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectBackupRepoApi } from "~/projects/api/hooks";
import type {
    ProjectBackupRepo_Cleanup_Req,
    ProjectBackupRepo_Cleanup_Res,
    ProjectBackupRepo_CreateOne_Req,
    ProjectBackupRepo_CreateOne_Res,
    ProjectBackupRepo_DeleteOne_Req,
    ProjectBackupRepo_DeleteOne_Res,
    ProjectBackupRepo_Sync_Req,
    ProjectBackupRepo_Sync_Res,
    ProjectBackupRepo_UpdateOne_Req,
    ProjectBackupRepo_UpdateOne_Res,
    ProjectBackupRepo_UpdatePassword_Req,
    ProjectBackupRepo_UpdatePassword_Res,
    ProjectBackupRepo_UpdateStatus_Req,
    ProjectBackupRepo_UpdateStatus_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type CreateOneReq = ProjectBackupRepo_CreateOne_Req["data"];
type CreateOneRes = ProjectBackupRepo_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectBackupRepoApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.backup-repos.$.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = ProjectBackupRepo_UpdateOne_Req["data"];
type UpdateOneRes = ProjectBackupRepo_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectBackupRepoApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.backup-repos.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.backup-repos.$.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = ProjectBackupRepo_UpdateStatus_Req["data"];
type UpdateStatusRes = ProjectBackupRepo_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useProjectBackupRepoApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.backup-repos.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.backup-repos.$.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = ProjectBackupRepo_DeleteOne_Req["data"];
type DeleteOneRes = ProjectBackupRepo_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectBackupRepoApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.backup-repos.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.backup-repos.$.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdatePasswordReq = ProjectBackupRepo_UpdatePassword_Req["data"];
type UpdatePasswordRes = ProjectBackupRepo_UpdatePassword_Res;
type UpdatePasswordOptions = Omit<UseMutationOptions<UpdatePasswordRes, Error, UpdatePasswordReq>, "mutationFn">;

function useUpdatePassword({ onSuccess, ...options }: UpdatePasswordOptions = {}) {
    const { mutations } = useProjectBackupRepoApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updatePassword,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.backup-repos.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.backup-repos.$.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type CleanupReq = ProjectBackupRepo_Cleanup_Req["data"];
type CleanupRes = ProjectBackupRepo_Cleanup_Res;
type CleanupOptions = Omit<UseMutationOptions<CleanupRes, Error, CleanupReq>, "mutationFn">;

function useCleanup({ onSuccess, ...options }: CleanupOptions = {}) {
    const { mutations } = useProjectBackupRepoApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.cleanup,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.backup-repos.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.backup-repos.$.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type SyncReq = ProjectBackupRepo_Sync_Req["data"];
type SyncRes = ProjectBackupRepo_Sync_Res;
type SyncOptions = Omit<UseMutationOptions<SyncRes, Error, SyncReq>, "mutationFn">;

function useSync({ onSuccess, ...options }: SyncOptions = {}) {
    const { mutations } = useProjectBackupRepoApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.sync,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.backup-repos.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.backup-repos.$.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const ProjectBackupRepoCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useUpdatePassword,
    useCleanup,
    useSync,
    useDeleteOne,
});

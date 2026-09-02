import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectBackupRepoApi } from "~/projects/api/hooks";
import type {
    ProjectBackupRepo_DeleteOne_Req,
    ProjectBackupRepo_DeleteOne_Res,
    ProjectBackupRepo_UpdateStatus_Req,
    ProjectBackupRepo_UpdateStatus_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

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

export const ProjectBackupRepoCommands = Object.freeze({
    useUpdateStatus,
    useDeleteOne,
});

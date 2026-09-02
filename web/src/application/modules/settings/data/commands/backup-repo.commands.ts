import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useBackupRepoApi } from "~/settings/api/hooks";
import type {
    BackupRepo_DeleteOne_Req,
    BackupRepo_DeleteOne_Res,
    BackupRepo_UpdateStatus_Req,
    BackupRepo_UpdateStatus_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

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

export const BackupRepoCommands = Object.freeze({
    useUpdateStatus,
    useDeleteOne,
});

import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useBackupSnapshotApi } from "~/settings/api/hooks";
import type {
    BackupSnapshot_DeleteOne_Req,
    BackupSnapshot_DeleteOne_Res,
    BackupSnapshot_Restore_Req,
    BackupSnapshot_Restore_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type DeleteOneReq = BackupSnapshot_DeleteOne_Req["data"];
type DeleteOneRes = BackupSnapshot_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useBackupSnapshotApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["settings.backup-snapshots.find-many-paginated"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type RestoreReq = BackupSnapshot_Restore_Req["data"];
type RestoreRes = BackupSnapshot_Restore_Res;
type RestoreOptions = Omit<UseMutationOptions<RestoreRes, Error, RestoreReq>, "mutationFn">;

function useRestore(options: RestoreOptions = {}) {
    const { mutations } = useBackupSnapshotApi();

    return useMutation({
        mutationFn: mutations.restore,
        ...options,
    });
}

export const BackupSnapshotCommands = Object.freeze({
    useDeleteOne,
    useRestore,
});

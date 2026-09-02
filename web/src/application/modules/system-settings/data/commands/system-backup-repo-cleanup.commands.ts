import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSystemBackupRepoCleanupApi } from "~/system-settings/api/hooks";
import type {
    SystemBackupRepoCleanup_Execute_Req,
    SystemBackupRepoCleanup_Execute_Res,
    SystemBackupRepoCleanup_UpdateOne_Req,
    SystemBackupRepoCleanup_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = SystemBackupRepoCleanup_UpdateOne_Req["data"];
type UpdateOneRes = SystemBackupRepoCleanup_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;
type ExecuteReq = SystemBackupRepoCleanup_Execute_Req["data"];
type ExecuteRes = SystemBackupRepoCleanup_Execute_Res;
type ExecuteOptions = Omit<UseMutationOptions<ExecuteRes, Error, ExecuteReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useSystemBackupRepoCleanupApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.backup-repo-cleanup.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

function useExecute({ onSuccess, ...options }: ExecuteOptions = {}) {
    const { mutations } = useSystemBackupRepoCleanupApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.execute,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.backup-repo-cleanup.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const SystemBackupRepoCleanupCommands = Object.freeze({
    useUpdateOne,
    useExecute,
});

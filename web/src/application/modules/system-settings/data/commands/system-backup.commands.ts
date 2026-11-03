import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSystemBackupApi } from "~/system-settings/api/hooks";
import type {
    SystemBackup_Execute_Req,
    SystemBackup_Execute_Res,
    SystemBackup_UpdateOne_Req,
    SystemBackup_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = SystemBackup_UpdateOne_Req["data"];
type UpdateOneRes = SystemBackup_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;
type ExecuteReq = SystemBackup_Execute_Req["data"];
type ExecuteRes = SystemBackup_Execute_Res;
type ExecuteOptions = Omit<UseMutationOptions<ExecuteRes, Error, ExecuteReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useSystemBackupApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.backup.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

function useExecute({ onSuccess, ...options }: ExecuteOptions = {}) {
    const { mutations } = useSystemBackupApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.execute,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.backup.find-one"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.backup-files.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.backup-files.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const SystemBackupCommands = Object.freeze({
    useUpdateOne,
    useExecute,
});

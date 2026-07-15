import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSystemBackupFileApi } from "~/system-settings/api/hooks";
import type { SystemBackupFile_DeleteOne_Req, SystemBackupFile_DeleteOne_Res } from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type DeleteOneReq = SystemBackupFile_DeleteOne_Req["data"];
type DeleteOneRes = SystemBackupFile_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useSystemBackupFileApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.backup-files.find-many-paginated"]],
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const SystemBackupFileCommands = Object.freeze({
    useDeleteOne,
});

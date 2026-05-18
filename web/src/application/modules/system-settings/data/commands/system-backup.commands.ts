import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSystemBackupApi } from "~/system-settings/api/hooks";
import type { SystemBackup_UpdateOne_Req, SystemBackup_UpdateOne_Res } from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = SystemBackup_UpdateOne_Req["data"];
type UpdateOneRes = SystemBackup_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

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

export const SystemBackupCommands = Object.freeze({
    useUpdateOne,
});

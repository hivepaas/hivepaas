import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSystemCleanupApi } from "~/system-settings/api/hooks";
import type {
    SystemCleanup_Execute_Req,
    SystemCleanup_Execute_Res,
    SystemCleanup_UpdateOne_Req,
    SystemCleanup_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = SystemCleanup_UpdateOne_Req["data"];
type UpdateOneRes = SystemCleanup_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;
type ExecuteReq = SystemCleanup_Execute_Req["data"];
type ExecuteRes = SystemCleanup_Execute_Res;
type ExecuteOptions = Omit<UseMutationOptions<ExecuteRes, Error, ExecuteReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useSystemCleanupApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.cleanup.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

function useExecute({ onSuccess, ...options }: ExecuteOptions = {}) {
    const { mutations } = useSystemCleanupApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.execute,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.cleanup.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const SystemCleanupCommands = Object.freeze({
    useUpdateOne,
    useExecute,
});

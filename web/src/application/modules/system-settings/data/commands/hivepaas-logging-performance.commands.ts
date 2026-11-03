import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useHivePaaSLoggingPerformanceApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSLoggingPerformance_UpdateOne_Req,
    HivePaaSLoggingPerformance_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = HivePaaSLoggingPerformance_UpdateOne_Req["data"];
type UpdateOneRes = HivePaaSLoggingPerformance_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useHivePaaSLoggingPerformanceApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.hivepaas.logging-performance.find-one"]],
            });
            // Saved with the logging settings: their version moved too, and their page must not save over it.
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.hivepaas.logging.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const HivePaaSLoggingPerformanceCommands = Object.freeze({
    useUpdateOne,
});

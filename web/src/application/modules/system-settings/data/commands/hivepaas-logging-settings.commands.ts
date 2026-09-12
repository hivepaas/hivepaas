import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useHivePaaSLoggingSettingsApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSLoggingSettings_UpdateOne_Req,
    HivePaaSLoggingSettings_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = HivePaaSLoggingSettings_UpdateOne_Req["data"];
type UpdateOneRes = HivePaaSLoggingSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useHivePaaSLoggingSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.hivepaas.logging.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const HivePaaSLoggingSettingsCommands = Object.freeze({
    useUpdateOne,
});

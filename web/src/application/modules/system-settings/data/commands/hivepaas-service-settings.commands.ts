import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useHivePaaSServiceSettingsApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSServiceSettings_UpdateOne_Req,
    HivePaaSServiceSettings_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = HivePaaSServiceSettings_UpdateOne_Req["data"];
type UpdateOneRes = HivePaaSServiceSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useHivePaaSServiceSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.hivepaas.service-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const HivePaaSServiceSettingsCommands = Object.freeze({
    useUpdateOne,
});

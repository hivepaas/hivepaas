import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useHivePaaSRoutingSettingsApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSRoutingSettings_UpdateOne_Req,
    HivePaaSRoutingSettings_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = HivePaaSRoutingSettings_UpdateOne_Req["data"];
type UpdateOneRes = HivePaaSRoutingSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useHivePaaSRoutingSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.hivepaas.routing-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const HivePaaSRoutingSettingsCommands = Object.freeze({
    useUpdateOne,
});

export { HivePaaSRoutingSettingsCommands as HivePaaSHttpSettingsCommands };

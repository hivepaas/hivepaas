import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useHivePaaSHttpSettingsApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSHttpSettings_UpdateOne_Req,
    HivePaaSHttpSettings_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = HivePaaSHttpSettings_UpdateOne_Req["data"];
type UpdateOneRes = HivePaaSHttpSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useHivePaaSHttpSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.hivepaas.http-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const HivePaaSHttpSettingsCommands = Object.freeze({
    useUpdateOne,
});

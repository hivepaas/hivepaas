import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useLocalPaaSServiceSettingsApi } from "~/system-settings/api/hooks";
import type {
    LocalPaaSServiceSettings_UpdateOne_Req,
    LocalPaaSServiceSettings_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = LocalPaaSServiceSettings_UpdateOne_Req["data"];
type UpdateOneRes = LocalPaaSServiceSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useLocalPaaSServiceSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.localpaas.service-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const LocalPaaSServiceSettingsCommands = Object.freeze({
    useUpdateOne,
});

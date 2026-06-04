import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useTraefikServiceSettingsApi } from "~/system-settings/api/hooks";
import type {
    TraefikServiceSettings_UpdateOne_Req,
    TraefikServiceSettings_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = TraefikServiceSettings_UpdateOne_Req["data"];
type UpdateOneRes = TraefikServiceSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useTraefikServiceSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.traefik.service-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const TraefikServiceSettingsCommands = Object.freeze({
    useUpdateOne,
});

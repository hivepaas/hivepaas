import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useTraefikConfigOptionsApi } from "~/system-settings/api/hooks";
import type {
    TraefikConfigOptions_UpdateOne_Req,
    TraefikConfigOptions_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = TraefikConfigOptions_UpdateOne_Req["data"];
type UpdateOneRes = TraefikConfigOptions_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useTraefikConfigOptionsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.traefik.config-options.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const TraefikConfigOptionsCommands = Object.freeze({
    useUpdateOne,
});

import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAppPlacementSettingsApi } from "~/settings/api/hooks";
import type {
    AppPlacementSettings_UpdateOne_Req,
    AppPlacementSettings_UpdateOne_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type UpdateOneReq = AppPlacementSettings_UpdateOne_Req["data"];
type UpdateOneRes = AppPlacementSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppPlacementSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.app-placement-settings.find-one"]],
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppPlacementSettingsCommands = Object.freeze({
    useUpdateOne,
});

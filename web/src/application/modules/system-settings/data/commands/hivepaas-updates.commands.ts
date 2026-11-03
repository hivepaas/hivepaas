import { type UseMutationOptions, useMutation } from "@tanstack/react-query";
import { useHivePaaSUpdatesApi } from "~/system-settings/api/hooks";
import type { HivePaaSUpdates_Update_Req, HivePaaSUpdates_Update_Res } from "~/system-settings/api/services";

type UpdateReq = HivePaaSUpdates_Update_Req["data"];
type UpdateOptions = Omit<UseMutationOptions<HivePaaSUpdates_Update_Res, Error, UpdateReq>, "mutationFn">;

function useUpdate(options: UpdateOptions = {}) {
    const { mutations } = useHivePaaSUpdatesApi();

    return useMutation({
        mutationFn: mutations.update,
        ...options,
    });
}

export const HivePaaSUpdatesCommands = Object.freeze({
    useUpdate,
});

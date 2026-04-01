import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppNetworkSettingsApi } from "../../../api/hooks/project-apps";
import { type AppNetworkSettings_UpdateOne_Req, type AppNetworkSettings_UpdateOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

type UpdateOneReq = AppNetworkSettings_UpdateOne_Req["data"];
type UpdateOneRes = AppNetworkSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppNetworkSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [
                    QK["projects.apps.network-settings.$.find-one"],
                    { projectID: request.projectID, appID: request.appID },
                ],
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppNetworkSettingsCommands = Object.freeze({
    useUpdateOne,
});

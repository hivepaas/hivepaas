import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppResourceSettingsApi } from "../../../api/hooks/project-apps";
import { type AppResourceSettings_UpdateOne_Req, type AppResourceSettings_UpdateOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

type UpdateOneReq = AppResourceSettings_UpdateOne_Req["data"];
type UpdateOneRes = AppResourceSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppResourceSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [
                    QK["projects.apps.resource-settings.$.find-one"],
                    { projectID: request.projectID, appID: request.appID },
                ],
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppResourceSettingsCommands = Object.freeze({
    useUpdateOne,
});

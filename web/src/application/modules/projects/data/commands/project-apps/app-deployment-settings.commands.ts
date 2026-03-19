import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppDeploymentSettingsApi } from "../../../api/hooks/project-apps";
import {
    type AppDeploymentSettings_UpdateOne_Req,
    type AppDeploymentSettings_UpdateOne_Res,
} from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

type UpdateOneReq = AppDeploymentSettings_UpdateOne_Req["data"];
type UpdateOneRes = AppDeploymentSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppDeploymentSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [
                    QK["projects.apps.deployment-settings.$.find-one"],
                    { projectID: request.projectID, appID: request.appID },
                ],
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppDeploymentSettingsCommands = Object.freeze({
    useUpdateOne,
});

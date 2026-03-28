import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppContainerSettingsApi } from "../../../api/hooks/project-apps";
import {
    type AppContainerSettings_UpdateOne_Req,
    type AppContainerSettings_UpdateOne_Res,
} from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

type UpdateOneReq = AppContainerSettings_UpdateOne_Req["data"];
type UpdateOneRes = AppContainerSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppContainerSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [
                    QK["projects.apps.container-settings.$.find-one"],
                    { projectID: request.projectID, appID: request.appID },
                ],
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppContainerSettingsCommands = Object.freeze({
    useUpdateOne,
});

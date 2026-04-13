import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppStorageSettingsApi } from "../../../api/hooks/project-apps";
import { type AppStorageSettings_UpdateOne_Req, type AppStorageSettings_UpdateOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

type UpdateOneReq = AppStorageSettings_UpdateOne_Req["data"];
type UpdateOneRes = AppStorageSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppStorageSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [
                    QK["projects.apps.storage-settings.$.find-one"],
                    { projectID: request.projectID, appID: request.appID },
                ],
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppStorageSettingsCommands = Object.freeze({
    useUpdateOne,
});

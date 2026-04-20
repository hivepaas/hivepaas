import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppHttpSettingsApi } from "../../../api/hooks/project-apps";
import { type AppHttpSettings_UpdateOne_Req, type AppHttpSettings_UpdateOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

type UpdateOneReq = AppHttpSettings_UpdateOne_Req["data"];
type UpdateOneRes = AppHttpSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppHttpSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [
                    QK["projects.apps.http-settings.$.find-one"],
                    { projectID: request.projectID, appID: request.appID },
                ],
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppHttpSettingsCommands = Object.freeze({
    useUpdateOne,
});

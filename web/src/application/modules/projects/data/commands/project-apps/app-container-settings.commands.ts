import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppContainerSettingsApi } from "../../../api/hooks/project-apps";
import {
    type AppContainerSettings_CheckPort_Req,
    type AppContainerSettings_CheckPort_Res,
    type AppContainerSettings_UpdateOne_Req,
    type AppContainerSettings_UpdateOne_Res,
} from "../../../api/services";

import { invalidateSingleAppConfigurationQueries } from "./app-configuration-cache.helpers";

type UpdateOneReq = AppContainerSettings_UpdateOne_Req["data"];
type UpdateOneRes = AppContainerSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppContainerSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            invalidateSingleAppConfigurationQueries(queryClient, {
                projectID: request.projectID,
                appID: request.appID,
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

type CheckPortReq = AppContainerSettings_CheckPort_Req["data"];
type CheckPortRes = AppContainerSettings_CheckPort_Res;
type CheckPortOptions = Omit<UseMutationOptions<CheckPortRes, Error, CheckPortReq>, "mutationFn">;

function useCheckPort(options: CheckPortOptions = {}) {
    const { mutations } = useAppContainerSettingsApi();

    return useMutation({
        mutationFn: mutations.checkPort,
        ...options,
    });
}

export const AppContainerSettingsCommands = Object.freeze({
    useUpdateOne,
    useCheckPort,
});

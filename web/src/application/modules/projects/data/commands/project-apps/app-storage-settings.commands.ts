import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppStorageSettingsApi } from "../../../api/hooks/project-apps";
import {
    type AppStorageSettings_Preflight_Req,
    type AppStorageSettings_Preflight_Res,
    type AppStorageSettings_UpdateOne_Req,
    type AppStorageSettings_UpdateOne_Res,
} from "../../../api/services";

import { invalidateSingleAppConfigurationQueries } from "./app-configuration-cache.helpers";

type UpdateOneReq = AppStorageSettings_UpdateOne_Req["data"];
type UpdateOneRes = AppStorageSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppStorageSettingsApi();
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

type PreflightReq = AppStorageSettings_Preflight_Req["data"];
type PreflightRes = AppStorageSettings_Preflight_Res;
type PreflightOptions = Omit<UseMutationOptions<PreflightRes, Error, PreflightReq>, "mutationFn">;

/** Asks what saving would land on. It writes nothing, so it invalidates nothing. */
function usePreflight(options: PreflightOptions = {}) {
    const { mutations } = useAppStorageSettingsApi();

    return useMutation({
        mutationFn: mutations.preflight,
        ...options,
    });
}

export const AppStorageSettingsCommands = Object.freeze({
    usePreflight,
    useUpdateOne,
});

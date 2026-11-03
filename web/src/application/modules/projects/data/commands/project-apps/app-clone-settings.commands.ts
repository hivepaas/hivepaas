import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppCloneSettingsApi } from "../../../api/hooks/project-apps";
import type {
    AppCloneSettings_Execute_Req,
    AppCloneSettings_Execute_Res,
    AppCloneSettings_UpdateOne_Req,
    AppCloneSettings_UpdateOne_Res,
} from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

import { invalidateSingleAppConfigurationQueries } from "./app-configuration-cache.helpers";

type UpdateOneReq = AppCloneSettings_UpdateOne_Req["data"];
type UpdateOneRes = AppCloneSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppCloneSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [
                    QK["projects.apps.clone-settings.$.find-one"],
                    {
                        projectID: request.projectID,
                        env: request.env,
                        appID: request.appID,
                    },
                ],
            });
            invalidateSingleAppConfigurationQueries(queryClient, {
                projectID: request.projectID,
                appID: request.appID,
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

type ExecuteReq = AppCloneSettings_Execute_Req["data"];
type ExecuteRes = AppCloneSettings_Execute_Res;
type ExecuteOptions = Omit<UseMutationOptions<ExecuteRes, Error, ExecuteReq>, "mutationFn">;

function useExecute({ onSuccess, ...options }: ExecuteOptions = {}) {
    const { mutations } = useAppCloneSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.execute,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.$.find-one-by-id"], { projectID: request.projectID }],
            });
            void queryClient.invalidateQueries({
                queryKey: [
                    QK["projects.apps.clone-settings.$.find-one"],
                    {
                        projectID: request.projectID,
                        env: request.env,
                        appID: request.appID,
                    },
                ],
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppCloneSettingsCommands = Object.freeze({
    useUpdateOne,
    useExecute,
});

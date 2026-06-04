import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectAppsApi } from "~/projects/api/hooks";
import type {
    ProjectApps_CreateOne_Req,
    ProjectApps_CreateOne_Res,
    ProjectApps_DeleteOne_Req,
    ProjectApps_DeleteOne_Res,
    ProjectApps_Deploy_Req,
    ProjectApps_Deploy_Res,
    ProjectApps_Restart_Req,
    ProjectApps_Restart_Res,
    ProjectApps_UpdateOne_Req,
    ProjectApps_UpdateOne_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

import { invalidateSingleAppSummaryQueries } from "./app-configuration-cache.helpers";

/**
 * Create a project app command
 */
type CreateOneReq = ProjectApps_CreateOne_Req["data"];
type CreateOneRes = ProjectApps_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectAppsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

/**
 * Delete a project app command
 */
type DeleteOneReq = ProjectApps_DeleteOne_Req["data"];
type DeleteOneRes = ProjectApps_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectAppsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.$.find-many-paginated"]],
            });
            queryClient.removeQueries({
                queryKey: [
                    QK["projects.apps.$.find-one-by-id"],
                    { projectID: request.projectID, appID: request.appID },
                ],
                exact: true,
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.$.find-one-by-id"], { projectID: request.projectID }],
            });

            if (onSuccess) {
                onSuccess(response, request, ...rest);
            }
        },
        ...options,
    });
}

/**
 * Update a project app command
 */
type UpdateOneReq = ProjectApps_UpdateOne_Req["data"];
type UpdateOneRes = ProjectApps_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectAppsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            invalidateSingleAppSummaryQueries(queryClient, { projectID: request.projectID, appID: request.appID });

            if (onSuccess) {
                onSuccess(response, request, ...rest);
            }
        },
        ...options,
    });
}

/**
 * Deploy a project app command
 */
type DeployReq = ProjectApps_Deploy_Req["data"];
type DeployRes = ProjectApps_Deploy_Res;
type DeployOptions = Omit<UseMutationOptions<DeployRes, Error, DeployReq>, "mutationFn">;

function useDeploy({ onSuccess, ...options }: DeployOptions = {}) {
    const { mutations } = useProjectAppsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deploy,
        onSuccess: (response, request, ...rest) => {
            invalidateSingleAppSummaryQueries(queryClient, { projectID: request.projectID, appID: request.appID });

            if (onSuccess) {
                onSuccess(response, request, ...rest);
            }
        },
        ...options,
    });
}

/**
 * Restart a project app command
 */
type RestartReq = ProjectApps_Restart_Req["data"];
type RestartRes = ProjectApps_Restart_Res;
type RestartOptions = Omit<UseMutationOptions<RestartRes, Error, RestartReq>, "mutationFn">;

function useRestart({ onSuccess, ...options }: RestartOptions = {}) {
    const { mutations } = useProjectAppsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.restart,
        onSuccess: (response, request, ...rest) => {
            invalidateSingleAppSummaryQueries(queryClient, { projectID: request.projectID, appID: request.appID });

            if (onSuccess) {
                onSuccess(response, request, ...rest);
            }
        },
        ...options,
    });
}

export const ProjectAppsCommands = Object.freeze({
    useCreateOne,
    useDeleteOne,
    useUpdateOne,
    useDeploy,
    useRestart,
});

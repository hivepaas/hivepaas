import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectAppsApi } from "~/projects/api/hooks";
import type {
    ProjectApps_Copy_Req,
    ProjectApps_Copy_Res,
    ProjectApps_CreateOne_Req,
    ProjectApps_CreateOne_Res,
    ProjectApps_DeleteOne_Req,
    ProjectApps_DeleteOne_Res,
    ProjectApps_Deploy_Req,
    ProjectApps_Deploy_Res,
    ProjectApps_Restart_Req,
    ProjectApps_Restart_Res,
    ProjectApps_SetRunning_Req,
    ProjectApps_SetRunning_Res,
    ProjectApps_UpdateOne_Req,
    ProjectApps_UpdateOne_Res,
    ProjectApps_UpdateStatus_Req,
    ProjectApps_UpdateStatus_Res,
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
 * Copy a project app command
 */
type CopyReq = ProjectApps_Copy_Req["data"];
type CopyRes = ProjectApps_Copy_Res;
type CopyOptions = Omit<UseMutationOptions<CopyRes, Error, CopyReq>, "mutationFn">;

function useCopy({ onSuccess, ...options }: CopyOptions = {}) {
    const { mutations } = useProjectAppsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.copy,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.$.find-many-paginated"]],
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
 * Update a project app status command
 */
type UpdateStatusReq = ProjectApps_UpdateStatus_Req["data"];
type UpdateStatusRes = ProjectApps_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useProjectAppsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
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
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.deployments.$.find-many-paginated"]],
            });
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

/**
 * Set project app running status command
 */
type SetRunningReq = ProjectApps_SetRunning_Req["data"];
type SetRunningRes = ProjectApps_SetRunning_Res;
type SetRunningOptions = Omit<UseMutationOptions<SetRunningRes, Error, SetRunningReq>, "mutationFn">;

function useSetRunning({ onSuccess, ...options }: SetRunningOptions = {}) {
    const { mutations } = useProjectAppsApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.setRunning,
        onSuccess: (response, request, ...rest) => {
            invalidateSingleAppSummaryQueries(queryClient, { projectID: request.projectID, appID: request.appID });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.service-tasks.$.find-many"]],
            });

            if (onSuccess) {
                onSuccess(response, request, ...rest);
            }
        },
        ...options,
    });
}

export const ProjectAppsCommands = Object.freeze({
    useCreateOne,
    useCopy,
    useDeleteOne,
    useUpdateOne,
    useUpdateStatus,
    useDeploy,
    useRestart,
    useSetRunning,
});

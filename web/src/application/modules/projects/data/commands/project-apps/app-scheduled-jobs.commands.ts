import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAppScheduledJobsApi } from "~/projects/api/hooks/project-apps";
import type {
    AppScheduledJobTasks_Cancel_Req,
    AppScheduledJobTasks_Cancel_Res,
    AppScheduledJobs_CreateOne_Req,
    AppScheduledJobs_CreateOne_Res,
    AppScheduledJobs_DeleteOne_Req,
    AppScheduledJobs_DeleteOne_Res,
    AppScheduledJobs_RunNow_Req,
    AppScheduledJobs_RunNow_Res,
    AppScheduledJobs_UpdateOne_Req,
    AppScheduledJobs_UpdateOne_Res,
    AppScheduledJobs_UpdateStatus_Req,
    AppScheduledJobs_UpdateStatus_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

import { invalidateSingleAppConfigurationQueries } from "./app-configuration-cache.helpers";

type CreateOneReq = AppScheduledJobs_CreateOne_Req["data"];
type CreateOneRes = AppScheduledJobs_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useAppScheduledJobsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
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

type UpdateOneReq = AppScheduledJobs_UpdateOne_Req["data"];
type UpdateOneRes = AppScheduledJobs_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppScheduledJobsApi();
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

type UpdateStatusReq = AppScheduledJobs_UpdateStatus_Req["data"];
type UpdateStatusRes = AppScheduledJobs_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useAppScheduledJobsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
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

type DeleteOneReq = AppScheduledJobs_DeleteOne_Req["data"];
type DeleteOneRes = AppScheduledJobs_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useAppScheduledJobsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
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

type RunNowReq = AppScheduledJobs_RunNow_Req["data"];
type RunNowRes = AppScheduledJobs_RunNow_Res;
type RunNowOptions = Omit<UseMutationOptions<RunNowRes, Error, RunNowReq>, "mutationFn">;

function useRunNow({ onSuccess, ...options }: RunNowOptions = {}) {
    const { mutations } = useAppScheduledJobsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.runNow,
        onSuccess: (response, request, ...rest) => {
            invalidateSingleAppConfigurationQueries(queryClient, {
                projectID: request.projectID,
                appID: request.appID,
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.scheduled-jobs.tasks.$.find-many-paginated"]],
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

type CancelTaskReq = AppScheduledJobTasks_Cancel_Req["data"];
type CancelTaskRes = AppScheduledJobTasks_Cancel_Res;
type CancelTaskOptions = Omit<UseMutationOptions<CancelTaskRes, Error, CancelTaskReq>, "mutationFn">;

function useCancelTask({ onSuccess, ...options }: CancelTaskOptions = {}) {
    const { mutations } = useAppScheduledJobsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.cancelTask,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.scheduled-jobs.tasks.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.scheduled-jobs.tasks.$.find-one-by-id"]],
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppScheduledJobsCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
    useRunNow,
    useCancelTask,
});

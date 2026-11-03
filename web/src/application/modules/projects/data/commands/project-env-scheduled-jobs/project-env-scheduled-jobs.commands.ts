import { type QueryClient, type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useEnvScheduledJobsApi } from "~/projects/api/hooks";
import type {
    EnvScheduledJobs_CreateOne_Req,
    EnvScheduledJobs_CreateOne_Res,
    EnvScheduledJobs_DeleteOne_Req,
    EnvScheduledJobs_DeleteOne_Res,
    EnvScheduledJobs_RunNow_Req,
    EnvScheduledJobs_RunNow_Res,
    EnvScheduledJobs_UpdateOne_Req,
    EnvScheduledJobs_UpdateOne_Res,
    EnvScheduledJobs_UpdateStatus_Req,
    EnvScheduledJobs_UpdateStatus_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

/** An env's list holds its apps' jobs too: their lists are refreshed with it. */
function invalidateScheduledJobs(queryClient: QueryClient): void {
    void queryClient.invalidateQueries({ queryKey: [QK["projects.env.scheduled-jobs.$.find-many-paginated"]] });
    void queryClient.invalidateQueries({ queryKey: [QK["projects.env.scheduled-jobs.$.find-one-by-id"]] });
    void queryClient.invalidateQueries({ queryKey: [QK["projects.apps.scheduled-jobs.$.find-many-paginated"]] });
}

type CreateOneOptions = Omit<
    UseMutationOptions<EnvScheduledJobs_CreateOne_Res, Error, EnvScheduledJobs_CreateOne_Req["data"]>,
    "mutationFn"
>;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useEnvScheduledJobsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (...args) => {
            invalidateScheduledJobs(queryClient);
            onSuccess?.(...args);
        },
        ...options,
    });
}

type UpdateOneOptions = Omit<
    UseMutationOptions<EnvScheduledJobs_UpdateOne_Res, Error, EnvScheduledJobs_UpdateOne_Req["data"]>,
    "mutationFn"
>;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useEnvScheduledJobsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (...args) => {
            invalidateScheduledJobs(queryClient);
            onSuccess?.(...args);
        },
        ...options,
    });
}

type UpdateStatusOptions = Omit<
    UseMutationOptions<EnvScheduledJobs_UpdateStatus_Res, Error, EnvScheduledJobs_UpdateStatus_Req["data"]>,
    "mutationFn"
>;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useEnvScheduledJobsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (...args) => {
            invalidateScheduledJobs(queryClient);
            onSuccess?.(...args);
        },
        ...options,
    });
}

type DeleteOneOptions = Omit<
    UseMutationOptions<EnvScheduledJobs_DeleteOne_Res, Error, EnvScheduledJobs_DeleteOne_Req["data"]>,
    "mutationFn"
>;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useEnvScheduledJobsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (...args) => {
            invalidateScheduledJobs(queryClient);
            onSuccess?.(...args);
        },
        ...options,
    });
}

type RunNowOptions = Omit<
    UseMutationOptions<EnvScheduledJobs_RunNow_Res, Error, EnvScheduledJobs_RunNow_Req["data"]>,
    "mutationFn"
>;

function useRunNow(options: RunNowOptions = {}) {
    const { mutations } = useEnvScheduledJobsApi();

    return useMutation({
        mutationFn: mutations.runNow,
        ...options,
    });
}

export const EnvScheduledJobsCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
    useRunNow,
});

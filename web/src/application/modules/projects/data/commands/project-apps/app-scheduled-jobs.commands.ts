import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAppScheduledJobsApi } from "~/projects/api/hooks/project-apps";
import type {
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

function invalidateScheduledJobs(queryClient: ReturnType<typeof useQueryClient>) {
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.scheduled-jobs.$.find-many-paginated"]],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["projects.apps.scheduled-jobs.$.find-one-by-id"]],
    });
}

type CreateOneReq = AppScheduledJobs_CreateOne_Req["data"];
type CreateOneRes = AppScheduledJobs_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useAppScheduledJobsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.scheduled-jobs.$.find-many-paginated"]],
            });
            onSuccess?.(response, ...rest);
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
        onSuccess: (response, ...rest) => {
            invalidateScheduledJobs(queryClient);
            onSuccess?.(response, ...rest);
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
        onSuccess: (response, ...rest) => {
            invalidateScheduledJobs(queryClient);
            onSuccess?.(response, ...rest);
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
        onSuccess: (response, ...rest) => {
            invalidateScheduledJobs(queryClient);
            onSuccess?.(response, ...rest);
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
        onSuccess: (response, ...rest) => {
            invalidateScheduledJobs(queryClient);
            onSuccess?.(response, ...rest);
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
});

import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAppScheduledJobsApi } from "~/projects/api";
import type {
    AppScheduledJobTasks_FindManyPaginated_Req,
    AppScheduledJobTasks_FindManyPaginated_Res,
    AppScheduledJobTasks_FindOneById_Req,
    AppScheduledJobTasks_FindOneById_Res,
    AppScheduledJobTasks_GetLogs_Req,
    AppScheduledJobTasks_GetLogs_Res,
    AppScheduledJobs_FindManyPaginated_Req,
    AppScheduledJobs_FindManyPaginated_Res,
    AppScheduledJobs_FindOneById_Req,
    AppScheduledJobs_FindOneById_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

type FindManyPaginatedReq = AppScheduledJobs_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = AppScheduledJobs_FindManyPaginated_Res;
type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useAppScheduledJobsApi();

    return useQuery({
        queryKey: [QK["projects.apps.scheduled-jobs.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindOneByIdReq = AppScheduledJobs_FindOneById_Req["data"];
type FindOneByIdRes = AppScheduledJobs_FindOneById_Res;
type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useAppScheduledJobsApi();

    return useQuery({
        queryKey: [QK["projects.apps.scheduled-jobs.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

type FindTasksManyPaginatedReq = AppScheduledJobTasks_FindManyPaginated_Req["data"];
type FindTasksManyPaginatedRes = AppScheduledJobTasks_FindManyPaginated_Res;
type FindTasksManyPaginatedOptions = Omit<UseQueryOptions<FindTasksManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindTasksManyPaginated(request: FindTasksManyPaginatedReq, options: FindTasksManyPaginatedOptions = {}) {
    const { queries } = useAppScheduledJobsApi();

    return useQuery({
        queryKey: [QK["projects.apps.scheduled-jobs.tasks.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findTasksManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindTaskByIdReq = AppScheduledJobTasks_FindOneById_Req["data"];
type FindTaskByIdRes = AppScheduledJobTasks_FindOneById_Res;
type FindTaskByIdOptions = Omit<UseQueryOptions<FindTaskByIdRes>, "queryKey" | "queryFn">;

function useFindTaskById(request: FindTaskByIdReq, options: FindTaskByIdOptions = {}) {
    const { queries } = useAppScheduledJobsApi();

    return useQuery({
        queryKey: [QK["projects.apps.scheduled-jobs.tasks.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findTaskById(request, signal),
        ...options,
    });
}

type GetTaskLogsReq = AppScheduledJobTasks_GetLogs_Req["data"];
type GetTaskLogsRes = AppScheduledJobTasks_GetLogs_Res;
type GetTaskLogsOptions = Omit<UseQueryOptions<GetTaskLogsRes>, "queryKey" | "queryFn">;

function useGetTaskLogs(request: GetTaskLogsReq, options: GetTaskLogsOptions = {}) {
    const { queries } = useAppScheduledJobsApi();

    return useQuery({
        queryKey: [QK["projects.apps.scheduled-jobs.tasks.$.get-logs"], request],
        queryFn: ({ signal }) => queries.getTaskLogs(request, signal),
        ...options,
    });
}

export const AppScheduledJobsQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
    useFindTasksManyPaginated,
    useFindTaskById,
    useGetTaskLogs,
});

import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useEnvScheduledJobsApi } from "~/projects/api/hooks";
import type {
    EnvScheduledJobs_FindManyPaginated_Req,
    EnvScheduledJobs_FindManyPaginated_Res,
    EnvScheduledJobs_FindOneById_Req,
    EnvScheduledJobs_FindOneById_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

type FindManyPaginatedReq = EnvScheduledJobs_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = EnvScheduledJobs_FindManyPaginated_Res;
type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useEnvScheduledJobsApi();

    return useQuery({
        queryKey: [QK["projects.env.scheduled-jobs.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindOneByIdReq = EnvScheduledJobs_FindOneById_Req["data"];
type FindOneByIdRes = EnvScheduledJobs_FindOneById_Res;
type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useEnvScheduledJobsApi();

    return useQuery({
        queryKey: [QK["projects.env.scheduled-jobs.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const EnvScheduledJobsQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

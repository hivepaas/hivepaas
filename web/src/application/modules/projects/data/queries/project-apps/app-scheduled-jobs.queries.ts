import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAppScheduledJobsApi } from "~/projects/api";
import type {
    AppScheduledJobs_FindManyPaginated_Req,
    AppScheduledJobs_FindManyPaginated_Res,
    AppScheduledJobs_FindOneById_Req,
    AppScheduledJobs_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyPaginatedReq = AppScheduledJobs_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = AppScheduledJobs_FindManyPaginated_Res;
type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useAppScheduledJobsApi();

    return useQuery({
        queryKey: [QK["projects.apps.scheduled-jobs.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
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

export const AppScheduledJobsQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

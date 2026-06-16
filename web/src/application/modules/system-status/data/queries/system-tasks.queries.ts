import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useSystemTasksApi } from "~/system-status/api";
import type {
    SystemTasks_FindManyPaginated_Req,
    SystemTasks_FindManyPaginated_Res,
    SystemTasks_FindOneById_Req,
    SystemTasks_FindOneById_Res,
} from "~/system-status/api/services";
import { QK } from "~/system-status/data/constants";

type FindManyPaginatedReq = SystemTasks_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = SystemTasks_FindManyPaginated_Res;
type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useSystemTasksApi();

    return useQuery({
        queryKey: [QK["system-status.tasks.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = SystemTasks_FindOneById_Req["data"];
type FindOneByIdRes = SystemTasks_FindOneById_Res;
type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useSystemTasksApi();

    return useQuery({
        queryKey: [QK["system-status.tasks.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        enabled: Boolean(request.taskID),
        ...options,
    });
}

export const SystemTasksQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

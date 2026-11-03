import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useSystemTasksApi } from "~/operations/api";
import type {
    SystemTasks_FindManyPaginated_Req,
    SystemTasks_FindManyPaginated_Res,
    SystemTasks_FindOneById_Req,
    SystemTasks_FindOneById_Res,
    SystemTasks_FindTargetObjects_Req,
    SystemTasks_FindTargetObjects_Res,
    SystemTasks_FindTypes_Req,
    SystemTasks_FindTypes_Res,
} from "~/operations/api/services";
import { QK } from "~/operations/data/constants";

type FindManyPaginatedReq = SystemTasks_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = SystemTasks_FindManyPaginated_Res;
type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useSystemTasksApi();

    return useQuery({
        queryKey: [QK["operations.tasks.find-many-paginated"], request],
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
        queryKey: [QK["operations.tasks.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        enabled: Boolean(request.taskID),
        ...options,
    });
}

type FindTypesReq = SystemTasks_FindTypes_Req["data"];
type FindTypesRes = SystemTasks_FindTypes_Res;
type FindTypesOptions = Omit<UseQueryOptions<FindTypesRes>, "queryKey" | "queryFn">;

function useFindTypes(request: FindTypesReq = {}, options: FindTypesOptions = {}) {
    const { queries } = useSystemTasksApi();

    return useQuery({
        queryKey: [QK["operations.tasks.find-types"], request],
        queryFn: ({ signal }) => queries.findTypes(request, signal),
        staleTime: Infinity,
        gcTime: Infinity,
        refetchOnMount: false,
        refetchOnWindowFocus: false,
        refetchOnReconnect: false,
        ...options,
    });
}

type FindTargetObjectsReq = SystemTasks_FindTargetObjects_Req["data"];
type FindTargetObjectsRes = SystemTasks_FindTargetObjects_Res;
type FindTargetObjectsOptions = Omit<UseQueryOptions<FindTargetObjectsRes>, "queryKey" | "queryFn">;

function useFindTargetObjects(request: FindTargetObjectsReq = {}, options: FindTargetObjectsOptions = {}) {
    const { queries } = useSystemTasksApi();

    return useQuery({
        queryKey: [QK["operations.tasks.find-target-objects"], request],
        queryFn: ({ signal }) => queries.findTargetObjects(request, signal),
        staleTime: 60_000,
        ...options,
    });
}

export const SystemTasksQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
    useFindTypes,
    useFindTargetObjects,
});

import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAuditLogsApi } from "~/system-status/api";
import type {
    AuditLogs_FindManyPaginated_Req,
    AuditLogs_FindManyPaginated_Res,
    AuditLogs_FindOneById_Req,
    AuditLogs_FindOneById_Res,
    AuditLogs_FindTypes_Req,
    AuditLogs_FindTypes_Res,
} from "~/system-status/api/services";
import { QK } from "~/system-status/data/constants";

type FindManyPaginatedReq = AuditLogs_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = AuditLogs_FindManyPaginated_Res;
type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useAuditLogsApi();

    return useQuery({
        queryKey: [QK["system-status.audit-logs.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = AuditLogs_FindOneById_Req["data"];
type FindOneByIdRes = AuditLogs_FindOneById_Res;
type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useAuditLogsApi();

    return useQuery({
        queryKey: [QK["system-status.audit-logs.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        enabled: Boolean(request.itemID),
        staleTime: Infinity,
        gcTime: Infinity,
        refetchOnMount: false,
        refetchOnWindowFocus: false,
        refetchOnReconnect: false,
        ...options,
    });
}

type FindTypesReq = AuditLogs_FindTypes_Req["data"];
type FindTypesRes = AuditLogs_FindTypes_Res;
type FindTypesOptions = Omit<UseQueryOptions<FindTypesRes>, "queryKey" | "queryFn">;

function useFindTypes(request: FindTypesReq = {}, options: FindTypesOptions = {}) {
    const { queries } = useAuditLogsApi();

    return useQuery({
        queryKey: [QK["system-status.audit-logs.find-types"], request],
        queryFn: ({ signal }) => queries.findTypes(request, signal),
        staleTime: Infinity,
        gcTime: Infinity,
        refetchOnMount: false,
        refetchOnWindowFocus: false,
        refetchOnReconnect: false,
        ...options,
    });
}

export const AuditLogsQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
    useFindTypes,
});

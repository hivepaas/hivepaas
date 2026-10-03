import { type UseQueryOptions, useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useAppLogsApi } from "~/projects/api";
import type {
    AppLogs_GetDependencyMetrics_Req,
    AppLogs_GetFunctionMetrics_Req,
    AppLogs_GetHistory_Req,
    AppLogs_GetHttpMetrics_Req,
    AppLogs_GetInfo_Req,
    AppLogs_GetInfo_Res,
    AppLogs_GetLogs_Req,
    AppLogs_GetLogs_Res,
    AppLogs_GetResourceMetrics_Req,
    AppLogs_GetRouteMetrics_Req,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type GetInfoReq = AppLogs_GetInfo_Req["data"];
type GetInfoRes = AppLogs_GetInfo_Res;
type GetInfoOptions = Omit<UseQueryOptions<GetInfoRes>, "queryKey" | "queryFn">;

function useGetInfo(request: GetInfoReq, options: GetInfoOptions = {}) {
    const { queries } = useAppLogsApi();

    return useQuery({
        queryKey: [QK["projects.apps.logs.$.get-info"], request],
        queryFn: ({ signal }) => queries.getInfo(request, signal),
        ...options,
    });
}

type GetLogsReq = AppLogs_GetLogs_Req["data"];
type GetLogsRes = AppLogs_GetLogs_Res;
type GetLogsOptions = Omit<UseQueryOptions<GetLogsRes>, "queryKey" | "queryFn">;

function useGetLogs(request: GetLogsReq, options: GetLogsOptions = {}) {
    const { queries } = useAppLogsApi();

    return useQuery({
        queryKey: [QK["projects.apps.logs.$.get-logs"], request],
        queryFn: ({ signal }) => queries.getLogs(request, signal),
        ...options,
    });
}

type GetHistoryReq = AppLogs_GetHistory_Req["data"];

/**
 * Stored logs, newest page first. The first page ends where the caller's window
 * ends; each further page is older, ending at the previous page's `nextEnd`,
 * passed back as the server wrote it.
 */
function useGetHistory(request: GetHistoryReq, options: { enabled?: boolean } = {}) {
    const { queries } = useAppLogsApi();

    return useInfiniteQuery({
        queryKey: [QK["projects.apps.logs.$.get-history"], request],
        queryFn: ({ signal, pageParam }) => queries.getHistory({ ...request, end: pageParam ?? request.end }, signal),
        initialPageParam: undefined as string | undefined,
        getNextPageParam: last => (last.data.truncated && last.data.nextEnd ? last.data.nextEnd : undefined),
        ...options,
    });
}

/** A function's calls over a range ending now, counted from its logs. */
function useGetFunctionMetrics(request: AppLogs_GetFunctionMetrics_Req["data"], options: { enabled?: boolean } = {}) {
    const { queries } = useAppLogsApi();

    return useQuery({
        queryKey: [QK["projects.apps.function-metrics.$.get"], request],
        queryFn: ({ signal }) => queries.getFunctionMetrics(request, signal),
        ...options,
    });
}

function useGetHttpMetrics(request: AppLogs_GetHttpMetrics_Req["data"], options: { enabled?: boolean } = {}) {
    const { queries } = useAppLogsApi();

    return useQuery({
        queryKey: [QK["projects.apps.http-metrics.$.get"], request],
        queryFn: ({ signal }) => queries.getHttpMetrics(request, signal),
        ...options,
    });
}

function useGetResourceMetrics(request: AppLogs_GetResourceMetrics_Req["data"], options: { enabled?: boolean } = {}) {
    const { queries } = useAppLogsApi();

    return useQuery({
        queryKey: [QK["projects.apps.resource-metrics.$.get"], request],
        queryFn: ({ signal }) => queries.getResourceMetrics(request, signal),
        ...options,
    });
}

/** What an app served over a range ending now, by route, as OBI saw it in its containers. */
function useGetRouteMetrics(request: AppLogs_GetRouteMetrics_Req["data"], options: { enabled?: boolean } = {}) {
    const { queries } = useAppLogsApi();

    return useQuery({
        queryKey: [QK["projects.apps.route-metrics.$.get"], request],
        queryFn: ({ signal }) => queries.getRouteMetrics(request, signal),
        ...options,
    });
}

/** What an app called over a range ending now - other apps, databases, outside hosts - as OBI saw it. */
function useGetDependencyMetrics(
    request: AppLogs_GetDependencyMetrics_Req["data"],
    options: { enabled?: boolean } = {},
) {
    const { queries } = useAppLogsApi();

    return useQuery({
        queryKey: [QK["projects.apps.dependency-metrics.$.get"], request],
        queryFn: ({ signal }) => queries.getDependencyMetrics(request, signal),
        ...options,
    });
}

export const AppLogsQueries = Object.freeze({
    useGetInfo,
    useGetLogs,
    useGetHistory,
    useGetFunctionMetrics,
    useGetHttpMetrics,
    useGetResourceMetrics,
    useGetRouteMetrics,
    useGetDependencyMetrics,
});

import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useAppLogsApi } from "~/projects/api";
import type {
    AppLogs_GetInfo_Req,
    AppLogs_GetInfo_Res,
    AppLogs_GetLogs_Req,
    AppLogs_GetLogs_Res,
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

export const AppLogsQueries = Object.freeze({
    useGetInfo,
    useGetLogs,
});

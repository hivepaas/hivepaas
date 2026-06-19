import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useAppTerminalApi } from "~/projects/api";
import type { AppTerminal_GetInfo_Req, AppTerminal_GetInfo_Res } from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type GetInfoReq = AppTerminal_GetInfo_Req["data"];
type GetInfoRes = AppTerminal_GetInfo_Res;
type GetInfoOptions = Omit<UseQueryOptions<GetInfoRes>, "queryKey" | "queryFn">;

function useGetInfo(request: GetInfoReq, options: GetInfoOptions = {}) {
    const { queries } = useAppTerminalApi();

    return useQuery({
        queryKey: [QK["projects.apps.terminal.$.get-info"], request],
        queryFn: ({ signal }) => queries.getInfo(request, signal),
        ...options,
    });
}

export const AppTerminalQueries = Object.freeze({
    useGetInfo,
});

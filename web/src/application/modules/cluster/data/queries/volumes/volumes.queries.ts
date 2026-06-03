import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useClusterVolumesApi } from "~/cluster/api";
import type { ClusterVolumes_List_Req, ClusterVolumes_List_Res } from "~/cluster/api/services";
import { QK } from "~/cluster/data/constants";

type ListReq = ClusterVolumes_List_Req["data"];
type ListRes = ClusterVolumes_List_Res;
type ListOptions = Omit<UseQueryOptions<ListRes>, "queryKey" | "queryFn">;

function useList(request: ListReq, options: ListOptions = {}) {
    const { queries } = useClusterVolumesApi();

    return useQuery({
        queryKey: [QK["volumes.$.list"], request],
        queryFn: ({ signal }) => queries.list(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const ClusterVolumesQueries = Object.freeze({
    useList,
});

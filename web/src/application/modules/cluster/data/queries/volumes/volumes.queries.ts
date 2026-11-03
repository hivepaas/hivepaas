import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useClusterVolumesApi } from "~/cluster/api";
import type {
    ClusterVolumes_FindManyPaginated_Req,
    ClusterVolumes_FindManyPaginated_Res,
    ClusterVolumes_FindOneById_Req,
    ClusterVolumes_FindOneById_Res,
    ClusterVolumes_List_Req,
    ClusterVolumes_List_Res,
} from "~/cluster/api/services";
import { QK } from "~/cluster/data/constants";

type FindManyPaginatedReq = ClusterVolumes_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ClusterVolumes_FindManyPaginated_Res;
type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq = {}, options: FindManyPaginatedOptions = {}) {
    const { queries } = useClusterVolumesApi();

    return useQuery({
        queryKey: [QK["volumes.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = ClusterVolumes_FindOneById_Req["data"];
type FindOneByIdRes = ClusterVolumes_FindOneById_Res;
type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useClusterVolumesApi();

    return useQuery({
        queryKey: [QK["volumes.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

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
    useFindManyPaginated,
    useFindOneById,
    useList,
});

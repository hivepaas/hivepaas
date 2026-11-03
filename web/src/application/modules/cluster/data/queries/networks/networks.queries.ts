import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useClusterNetworksApi } from "~/cluster/api";
import type {
    ClusterNetworks_FindManyPaginated_Req,
    ClusterNetworks_FindManyPaginated_Res,
    ClusterNetworks_FindOneById_Req,
    ClusterNetworks_FindOneById_Res,
} from "~/cluster/api/services";
import { QK } from "~/cluster/data/constants";

type FindManyPaginatedReq = ClusterNetworks_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ClusterNetworks_FindManyPaginated_Res;
type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq = {}, options: FindManyPaginatedOptions = {}) {
    const { queries } = useClusterNetworksApi();

    return useQuery({
        queryKey: [QK["networks.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = ClusterNetworks_FindOneById_Req["data"];
type FindOneByIdRes = ClusterNetworks_FindOneById_Res;
type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useClusterNetworksApi();

    return useQuery({
        queryKey: [QK["networks.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ClusterNetworksQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

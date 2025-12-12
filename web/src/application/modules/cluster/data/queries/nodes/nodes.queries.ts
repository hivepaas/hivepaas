import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useNodesApi } from "~/cluster/api";
import {
    type Nodes_FindManyPaginated_Req,
    type Nodes_FindManyPaginated_Res,
    type Nodes_FindOneById_Req,
    type Nodes_FindOneById_Res,
} from "~/cluster/api/services";
import { QK } from "~/cluster/data/constants";

/**
 * Find many nodes paginated query
 */
type FindManyPaginatedReq = Nodes_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = Nodes_FindManyPaginated_Res;

type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq = {}, options: FindManyPaginatedOptions = {}) {
    const { queries } = useNodesApi();

    return useQuery({
        queryKey: [QK["nodes.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,

        ...options,
    });
}

/**
 * Find one node by id query
 */
type FindOneByIdReq = Nodes_FindOneById_Req["data"];
type FindOneByIdRes = Nodes_FindOneById_Res;

type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useNodesApi();

    return useQuery({
        queryKey: [QK["nodes.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const NodesQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

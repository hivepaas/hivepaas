import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectNetworksApi } from "~/projects/api";
import {
    type ProjectNetworks_FindManyPaginated_Req,
    type ProjectNetworks_FindManyPaginated_Res,
    type ProjectNetworks_FindOneById_Req,
    type ProjectNetworks_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectNetworks_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectNetworks_FindManyPaginated_Res;

type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useProjectNetworksApi();

    return useQuery({
        queryKey: [QK["projects.networks.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = ProjectNetworks_FindOneById_Req["data"];
type FindOneByIdRes = ProjectNetworks_FindOneById_Res;
type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useProjectNetworksApi();

    return useQuery({
        queryKey: [QK["projects.networks.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectNetworksQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectNetworksApi } from "~/projects/api";
import {
    type ProjectNetworks_FindManyPaginated_Req,
    type ProjectNetworks_FindManyPaginated_Res,
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

export const ProjectNetworksQueries = Object.freeze({
    useFindManyPaginated,
});

import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAppDeploymentsApi } from "~/projects/api";
import type {
    AppDeployments_FindManyPaginated_Req,
    AppDeployments_FindManyPaginated_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyPaginatedReq = AppDeployments_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = AppDeployments_FindManyPaginated_Res;
type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useAppDeploymentsApi();

    return useQuery({
        queryKey: [QK["projects.apps.deployments.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppDeploymentsQueries = Object.freeze({
    useFindManyPaginated,
});

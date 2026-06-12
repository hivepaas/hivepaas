import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAppDeploymentsApi } from "~/projects/api";
import type {
    AppDeployments_FindManyPaginated_Req,
    AppDeployments_FindManyPaginated_Res,
    AppDeployments_FindOneById_Req,
    AppDeployments_FindOneById_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

type FindManyPaginatedReq = AppDeployments_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = AppDeployments_FindManyPaginated_Res;
type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useAppDeploymentsApi();

    return useQuery({
        queryKey: [QK["projects.apps.deployments.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindOneByIdReq = AppDeployments_FindOneById_Req["data"];
type FindOneByIdRes = AppDeployments_FindOneById_Res;
type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useAppDeploymentsApi();

    return useQuery({
        queryKey: [QK["projects.apps.deployments.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const AppDeploymentsQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

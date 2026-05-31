import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAppHealthChecksApi } from "~/projects/api";
import type {
    AppHealthChecks_FindManyPaginated_Req,
    AppHealthChecks_FindManyPaginated_Res,
    AppHealthChecks_FindOneById_Req,
    AppHealthChecks_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyPaginatedReq = AppHealthChecks_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = AppHealthChecks_FindManyPaginated_Res;
type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useAppHealthChecksApi();

    return useQuery({
        queryKey: [QK["projects.apps.health-checks.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = AppHealthChecks_FindOneById_Req["data"];
type FindOneByIdRes = AppHealthChecks_FindOneById_Res;
type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useAppHealthChecksApi();

    return useQuery({
        queryKey: [QK["projects.apps.health-checks.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const AppHealthChecksQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

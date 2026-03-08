import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectAppSecretsApi } from "~/projects/api";
import type {
    AppSecrets_FindManyPaginated_Req,
    AppSecrets_FindManyPaginated_Res,
    AppSecrets_FindOneById_Req,
    AppSecrets_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

/**
 * Find many app secrets paginated query
 */
type FindManyPaginatedReq = AppSecrets_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = AppSecrets_FindManyPaginated_Res;

type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useProjectAppSecretsApi();

    return useQuery({
        queryKey: [QK["projects.apps.secrets.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

/**
 * Find one app secret by id query
 */
type FindOneByIdReq = AppSecrets_FindOneById_Req["data"];
type FindOneByIdRes = AppSecrets_FindOneById_Res;

type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useProjectAppSecretsApi();

    return useQuery({
        queryKey: [QK["projects.apps.secrets.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectAppSecretsQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

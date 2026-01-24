import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectSecretsApi } from "~/projects/api";
import type {
    ProjectSecrets_FindManyPaginated_Req,
    ProjectSecrets_FindManyPaginated_Res,
    ProjectSecrets_FindOneById_Req,
    ProjectSecrets_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

/**
 * Find many project secrets paginated query
 */
type FindManyPaginatedReq = ProjectSecrets_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectSecrets_FindManyPaginated_Res;

type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useProjectSecretsApi();

    return useQuery({
        queryKey: [QK["projects.secrets.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

/**
 * Find one project secret by id query
 */
type FindOneByIdReq = ProjectSecrets_FindOneById_Req["data"];
type FindOneByIdRes = ProjectSecrets_FindOneById_Res;

type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useProjectSecretsApi();

    return useQuery({
        queryKey: [QK["projects.secrets.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectSecretsQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

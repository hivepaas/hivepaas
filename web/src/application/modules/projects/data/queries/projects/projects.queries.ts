import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectsApi } from "~/projects/api";
import {
    type Projects_FindManyPaginated_Req,
    type Projects_FindManyPaginated_Res,
    type Projects_FindOneById_Req,
    type Projects_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

/**
 * Find many projects paginated query
 */
type FindManyPaginatedReq = Projects_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = Projects_FindManyPaginated_Res;

type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq = {}, options: FindManyPaginatedOptions = {}) {
    const { queries } = useProjectsApi();

    return useQuery({
        queryKey: [QK["projects.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

/**
 * Find one project by id query
 */
type FindOneByIdReq = Projects_FindOneById_Req["data"];
type FindOneByIdRes = Projects_FindOneById_Res;

type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useProjectsApi();

    return useQuery({
        queryKey: [QK["projects.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectsQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

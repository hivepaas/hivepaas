import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectConfigFilesApi } from "~/projects/api";
import type {
    ProjectConfigFiles_FindManyPaginated_Req,
    ProjectConfigFiles_FindManyPaginated_Res,
    ProjectConfigFiles_FindOneById_Req,
    ProjectConfigFiles_FindOneById_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

/**
 * Find many project config files paginated query
 */
type FindManyPaginatedReq = ProjectConfigFiles_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectConfigFiles_FindManyPaginated_Res;

type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useProjectConfigFilesApi();

    return useQuery({
        queryKey: [QK["projects.config-files.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

/**
 * Find one project config file by id query
 */
type FindOneByIdReq = ProjectConfigFiles_FindOneById_Req["data"];
type FindOneByIdRes = ProjectConfigFiles_FindOneById_Res;

type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useProjectConfigFilesApi();

    return useQuery({
        queryKey: [QK["projects.config-files.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectConfigFilesQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

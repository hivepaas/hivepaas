import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectAppsApi } from "~/projects/api";
import type {
    ProjectApps_FindManyPaginated_Req,
    ProjectApps_FindManyPaginated_Res,
    ProjectApps_FindOneById_Req,
    ProjectApps_FindOneById_Res,
    ProjectApps_PrepareCopy_Req,
    ProjectApps_PrepareCopy_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

/**
 * Find many project apps paginated query
 */
type FindManyPaginatedReq = ProjectApps_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectApps_FindManyPaginated_Res;

type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useProjectAppsApi();

    return useQuery({
        queryKey: [QK["projects.apps.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

/**
 * Find one project app by id query
 */
type FindOneByIdReq = ProjectApps_FindOneById_Req["data"];
type FindOneByIdRes = ProjectApps_FindOneById_Res;

type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useProjectAppsApi();

    return useQuery({
        queryKey: [QK["projects.apps.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

type PrepareCopyReq = ProjectApps_PrepareCopy_Req["data"];
type PrepareCopyRes = ProjectApps_PrepareCopy_Res;

type PrepareCopyOptions = Omit<UseQueryOptions<PrepareCopyRes>, "queryKey" | "queryFn">;

function usePrepareCopy(request: PrepareCopyReq, options: PrepareCopyOptions = {}) {
    const { queries } = useProjectAppsApi();

    return useQuery({
        queryKey: [QK["projects.apps.$.copy.prepare"], request],
        queryFn: ({ signal }) => queries.prepareCopy(request, signal),
        ...options,
    });
}

export const ProjectAppsQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
    usePrepareCopy,
});

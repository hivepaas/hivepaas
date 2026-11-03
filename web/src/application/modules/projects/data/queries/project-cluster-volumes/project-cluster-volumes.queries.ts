import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectClusterVolumesApi } from "~/projects/api";
import type {
    ProjectClusterVolumes_FindManyPaginated_Req,
    ProjectClusterVolumes_FindManyPaginated_Res,
    ProjectClusterVolumes_FindOneById_Req,
    ProjectClusterVolumes_FindOneById_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectClusterVolumes_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectClusterVolumes_FindManyPaginated_Res;
type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useProjectClusterVolumesApi();

    return useQuery({
        queryKey: [QK["projects.cluster-volumes.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindOneByIdReq = ProjectClusterVolumes_FindOneById_Req["data"];
type FindOneByIdRes = ProjectClusterVolumes_FindOneById_Res;
type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useProjectClusterVolumesApi();

    return useQuery({
        queryKey: [QK["projects.cluster-volumes.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectClusterVolumesQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

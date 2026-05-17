import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectCloudStorageApi } from "~/projects/api/hooks";
import type {
    ProjectCloudStorage_FindManyPaginated_Req,
    ProjectCloudStorage_FindManyPaginated_Res,
    ProjectCloudStorage_FindOneById_Req,
    ProjectCloudStorage_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectCloudStorage_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectCloudStorage_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectCloudStorageApi();

    return useQuery({
        queryKey: [QK["projects.cloud-storage.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = ProjectCloudStorage_FindOneById_Req["data"];
type FindOneByIdRes = ProjectCloudStorage_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectCloudStorageApi();

    return useQuery({
        queryKey: [QK["projects.cloud-storage.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectCloudStorageQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

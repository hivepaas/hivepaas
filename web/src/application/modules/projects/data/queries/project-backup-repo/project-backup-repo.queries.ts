import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectBackupRepoApi } from "~/projects/api/hooks";
import type {
    ProjectBackupRepo_FindManyPaginated_Req,
    ProjectBackupRepo_FindManyPaginated_Res,
    ProjectBackupRepo_FindOneById_Req,
    ProjectBackupRepo_FindOneById_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectBackupRepo_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectBackupRepo_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectBackupRepoApi();

    return useQuery({
        queryKey: [QK["projects.backup-repos.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindOneByIdReq = ProjectBackupRepo_FindOneById_Req["data"];
type FindOneByIdRes = ProjectBackupRepo_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectBackupRepoApi();

    return useQuery({
        queryKey: [QK["projects.backup-repos.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectBackupRepoQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

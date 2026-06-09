import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectGitCredentialsApi } from "~/projects/api/hooks";
import type {
    ProjectGitCredentials_FindManyPaginated_Req,
    ProjectGitCredentials_FindManyPaginated_Res,
    ProjectGitCredentials_FindManyRepos_Req,
    ProjectGitCredentials_FindManyRepos_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectGitCredentials_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectGitCredentials_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectGitCredentialsApi();

    return useQuery({
        queryKey: [QK["projects.git-credentials.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindManyReposReq = ProjectGitCredentials_FindManyRepos_Req["data"];
type FindManyReposRes = ProjectGitCredentials_FindManyRepos_Res;

function useFindManyRepos(
    request: FindManyReposReq,
    options: Omit<UseQueryOptions<FindManyReposRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectGitCredentialsApi();

    return useQuery({
        queryKey: [QK["projects.git-credentials.$.find-many-repos"], request],
        queryFn: ({ signal }) => queries.findManyRepos(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

export const ProjectGitCredentialsQueries = Object.freeze({
    useFindManyPaginated,
    useFindManyRepos,
});

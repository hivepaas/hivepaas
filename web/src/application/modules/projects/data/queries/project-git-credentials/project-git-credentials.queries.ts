import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectGitCredentialsApi } from "~/projects/api/hooks";
import type {
    ProjectGitCredentials_FindManyBranches_Req,
    ProjectGitCredentials_FindManyBranches_Res,
    ProjectGitCredentials_FindManyPaginated_Req,
    ProjectGitCredentials_FindManyPaginated_Res,
    ProjectGitCredentials_FindManyPullRequests_Req,
    ProjectGitCredentials_FindManyPullRequests_Res,
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

type FindManyBranchesReq = ProjectGitCredentials_FindManyBranches_Req["data"];
type FindManyBranchesRes = ProjectGitCredentials_FindManyBranches_Res;

function useFindManyBranches(
    request: FindManyBranchesReq,
    options: Omit<UseQueryOptions<FindManyBranchesRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectGitCredentialsApi();

    return useQuery({
        queryKey: [QK["projects.git-credentials.$.find-many-branches"], request],
        queryFn: ({ signal }) => queries.findManyBranches(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindManyPullRequestsReq = ProjectGitCredentials_FindManyPullRequests_Req["data"];
type FindManyPullRequestsRes = ProjectGitCredentials_FindManyPullRequests_Res;

function useFindManyPullRequests(
    request: FindManyPullRequestsReq,
    options: Omit<UseQueryOptions<FindManyPullRequestsRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectGitCredentialsApi();

    return useQuery({
        queryKey: [QK["projects.git-credentials.$.find-many-pull-requests"], request],
        queryFn: ({ signal }) => queries.findManyPullRequests(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

export const ProjectGitCredentialsQueries = Object.freeze({
    useFindManyPaginated,
    useFindManyRepos,
    useFindManyBranches,
    useFindManyPullRequests,
});

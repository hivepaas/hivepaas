import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectGithubAppApi } from "~/projects/api/hooks";
import type {
    ProjectGithubApp_FindManyPaginated_Req,
    ProjectGithubApp_FindManyPaginated_Res,
    ProjectGithubApp_FindOneById_Req,
    ProjectGithubApp_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectGithubApp_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectGithubApp_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectGithubAppApi();

    return useQuery({
        queryKey: [QK["projects.github-app.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = ProjectGithubApp_FindOneById_Req["data"];
type FindOneByIdRes = ProjectGithubApp_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectGithubAppApi();

    return useQuery({
        queryKey: [QK["projects.github-app.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectGithubAppQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

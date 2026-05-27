import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useGithubAppApi } from "~/settings/api/hooks";
import type {
    GithubApp_FindManyPaginated_Req,
    GithubApp_FindManyPaginated_Res,
    GithubApp_FindOneById_Req,
    GithubApp_FindOneById_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = GithubApp_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = GithubApp_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useGithubAppApi();

    return useQuery({
        queryKey: [QK["settings.github-app.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = GithubApp_FindOneById_Req["data"];
type FindOneByIdRes = GithubApp_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useGithubAppApi();

    return useQuery({
        queryKey: [QK["settings.github-app.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const GithubAppQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useGitCredentialsApi } from "~/settings/api/hooks";
import type {
    GitCredentials_FindManyPaginated_Req,
    GitCredentials_FindManyPaginated_Res,
} from "~/settings/api/services/git-credentials-services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = GitCredentials_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = GitCredentials_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useGitCredentialsApi();

    return useQuery({
        queryKey: [QK["settings.git-credentials.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const GitCredentialsQueries = Object.freeze({
    useFindManyPaginated,
});

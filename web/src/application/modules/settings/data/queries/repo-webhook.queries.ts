import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useRepoWebhookApi } from "~/settings/api/hooks";
import type {
    RepoWebhook_FindManyPaginated_Req,
    RepoWebhook_FindManyPaginated_Res,
    RepoWebhook_FindOneById_Req,
    RepoWebhook_FindOneById_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = RepoWebhook_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = RepoWebhook_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useRepoWebhookApi();

    return useQuery({
        queryKey: [QK["settings.repo-webhook.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = RepoWebhook_FindOneById_Req["data"];
type FindOneByIdRes = RepoWebhook_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useRepoWebhookApi();

    return useQuery({
        queryKey: [QK["settings.repo-webhook.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const RepoWebhookQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

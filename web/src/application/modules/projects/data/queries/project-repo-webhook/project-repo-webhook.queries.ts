import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectRepoWebhookApi } from "~/projects/api/hooks";
import type {
    ProjectRepoWebhook_FindManyPaginated_Req,
    ProjectRepoWebhook_FindManyPaginated_Res,
    ProjectRepoWebhook_FindOneById_Req,
    ProjectRepoWebhook_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectRepoWebhook_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectRepoWebhook_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectRepoWebhookApi();

    return useQuery({
        queryKey: [QK["projects.repo-webhook.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = ProjectRepoWebhook_FindOneById_Req["data"];
type FindOneByIdRes = ProjectRepoWebhook_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectRepoWebhookApi();

    return useQuery({
        queryKey: [QK["projects.repo-webhook.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectRepoWebhookQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

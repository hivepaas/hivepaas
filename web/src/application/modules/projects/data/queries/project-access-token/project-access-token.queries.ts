import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectAccessTokenApi } from "~/projects/api/hooks";
import type {
    ProjectAccessToken_FindManyPaginated_Req,
    ProjectAccessToken_FindManyPaginated_Res,
    ProjectAccessToken_FindOneById_Req,
    ProjectAccessToken_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectAccessToken_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectAccessToken_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectAccessTokenApi();

    return useQuery({
        queryKey: [QK["projects.access-token.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = ProjectAccessToken_FindOneById_Req["data"];
type FindOneByIdRes = ProjectAccessToken_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectAccessTokenApi();

    return useQuery({
        queryKey: [QK["projects.access-token.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectAccessTokenQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

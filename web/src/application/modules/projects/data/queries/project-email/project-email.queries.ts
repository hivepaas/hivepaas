import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectEmailApi } from "~/projects/api/hooks";
import type {
    ProjectEmail_FindManyPaginated_Req,
    ProjectEmail_FindManyPaginated_Res,
    ProjectEmail_FindOneById_Req,
    ProjectEmail_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectEmail_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectEmail_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectEmailApi();

    return useQuery({
        queryKey: [QK["projects.email.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = ProjectEmail_FindOneById_Req["data"];
type FindOneByIdRes = ProjectEmail_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectEmailApi();

    return useQuery({
        queryKey: [QK["projects.email.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectEmailQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

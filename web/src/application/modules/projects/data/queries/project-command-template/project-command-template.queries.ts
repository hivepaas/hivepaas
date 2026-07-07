import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectCommandTemplateApi } from "~/projects/api/hooks";
import type {
    ProjectCommandTemplate_FindManyPaginated_Req,
    ProjectCommandTemplate_FindManyPaginated_Res,
    ProjectCommandTemplate_FindOneById_Req,
    ProjectCommandTemplate_FindOneById_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectCommandTemplate_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectCommandTemplate_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectCommandTemplateApi();

    return useQuery({
        queryKey: [QK["projects.command-templates.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindOneByIdReq = ProjectCommandTemplate_FindOneById_Req["data"];
type FindOneByIdRes = ProjectCommandTemplate_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectCommandTemplateApi();

    return useQuery({
        queryKey: [QK["projects.command-templates.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectCommandTemplateQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

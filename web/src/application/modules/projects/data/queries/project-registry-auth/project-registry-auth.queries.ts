import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectRegistryAuthApi } from "~/projects/api/hooks";
import type {
    ProjectRegistryAuth_FindManyPaginated_Req,
    ProjectRegistryAuth_FindManyPaginated_Res,
    ProjectRegistryAuth_FindOneById_Req,
    ProjectRegistryAuth_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectRegistryAuth_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectRegistryAuth_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectRegistryAuthApi();

    return useQuery({
        queryKey: [QK["projects.registry-auth.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = ProjectRegistryAuth_FindOneById_Req["data"];
type FindOneByIdRes = ProjectRegistryAuth_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectRegistryAuthApi();

    return useQuery({
        queryKey: [QK["projects.registry-auth.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectRegistryAuthQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

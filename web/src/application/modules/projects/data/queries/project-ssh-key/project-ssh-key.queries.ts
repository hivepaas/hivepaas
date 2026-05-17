import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectSSHKeyApi } from "~/projects/api/hooks";
import type {
    ProjectSSHKey_FindManyPaginated_Req,
    ProjectSSHKey_FindManyPaginated_Res,
    ProjectSSHKey_FindOneById_Req,
    ProjectSSHKey_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectSSHKey_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectSSHKey_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectSSHKeyApi();

    return useQuery({
        queryKey: [QK["projects.ssh-key.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = ProjectSSHKey_FindOneById_Req["data"];
type FindOneByIdRes = ProjectSSHKey_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectSSHKeyApi();

    return useQuery({
        queryKey: [QK["projects.ssh-key.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectSSHKeyQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

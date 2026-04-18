import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectSslCertApi } from "~/projects/api/hooks";
import type {
    ProjectSslCert_FindManyPaginated_Req,
    ProjectSslCert_FindManyPaginated_Res,
    ProjectSslCert_FindOneById_Req,
    ProjectSslCert_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectSslCert_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectSslCert_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectSslCertApi();

    return useQuery({
        queryKey: [QK["projects.ssl-cert.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = ProjectSslCert_FindOneById_Req["data"];
type FindOneByIdRes = ProjectSslCert_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectSslCertApi();

    return useQuery({
        queryKey: [QK["projects.ssl-cert.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectSslCertQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

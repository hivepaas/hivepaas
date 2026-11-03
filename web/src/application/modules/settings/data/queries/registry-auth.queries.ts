import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useRegistryAuthApi } from "~/settings/api/hooks";
import type {
    RegistryAuth_FindManyPaginated_Req,
    RegistryAuth_FindManyPaginated_Res,
    RegistryAuth_FindOneById_Req,
    RegistryAuth_FindOneById_Res,
} from "~/settings/api/services/registry-auth-services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = RegistryAuth_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = RegistryAuth_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useRegistryAuthApi();

    return useQuery({
        queryKey: [QK["settings.registry-auth.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = RegistryAuth_FindOneById_Req["data"];
type FindOneByIdRes = RegistryAuth_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useRegistryAuthApi();

    return useQuery({
        queryKey: [QK["settings.registry-auth.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const RegistryAuthQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

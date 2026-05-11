import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAppConfigFilesApi } from "~/projects/api";
import type {
    AppConfigFiles_FindManyPaginated_Req,
    AppConfigFiles_FindManyPaginated_Res,
    AppConfigFiles_FindOneById_Req,
    AppConfigFiles_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

/**
 * Find many app config files paginated query
 */
type FindManyPaginatedReq = AppConfigFiles_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = AppConfigFiles_FindManyPaginated_Res;

type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useAppConfigFilesApi();

    return useQuery({
        queryKey: [QK["projects.apps.config-files.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

/**
 * Find one app config file by id query
 */
type FindOneByIdReq = AppConfigFiles_FindOneById_Req["data"];
type FindOneByIdRes = AppConfigFiles_FindOneById_Res;

type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useAppConfigFilesApi();

    return useQuery({
        queryKey: [QK["projects.apps.config-files.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const AppConfigFilesQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

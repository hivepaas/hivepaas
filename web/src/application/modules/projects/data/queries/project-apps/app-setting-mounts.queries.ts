import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAppSettingMountsApi } from "~/projects/api";
import type {
    AppSettingMounts_FindManyPaginated_Req,
    AppSettingMounts_FindManyPaginated_Res,
    AppSettingMounts_FindOneById_Req,
    AppSettingMounts_FindOneById_Res,
    AppSettingMounts_FindSources_Req,
    AppSettingMounts_FindSources_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

type FindManyPaginatedReq = AppSettingMounts_FindManyPaginated_Req["data"];
type FindManyPaginatedOptions = Omit<UseQueryOptions<AppSettingMounts_FindManyPaginated_Res>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useAppSettingMountsApi();

    return useQuery({
        queryKey: [QK["projects.apps.setting-mounts.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindOneByIdReq = AppSettingMounts_FindOneById_Req["data"];
type FindOneByIdOptions = Omit<UseQueryOptions<AppSettingMounts_FindOneById_Res>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useAppSettingMountsApi();

    return useQuery({
        queryKey: [QK["projects.apps.setting-mounts.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

type FindSourcesReq = AppSettingMounts_FindSources_Req["data"];
type FindSourcesOptions = Omit<UseQueryOptions<AppSettingMounts_FindSources_Res>, "queryKey" | "queryFn">;

function useFindSources(request: FindSourcesReq, options: FindSourcesOptions = {}) {
    const { queries } = useAppSettingMountsApi();

    return useQuery({
        queryKey: [QK["projects.apps.setting-mounts.$.find-sources"], request],
        queryFn: ({ signal }) => queries.findSources(request, signal),
        ...options,
    });
}

export const AppSettingMountsQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
    useFindSources,
});

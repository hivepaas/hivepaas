import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAppDataFilesApi } from "~/projects/api";
import type {
    AppDataFiles_FindManyPaginated_Req,
    AppDataFiles_FindManyPaginated_Res,
    AppDataFiles_FindOneById_Req,
    AppDataFiles_FindOneById_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

type FindManyPaginatedReq = AppDataFiles_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = AppDataFiles_FindManyPaginated_Res;
type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useAppDataFilesApi();

    return useQuery({
        queryKey: [QK["projects.apps.data-files.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindOneByIdReq = AppDataFiles_FindOneById_Req["data"];
type FindOneByIdOptions = Omit<UseQueryOptions<AppDataFiles_FindOneById_Res>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useAppDataFilesApi();

    return useQuery({
        queryKey: [QK["projects.apps.data-files.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const AppDataFilesQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAppDataFilesApi } from "~/projects/api";
import type { AppDataFiles_FindManyPaginated_Req, AppDataFiles_FindManyPaginated_Res } from "~/projects/api/services";
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

export const AppDataFilesQueries = Object.freeze({
    useFindManyPaginated,
});

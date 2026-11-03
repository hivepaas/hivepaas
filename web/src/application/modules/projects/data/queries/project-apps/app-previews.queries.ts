import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAppPreviewsApi } from "~/projects/api";
import type { AppPreviews_FindManyPaginated_Req, AppPreviews_FindManyPaginated_Res } from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

type FindManyPaginatedReq = AppPreviews_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = AppPreviews_FindManyPaginated_Res;
type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useAppPreviewsApi();

    return useQuery({
        queryKey: [QK["projects.apps.previews.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

export const AppPreviewsQueries = Object.freeze({
    useFindManyPaginated,
});

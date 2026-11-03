import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAppServiceTasksApi } from "~/projects/api";
import type { AppServiceTasks_FindMany_Req, AppServiceTasks_FindMany_Res } from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyReq = AppServiceTasks_FindMany_Req["data"];
type FindManyRes = AppServiceTasks_FindMany_Res;
type FindManyOptions = Omit<UseQueryOptions<FindManyRes>, "queryKey" | "queryFn">;

function useFindMany(request: FindManyReq, options: FindManyOptions = {}) {
    const { queries } = useAppServiceTasksApi();

    return useQuery({
        queryKey: [QK["projects.apps.service-tasks.$.find-many"], request],
        queryFn: ({ signal }) => queries.findMany(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppServiceTasksQueries = Object.freeze({
    useFindMany,
});

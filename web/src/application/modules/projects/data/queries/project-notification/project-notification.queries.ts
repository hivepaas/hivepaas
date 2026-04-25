import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectNotificationsApi } from "~/projects/api/hooks";
import type {
    ProjectNotification_FindManyPaginated_Req,
    ProjectNotification_FindManyPaginated_Res,
    ProjectNotification_FindOneById_Req,
    ProjectNotification_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectNotification_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectNotification_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectNotificationsApi();

    return useQuery({
        queryKey: [QK["projects.notifications.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = ProjectNotification_FindOneById_Req["data"];
type FindOneByIdRes = ProjectNotification_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectNotificationsApi();

    return useQuery({
        queryKey: [QK["projects.notifications.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectNotificationQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

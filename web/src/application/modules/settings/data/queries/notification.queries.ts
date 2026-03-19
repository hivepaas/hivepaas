import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useNotificationsApi } from "~/settings/api/hooks";
import {
    type Notifications_FindManyPaginated_Req,
    type Notifications_FindManyPaginated_Res,
} from "~/settings/api/services/notifications-services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = Notifications_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = Notifications_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useNotificationsApi();

    return useQuery({
        queryKey: [QK["settings.notifications.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const NotificationQueries = Object.freeze({
    useFindManyPaginated,
});

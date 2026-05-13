import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useImServiceApi } from "~/settings/api/hooks";
import type {
    ImService_FindManyPaginated_Req,
    ImService_FindManyPaginated_Res,
    ImService_FindOneById_Req,
    ImService_FindOneById_Res,
} from "~/settings/api/services/im-service-services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = ImService_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ImService_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useImServiceApi();

    return useQuery({
        queryKey: [QK["settings.im-service.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = ImService_FindOneById_Req["data"];
type FindOneByIdRes = ImService_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useImServiceApi();

    return useQuery({
        queryKey: [QK["settings.im-service.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ImServiceQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

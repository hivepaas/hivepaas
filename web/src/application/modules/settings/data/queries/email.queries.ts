import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useEmailApi } from "~/settings/api/hooks";
import type {
    Email_FindManyPaginated_Req,
    Email_FindManyPaginated_Res,
    Email_FindOneById_Req,
    Email_FindOneById_Res,
} from "~/settings/api/services/email-services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = Email_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = Email_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useEmailApi();

    return useQuery({
        queryKey: [QK["settings.email.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = Email_FindOneById_Req["data"];
type FindOneByIdRes = Email_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useEmailApi();

    return useQuery({
        queryKey: [QK["settings.email.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const EmailQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useCloudStorageApi } from "~/settings/api/hooks";
import type {
    CloudStorage_FindManyPaginated_Req,
    CloudStorage_FindManyPaginated_Res,
    CloudStorage_FindOneById_Req,
    CloudStorage_FindOneById_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = CloudStorage_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = CloudStorage_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useCloudStorageApi();

    return useQuery({
        queryKey: [QK["settings.cloud-storage.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = CloudStorage_FindOneById_Req["data"];
type FindOneByIdRes = CloudStorage_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useCloudStorageApi();

    return useQuery({
        queryKey: [QK["settings.cloud-storage.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const CloudStorageQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});

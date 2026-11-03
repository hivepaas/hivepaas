import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useHomeAttentionApi } from "~/home/api";
import type { HomeAttention_FindAll_Res } from "~/home/api/services";
import { QK } from "~/home/data/constants";

type FindAllOptions = Omit<UseQueryOptions<HomeAttention_FindAll_Res>, "queryKey" | "queryFn">;

/**
 * The server reads the cluster at most every 30 seconds, so asking more often
 * than that only asks again for the same answer.
 */
const REFRESH_MS = 30_000;

function useFindAll(options: FindAllOptions = {}) {
    const { queries } = useHomeAttentionApi();

    return useQuery({
        queryKey: [QK["home.attention.find-all"]],
        queryFn: ({ signal }) => queries.findAll(signal),
        refetchInterval: REFRESH_MS,
        staleTime: REFRESH_MS,
        ...options,
    });
}

export const HomeAttentionQueries = Object.freeze({
    useFindAll,
});

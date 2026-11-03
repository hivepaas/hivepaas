import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAuthApi } from "@application/authentication/api";
import { type Auth_GetLoginOptions_Res } from "@application/authentication/api/services";
import { QK } from "@application/authentication/data/constants";

/**
 * Get login options
 */
type GetLoginOptionsRes = Auth_GetLoginOptions_Res;

type GetLoginOptionsOptions = Omit<UseQueryOptions<GetLoginOptionsRes>, "queryKey" | "queryFn">;

function useGetLoginOptions(options: GetLoginOptionsOptions = {}) {
    const { queries } = useAuthApi();

    return useQuery({
        queryKey: [QK["auth.get-login-options"]],
        queryFn: ({ signal }) => queries.getLoginOptions(signal),
        placeholderData: keepPreviousData,

        ...options,
    });
}

export const AuthQueries = Object.freeze({
    useGetLoginOptions,
});

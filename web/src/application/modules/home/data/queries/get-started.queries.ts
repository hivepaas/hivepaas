import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useGetStartedApi } from "~/home/api";
import type { GetStarted_GetDashboardCert_Res } from "~/home/api/services";
import { QK } from "~/home/data/constants";

type DashboardCertOptions = Omit<UseQueryOptions<GetStarted_GetDashboardCert_Res>, "queryKey" | "queryFn">;

/** How often the certificate is read again while it is being obtained. */
const OBTAINING_REFRESH_MS = 5_000;

/**
 * Where the dashboard's certificate stands. It is read again only while it is
 * being obtained: the answer changes then, and at no other time on its own.
 */
function useDashboardCert(options: DashboardCertOptions = {}) {
    const { queries } = useGetStartedApi();

    return useQuery({
        queryKey: [QK["home.get-started.dashboard-cert"]],
        queryFn: ({ signal }) => queries.getDashboardCert(signal),
        refetchInterval: query => (query.state.data?.data.status === "obtaining" ? OBTAINING_REFRESH_MS : false),
        refetchOnWindowFocus: false,
        ...options,
    });
}

export const GetStartedQueries = Object.freeze({
    useDashboardCert,
});

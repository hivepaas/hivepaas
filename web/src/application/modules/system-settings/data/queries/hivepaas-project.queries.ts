import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useHivePaaSProjectApi } from "~/system-settings/api/hooks";
import type { HivePaaSProject_FindOne_Res } from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneRes = HivePaaSProject_FindOne_Res;

function useFindOne(options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useHivePaaSProjectApi();

    return useQuery({
        queryKey: [QK["system-settings.hivepaas.project.find-one"]],
        queryFn: ({ signal }) => queries.findOne({}, signal),
        // Made once, when HivePaaS is installed: it does not change under the page.
        staleTime: Infinity,
        ...options,
    });
}

export const HivePaaSProjectQueries = Object.freeze({
    useFindOne,
});

import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useMcpSettingsApi } from "~/system-settings/api/hooks";
import type { McpSettings_FindOne_Req, McpSettings_FindOne_Res } from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = McpSettings_FindOne_Req["data"];
type FindOneRes = McpSettings_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useMcpSettingsApi();

    return useQuery({
        queryKey: [QK["system-settings.mcp.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const McpSettingsQueries = Object.freeze({
    useFindOne,
});

import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppHttpSettingsApi } from "../../../api/hooks/project-apps";
import { type AppHttpSettings_FindOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

function useFindOne(
    request: { projectID: string; appID: string },
    options: Omit<UseQueryOptions<AppHttpSettings_FindOne_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppHttpSettingsApi();

    return useQuery({
        queryKey: [QK["projects.apps.http-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppHttpSettingsQueries = Object.freeze({
    useFindOne,
});

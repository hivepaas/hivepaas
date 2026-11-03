import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppKindSettingsApi } from "../../../api/hooks/project-apps";
import { type AppKindSettings_FindOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

function useFindOne(
    request: { projectID: string; env: string; appID: string; revealSecrets?: boolean },
    options: Omit<UseQueryOptions<AppKindSettings_FindOne_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppKindSettingsApi();

    return useQuery({
        queryKey: [QK["projects.apps.kind-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppKindSettingsQueries = Object.freeze({
    useFindOne,
});

import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useProjectStorageSettingsApi } from "../../../api/hooks/project-settings";
import { type ProjectStorageSettings_FindOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

function useFindOne(
    request: { projectID: string },
    options: Omit<UseQueryOptions<ProjectStorageSettings_FindOne_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectStorageSettingsApi();

    return useQuery({
        queryKey: [QK["projects.storage-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const ProjectStorageSettingsQueries = Object.freeze({
    useFindOne,
});

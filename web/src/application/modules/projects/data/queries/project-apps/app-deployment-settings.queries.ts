import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppDeploymentSettingsApi } from "../../../api/hooks/project-apps";
import {
    type AppDeploymentSettings_FindOne_Res,
    type AppDeploymentSettings_GetDockerfileTemplate_Res,
} from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

function useFindOne(
    request: { projectID: string; env: string; appID: string },
    options: Omit<UseQueryOptions<AppDeploymentSettings_FindOne_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppDeploymentSettingsApi();

    return useQuery({
        queryKey: [QK["projects.apps.deployment-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

function useGetDockerfileTemplate(
    request: { projectID: string; env: string; appID: string; type: string },
    options: Omit<UseQueryOptions<AppDeploymentSettings_GetDockerfileTemplate_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppDeploymentSettingsApi();

    return useQuery({
        queryKey: [QK["projects.apps.deployment-settings.$.dockerfile-template"], request],
        queryFn: ({ signal }) => queries.getDockerfileTemplate(request, signal),
        enabled: false,
        ...options,
    });
}

export const AppDeploymentSettingsQueries = Object.freeze({
    useFindOne,
    useGetDockerfileTemplate,
});

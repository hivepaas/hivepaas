import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppManagedApi } from "../../../api/hooks/project-apps";
import type {
    AppManaged_FindSettings_Res,
    AppManaged_FindStatus_Res,
    AppManaged_FindSwitch_Res,
} from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

type AppScope = { projectID: string; env: string; appID: string };

function useFindSettings(
    request: AppScope,
    options: Omit<UseQueryOptions<AppManaged_FindSettings_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppManagedApi();

    return useQuery({
        queryKey: [QK["projects.apps.managed.$.find-settings"], request],
        queryFn: ({ signal }) => queries.findSettings(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

function useFindStatus(
    request: AppScope,
    options: Omit<UseQueryOptions<AppManaged_FindStatus_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppManagedApi();

    return useQuery({
        queryKey: [QK["projects.apps.managed.$.find-status"], request],
        queryFn: ({ signal }) => queries.findStatus(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

function useFindSwitch(
    request: AppScope,
    options: Omit<UseQueryOptions<AppManaged_FindSwitch_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppManagedApi();

    return useQuery({
        queryKey: [QK["projects.apps.managed.$.find-switch"], request],
        queryFn: ({ signal }) => queries.findSwitch(request, signal),
        ...options,
    });
}

export const AppManagedQueries = Object.freeze({
    useFindSettings,
    useFindStatus,
    useFindSwitch,
});

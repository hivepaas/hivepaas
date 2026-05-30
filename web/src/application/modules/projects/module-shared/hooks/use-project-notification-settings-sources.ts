import { useMemo, useState } from "react";

import { ProjectNotificationQueries } from "~/projects/data/queries";

import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import type {
    NotificationSettingsManageLink,
    NotificationSettingsRef,
    NotificationSettingsSource,
    NotificationSettingsSources,
} from "@application/shared/form";

function toOptions(items: NotificationSettingsRef[]): NotificationSettingsSource["options"] {
    return items.map(item => ({
        value: { id: item.id, name: item.name },
        label: item.name,
    }));
}

export function useProjectNotificationSettingsSources(projectId: string): {
    sources: NotificationSettingsSources;
    manageLink: NotificationSettingsManageLink;
} {
    const [successSearch, setSuccessSearch] = useState("");
    const [failureSearch, setFailureSearch] = useState("");

    const successQuery = ProjectNotificationQueries.useFindManyPaginated(
        { projectID: projectId, search: successSearch },
        { enabled: projectId.length > 0 },
    );
    const failureQuery = ProjectNotificationQueries.useFindManyPaginated(
        { projectID: projectId, search: failureSearch },
        { enabled: projectId.length > 0 },
    );

    const successOptions = useMemo(() => {
        return toOptions((successQuery.data ?? DEFAULT_PAGINATED_DATA).data);
    }, [successQuery.data]);

    const failureOptions = useMemo(() => {
        return toOptions((failureQuery.data ?? DEFAULT_PAGINATED_DATA).data);
    }, [failureQuery.data]);

    return {
        sources: {
            success: {
                options: successOptions,
                isFetching: successQuery.isFetching,
                isRefreshing: successQuery.isRefetching,
                onSearch: setSuccessSearch,
                onRefresh: () => {
                    void successQuery.refetch();
                },
            },
            failure: {
                options: failureOptions,
                isFetching: failureQuery.isFetching,
                isRefreshing: failureQuery.isRefetching,
                onSearch: setFailureSearch,
                onRefresh: () => {
                    void failureQuery.refetch();
                },
            },
        },
        manageLink: {
            to: ROUTE.projects.single.providerConfiguration.notificationTargets.$route(projectId),
            label: "Configure notification settings",
        },
    };
}

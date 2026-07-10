import { useMemo, useState } from "react";

import { ProjectNotificationQueries } from "~/projects/data/queries";
import { NotificationQueries } from "~/settings/data";

import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import type {
    NotificationSettingsManageLink,
    NotificationSettingsRef,
    NotificationSettingsSource,
    NotificationSettingsSources,
} from "@application/shared/form";

export type NotificationSettingsScope =
    | {
          type: "settings";
      }
    | {
          type: "project";
          projectId: string;
      };

function toOptions(items: NotificationSettingsRef[]): NotificationSettingsSource["options"] {
    return items.map(item => ({
        value: { id: item.id, name: item.name },
        label: item.name,
    }));
}

export function useNotificationSettingsSources(scope: NotificationSettingsScope): {
    sources: NotificationSettingsSources;
    manageLink: NotificationSettingsManageLink;
} {
    const [successSearch, setSuccessSearch] = useState("");
    const [failureSearch, setFailureSearch] = useState("");
    const projectId = scope.type === "project" ? scope.projectId : "";

    const settingsSuccessQuery = NotificationQueries.useFindManyPaginated(
        { search: successSearch },
        { enabled: scope.type === "settings" },
    );
    const settingsFailureQuery = NotificationQueries.useFindManyPaginated(
        { search: failureSearch },
        { enabled: scope.type === "settings" },
    );

    const projectSuccessQuery = ProjectNotificationQueries.useFindManyPaginated(
        { projectID: projectId, search: successSearch },
        { enabled: scope.type === "project" && projectId.length > 0 },
    );
    const projectFailureQuery = ProjectNotificationQueries.useFindManyPaginated(
        { projectID: projectId, search: failureSearch },
        { enabled: scope.type === "project" && projectId.length > 0 },
    );

    const successQuery = scope.type === "project" ? projectSuccessQuery : settingsSuccessQuery;
    const failureQuery = scope.type === "project" ? projectFailureQuery : settingsFailureQuery;

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
        manageLink:
            scope.type === "project"
                ? {
                      to: ROUTE.projects.single.providerConfiguration.notificationTargets.$route(scope.projectId),
                      label: "Configure notification settings",
                  }
                : {
                      to: ROUTE.settings.notificationTargets.$route,
                      label: "Configure Notification Settings",
                  },
    };
}

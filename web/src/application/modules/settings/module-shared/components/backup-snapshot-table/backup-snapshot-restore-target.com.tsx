import { useMemo, useState } from "react";

import { type SearchableFilterItem, SearchableFilterSelect } from "~/operations/routes/tasks";
import type { BackupSnapshotScope } from "~/settings/domain";

import { AppsPublicQueries, ProjectsPublicQueries } from "@application/shared/data-public/queries";

/** The app a restore puts the data into. */
export interface RestoreTarget {
    id: string;
    name: string;
    projectId: string;
    env: string;
}

const LIST_ALL_PAGE = { page: 1, size: 1000 };

/**
 * Picks the app a restore goes into, among the apps the view reaches: at the global view a project, then one of
 * its apps; at a project's or an env's, one of its apps; at an app's, the app itself.
 */
export function BackupSnapshotRestoreTarget({ scope, value, onChange, disabled }: Props) {
    const [pickedProjectId, setPickedProjectId] = useState(value?.projectId ?? "");
    const projectId = scope.type === "settings" ? pickedProjectId : scope.projectId;
    const env = scope.type === "project" ? scope.env : undefined;

    const { data: projectsData } = ProjectsPublicQueries.useFindManyPaginated(
        { pagination: LIST_ALL_PAGE },
        { enabled: scope.type === "settings" },
    );
    const { data: appsData } = AppsPublicQueries.useFindMany(
        { projectID: projectId, env, pagination: LIST_ALL_PAGE },
        { enabled: scope.type !== "app" && Boolean(projectId) },
    );

    const projectItems: SearchableFilterItem[] = useMemo(
        () =>
            (projectsData?.data ?? []).map(project => ({
                value: project.id,
                label: project.name,
                searchKey: project.name,
                avatar: { name: project.name },
            })),
        [projectsData?.data],
    );
    const apps = useMemo(() => appsData?.data ?? [], [appsData?.data]);
    const appItems: SearchableFilterItem[] = useMemo(
        () =>
            apps.map(app => ({
                value: app.id,
                label: app.name,
                searchKey: `${app.name} ${app.env ?? ""}`,
                badge: env ? undefined : app.env,
                avatar: { name: app.name },
            })),
        [apps, env],
    );

    if (scope.type === "app") {
        return <span className="text-sm font-medium">{value?.name ?? scope.appId}</span>;
    }

    return (
        <div className="flex flex-col gap-2 sm:flex-row">
            {scope.type === "settings" && (
                <div className="sm:w-1/2">
                    <SearchableFilterSelect
                        value={pickedProjectId}
                        onValueChange={id => {
                            setPickedProjectId(id);
                            onChange(null);
                        }}
                        placeholder="Select a project"
                        searchPlaceholder="Search projects..."
                        emptyText="No projects found."
                        items={projectItems}
                    />
                </div>
            )}
            <div className={scope.type === "settings" ? "sm:w-1/2" : "w-full"}>
                <SearchableFilterSelect
                    value={value?.id ?? ""}
                    onValueChange={id => {
                        const app = apps.find(item => item.id === id);
                        if (app && !disabled) {
                            onChange({ id: app.id, name: app.name, projectId, env: app.env ?? env ?? "" });
                        }
                    }}
                    placeholder={value?.name ?? "Select an app"}
                    searchPlaceholder="Search apps..."
                    emptyText={projectId ? "No apps found." : "Select a project first."}
                    items={appItems}
                />
            </div>
        </div>
    );
}

interface Props {
    scope: BackupSnapshotScope;
    value: RestoreTarget | null;
    onChange: (target: RestoreTarget | null) => void;
    disabled?: boolean;
}

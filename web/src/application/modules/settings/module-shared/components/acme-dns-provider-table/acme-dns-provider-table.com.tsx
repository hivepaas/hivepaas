import { useMemo } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Plus } from "lucide-react";
import { PROJECT_SETTINGS_IMPORT_KIND } from "~/projects/data/commands";
import { ProjectAcmeDnsProviderQueries } from "~/projects/data/queries";
import { AcmeDnsProviderQueries } from "~/settings/data/queries";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { ProjectSettingsImportButton } from "../project-settings-import-button";
import { SettingsScopeCreateButton } from "../settings-scope-create-button";

import { AcmeDnsProviderTableDefs } from "./acme-dns-provider-table.defs";
import type { AcmeDnsProviderTableScope } from "./acme-dns-provider-table.types";

const ACME_DNS_PROVIDER_NOTE =
    "An ACME DNS-01 challenge provider allows you to request SSL certificates for wildcard domains from certificate providers. This is extremely useful because you can deploy applications without having to request a new SSL certificate for each deployment.";

function AcmeDnsProviderTableView({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

    const settingsQuery = AcmeDnsProviderQueries.useFindManyPaginated(
        {
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "settings",
        },
    );

    const projectQuery = ProjectAcmeDnsProviderQueries.useFindManyPaginated(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            env: scope.type === "project" ? scope.env : undefined,
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "project",
        },
    );

    const query = scope.type === "project" ? projectQuery : settingsQuery;
    const { data: { data: acmeDnsProviderItems, meta } = DEFAULT_PAGINATED_DATA, isFetching } = query;
    const columns = useMemo(() => AcmeDnsProviderTableDefs.columns(scope), [scope]);

    return (
        <div className="flex flex-col gap-4">
            <div className={cn(dashedBorderBox, "text-sm leading-6")}>
                <span className="font-semibold text-orange-500">Note:</span> {ACME_DNS_PROVIDER_NOTE}
            </div>

            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        {scope.type === "project" && (
                            <ProjectSettingsImportButton
                                projectId={scope.projectId}
                                env={scope.env}
                                settingKind={PROJECT_SETTINGS_IMPORT_KIND.AcmeDnsProvider}
                            />
                        )}
                        <SettingsScopeCreateButton
                            scope={scope}
                            onClick={() => {
                                navigate.modules(getAcmeDnsProviderCreateRoute(scope));
                            }}
                        >
                            <Plus className="size-4" />
                            New DNS-01 Provider
                        </SettingsScopeCreateButton>
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={acmeDnsProviderItems}
                pageSize={pagination.size}
                manualPagination
                totalCount={meta.page.total}
                manualSorting
                enableSorting
                enablePagination
                isLoading={isFetching}
                onPaginationChange={setPagination}
                onSortingChange={setSorting}
            />
        </div>
    );
}

function getAcmeDnsProviderCreateRoute(scope: AcmeDnsProviderTableScope) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.acmeDnsProviders.create.$route(scope.projectId);
    }

    return ROUTE.settings.acmeDnsProviders.create.$route;
}

interface Props {
    scope: AcmeDnsProviderTableScope;
}

export function SettingsAcmeDnsProviderTable() {
    return <AcmeDnsProviderTableView scope={{ type: "settings" }} />;
}

export function ProjectAcmeDnsProviderTable({ projectId, env }: ProjectProps) {
    return <AcmeDnsProviderTableView scope={{ type: "project", projectId, env }} />;
}

interface ProjectProps {
    projectId: string;
    env?: string;
}

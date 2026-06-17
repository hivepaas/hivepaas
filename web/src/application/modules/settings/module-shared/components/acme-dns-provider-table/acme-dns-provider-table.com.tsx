import { useMemo } from "react";

import { Plus } from "lucide-react";
import { PROJECT_SETTINGS_IMPORT_KIND } from "~/projects/data/commands";
import { ProjectAcmeDnsProviderQueries } from "~/projects/data/queries";
import { AcmeDnsProviderQueries } from "~/settings/data/queries";
import { useCreateOrEditAcmeDnsProviderDialog } from "~/settings/dialogs/create-or-edit-acme-dns-provider";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
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
    const createOrEditDialog = useCreateOrEditAcmeDnsProviderDialog();

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
            <div className="rounded-md border border-dashed border-primary/70 bg-muted/30 px-4 py-3 text-center text-sm">
                <span className="text-orange-500">Note:</span> {ACME_DNS_PROVIDER_NOTE}
            </div>
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        {scope.type === "project" && (
                            <ProjectSettingsImportButton
                                projectId={scope.projectId}
                                settingKind={PROJECT_SETTINGS_IMPORT_KIND.AcmeDnsProvider}
                            />
                        )}
                        <SettingsScopeCreateButton
                            scope={scope}
                            onClick={() => {
                                createOrEditDialog.actions.open(scope);
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
                showPageSizeSelector={false}
            />
        </div>
    );
}

interface Props {
    scope: AcmeDnsProviderTableScope;
}

export function SettingsAcmeDnsProviderTable() {
    return <AcmeDnsProviderTableView scope={{ type: "settings" }} />;
}

export function ProjectAcmeDnsProviderTable({ projectId }: ProjectProps) {
    return <AcmeDnsProviderTableView scope={{ type: "project", projectId }} />;
}

interface ProjectProps {
    projectId: string;
}

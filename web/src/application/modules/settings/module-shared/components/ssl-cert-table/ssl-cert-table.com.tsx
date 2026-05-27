import { useMemo } from "react";

import { Plus } from "lucide-react";
import { PROJECT_SETTINGS_IMPORT_KIND } from "~/projects/data/commands";
import { ProjectSslCertQueries } from "~/projects/data/queries";
import { SslCertQueries } from "~/settings/data/queries";
import { useCreateOrEditSslCertDialog } from "~/settings/dialogs/create-or-edit-ssl-cert";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { Button, DataTable } from "@/components/ui";

import { ProjectSettingsImportButton } from "../project-settings-import-button";

import { SslCertTableDefs } from "./ssl-cert-table.defs";
import type { SslCertTableScope } from "./ssl-cert-table.types";

function SslCertTableView({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const createOrEditDialog = useCreateOrEditSslCertDialog();

    const settingsQuery = SslCertQueries.useFindManyPaginated(
        {
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "settings",
        },
    );

    const projectQuery = ProjectSslCertQueries.useFindManyPaginated(
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
    const { data: { data: sslCertItems, meta } = DEFAULT_PAGINATED_DATA, isFetching } = query;
    const columns = useMemo(() => SslCertTableDefs.columns(scope), [scope]);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        {scope.type === "project" && (
                            <ProjectSettingsImportButton
                                projectId={scope.projectId}
                                settingKind={PROJECT_SETTINGS_IMPORT_KIND.SslCert}
                            />
                        )}
                        <Button
                            onClick={() => {
                                createOrEditDialog.actions.open(scope);
                            }}
                        >
                            <Plus className="size-4" />
                            New SSL Certificate
                        </Button>
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={sslCertItems}
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
    scope: SslCertTableScope;
}

export function SettingsSslCertTable() {
    return <SslCertTableView scope={{ type: "settings" }} />;
}

export function ProjectSslCertTable({ projectId }: ProjectProps) {
    return <SslCertTableView scope={{ type: "project", projectId }} />;
}

interface ProjectProps {
    projectId: string;
}

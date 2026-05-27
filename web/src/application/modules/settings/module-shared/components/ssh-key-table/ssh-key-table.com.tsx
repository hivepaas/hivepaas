import { useMemo } from "react";

import { Plus } from "lucide-react";
import { PROJECT_SETTINGS_IMPORT_KIND } from "~/projects/data/commands";
import { ProjectSSHKeyQueries } from "~/projects/data/queries";
import { SSHKeyQueries } from "~/settings/data/queries";
import { useCreateOrEditSSHKeyDialog } from "~/settings/dialogs/create-or-edit-ssh-key";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { Button, DataTable } from "@/components/ui";

import { ProjectSettingsImportButton } from "../project-settings-import-button";

import { SSHKeyTableDefs } from "./ssh-key-table.defs";
import type { SSHKeyTableScope } from "./ssh-key-table.types";

function SSHKeyTableView({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const createOrEditDialog = useCreateOrEditSSHKeyDialog();

    const settingsQuery = SSHKeyQueries.useFindManyPaginated(
        { pagination, sorting, search },
        { enabled: scope.type === "settings" },
    );

    const projectQuery = ProjectSSHKeyQueries.useFindManyPaginated(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            pagination,
            sorting,
            search,
        },
        { enabled: scope.type === "project" },
    );

    const query = scope.type === "project" ? projectQuery : settingsQuery;
    const { data: { data: sshKeys, meta } = DEFAULT_PAGINATED_DATA, isFetching } = query;
    const columns = useMemo(() => SSHKeyTableDefs.columns(scope), [scope]);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <div className="flex flex-wrap gap-3">
                        {scope.type === "project" && (
                            <ProjectSettingsImportButton
                                projectId={scope.projectId}
                                settingKind={PROJECT_SETTINGS_IMPORT_KIND.SSHKey}
                            />
                        )}
                        <Button
                            onClick={() => {
                                createOrEditDialog.actions.open(scope);
                            }}
                        >
                            <Plus className="size-4" />
                            New SSH Key
                        </Button>
                    </div>
                }
            />
            <DataTable
                columns={columns}
                data={sshKeys}
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
    scope: SSHKeyTableScope;
}

export function SettingsSSHKeyTable() {
    return <SSHKeyTableView scope={{ type: "settings" }} />;
}

export function ProjectSSHKeyTable({ projectId }: ProjectProps) {
    return <SSHKeyTableView scope={{ type: "project", projectId }} />;
}

interface ProjectProps {
    projectId: string;
}

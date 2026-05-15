import { useMemo } from "react";

import { Plus } from "lucide-react";
import { ProjectEmailQueries } from "~/projects/data/queries";
import { EmailQueries } from "~/settings/data/queries";
import { useCreateOrEditEmailAccountDialog } from "~/settings/dialogs/create-or-edit-email-account";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { Button, DataTable } from "@/components/ui";

import { EmailAccountTableDefs } from "./email-account-table.defs";
import type { EmailAccountTableScope } from "./email-account-table.types";

function EmailAccountTableView({ scope }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const createOrEditDialog = useCreateOrEditEmailAccountDialog();

    const settingsQuery = EmailQueries.useFindManyPaginated(
        {
            pagination,
            sorting,
            search,
        },
        {
            enabled: scope.type === "settings",
        },
    );

    const projectQuery = ProjectEmailQueries.useFindManyPaginated(
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
    const { data: { data: emailAccountItems, meta } = DEFAULT_PAGINATED_DATA, isFetching } = query;
    const columns = useMemo(() => EmailAccountTableDefs.columns(scope), [scope]);

    return (
        <div className="flex flex-col gap-4">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <Button
                        onClick={() => {
                            createOrEditDialog.actions.open(scope);
                        }}
                    >
                        <Plus className="size-4" />
                        New Email Account
                    </Button>
                }
            />
            <DataTable
                columns={columns}
                data={emailAccountItems}
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
    scope: EmailAccountTableScope;
}

export function SettingsEmailAccountsTable() {
    return <EmailAccountTableView scope={{ type: "settings" }} />;
}

export function ProjectEmailAccountsTable({ projectId }: ProjectProps) {
    return <EmailAccountTableView scope={{ type: "project", projectId }} />;
}

interface ProjectProps {
    projectId: string;
}

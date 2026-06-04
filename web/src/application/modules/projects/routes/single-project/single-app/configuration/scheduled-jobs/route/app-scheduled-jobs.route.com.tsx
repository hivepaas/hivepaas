import { useMemo } from "react";

import { Plus } from "lucide-react";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { AppScheduledJobsQueries } from "~/projects/data/queries";
import { useCreateOrEditAppScheduledJobDialog } from "~/projects/dialogs/create-or-edit-app-scheduled-job";
import { AppScheduledJobsTableDefs } from "~/projects/module-shared/definitions/tables/app-scheduled-jobs";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { Button, DataTable } from "@/components/ui";

export function AppScheduledJobsRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { actions: scheduledJobDialogActions } = useCreateOrEditAppScheduledJobDialog();
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();

    const { data: { data: scheduledJobs, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        AppScheduledJobsQueries.useFindManyPaginated(
            {
                projectID: projectId,
                appID: appId,
                pagination,
                sorting,
                search,
            },
            APP_CONFIGURATION_QUERY_OPTIONS,
        );

    const columns = useMemo(() => AppScheduledJobsTableDefs.columns(projectId, appId), [projectId, appId]);

    return (
        <div className="flex flex-col gap-6">
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderActions={
                    <PermissionTooltipAction
                        id={MODULE_IDS.Project}
                        action="write"
                    >
                        {({ isDenied }) => (
                            <Button
                                onClick={() => {
                                    scheduledJobDialogActions.open(projectId, appId);
                                }}
                                disabled={isDenied}
                            >
                                <Plus className="size-4" /> New Scheduled Job
                            </Button>
                        )}
                    </PermissionTooltipAction>
                }
            />

            <DataTable
                columns={columns}
                data={scheduledJobs}
                isLoading={isFetching}
                pageSize={pagination.size}
                manualPagination
                totalCount={meta.page.total}
                onPaginationChange={setPagination}
                manualSorting
                onSortingChange={setSorting}
                enablePagination
                showPageSizeSelector={false}
            />
        </div>
    );
}

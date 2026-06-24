import { useMemo } from "react";

import { Plus } from "lucide-react";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { AppHealthChecksQueries } from "~/projects/data/queries";
import { AppHealthChecksTableDefs } from "~/projects/module-shared/definitions/tables/app-health-checks";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { Button, DataTable } from "@/components/ui";

export function AppHealthChecksRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { navigate } = useAppNavigate();
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();

    const { data: { data: healthChecks, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        AppHealthChecksQueries.useFindManyPaginated(
            {
                projectID: projectId,
                appID: appId,
                pagination,
                sorting,
                search,
            },
            APP_CONFIGURATION_QUERY_OPTIONS,
        );

    const columns = useMemo(() => AppHealthChecksTableDefs.columns(projectId, appId), [projectId, appId]);

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
                                    navigate.modules(
                                        ROUTE.projects.single.apps.single.configuration.healthChecks.create.$route(
                                            projectId,
                                            appId,
                                        ),
                                    );
                                }}
                                disabled={isDenied}
                            >
                                <Plus className="size-4" /> New Health Check
                            </Button>
                        )}
                    </PermissionTooltipAction>
                }
            />

            <DataTable
                columns={columns}
                data={healthChecks}
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

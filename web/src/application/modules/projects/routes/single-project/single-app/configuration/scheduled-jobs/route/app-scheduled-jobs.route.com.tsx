import { useMemo } from "react";

import { Plus } from "lucide-react";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { AppScheduledJobsQueries } from "~/projects/data/queries";
import { AppScheduledJobsTableDefs } from "~/projects/module-shared/definitions/tables/app-scheduled-jobs";

import { AppLink, TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { isFeatureDisabledException } from "@infrastructure/api";

import { Button, DataTable } from "@/components/ui";

export function AppScheduledJobsRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { navigate } = useAppNavigate();
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();

    const {
        data: { data: scheduledJobs, meta } = DEFAULT_PAGINATED_DATA,
        isFetching,
        error,
    } = AppScheduledJobsQueries.useFindManyPaginated(
        {
            projectID: projectId,
            appID: appId,
            pagination,
            sorting,
            search,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

    const isFeatureDisabled = error instanceof Error && isFeatureDisabledException(error);
    const columns = useMemo(() => AppScheduledJobsTableDefs.columns(projectId, appId), [projectId, appId]);

    if (isFeatureDisabled) {
        return (
            <div className="flex flex-col gap-6">
                <p className="text-base">
                    Scheduled Jobs feature is disabled, enable it in{" "}
                    <AppLink.Basic
                        to={ROUTE.projects.single.apps.single.configuration.featureSettings.$route(projectId, appId)}
                        className="text-primary underline-offset-4 hover:underline"
                    >
                        Feature Settings
                    </AppLink.Basic>
                </p>
            </div>
        );
    }

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
                                        ROUTE.projects.single.apps.single.configuration.scheduledJobs.create.$route(
                                            projectId,
                                            appId,
                                        ),
                                    );
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

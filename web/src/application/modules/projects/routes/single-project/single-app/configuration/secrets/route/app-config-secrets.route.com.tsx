import { useMemo } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Plus } from "lucide-react";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { ProjectAppSecretsQueries } from "~/projects/data/queries";
import { AppSecretsTableDefs } from "~/projects/module-shared/definitions/tables/app-secrets";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { Button, DataTable } from "@/components/ui";

export function AppConfigSecretsRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { navigate } = useAppNavigate();
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();

    const { data: { data: secrets, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        ProjectAppSecretsQueries.useFindManyPaginated(
            {
                projectID: projectId,
                appID: appId,
                pagination,
                sorting,
                search,
            },
            APP_CONFIGURATION_QUERY_OPTIONS,
        );

    const columns = useMemo(() => AppSecretsTableDefs.columns(projectId, appId), [projectId, appId]);

    return (
        <div className="flex flex-col gap-6">
            <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
                <span className="text-orange-500">Note:</span> Secrets larger than 10 KB cannot be referenced from an
                environment variable using the syntax{" "}
                <span className="text-orange-500">MY_ENV={"${secrets.MY_SECRET}"}</span>.
            </div>

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
                                        ROUTE.projects.single.apps.single.configuration.secrets.create.$route(
                                            projectId,
                                            appId,
                                        ),
                                    );
                                }}
                                disabled={isDenied}
                            >
                                <Plus className="size-4" /> New Secret
                            </Button>
                        )}
                    </PermissionTooltipAction>
                }
            />

            <DataTable
                columns={columns}
                data={secrets}
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

import { useMemo } from "react";

import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppDeploymentsCommands, AppDeploymentsQueries } from "~/projects/data";
import { EAppDeploymentStatus } from "~/projects/module-shared/enums";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { TablePagination } from "@/components/ui";

import { DeploymentSummaryCard, DeploymentSummaryCardSkeleton, useDeploymentCurrentTime } from "../building-blocks";

export function AppDeploymentsRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const { pagination, setPagination, sorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data: { data: deployments, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        AppDeploymentsQueries.useFindManyPaginated({
            projectID: projectId,
            appID: appId,
            pagination,
            sorting,
            search,
        });

    const hasActiveDeployment = useMemo(
        () => deployments.some(deployment => deployment.status === EAppDeploymentStatus.InProgress),
        [deployments],
    );
    const now = useDeploymentCurrentTime(hasActiveDeployment);
    const pageCount = Math.max(1, Math.ceil(meta.page.total / pagination.size));

    const { mutate: cancelDeployment, isPending: isCancelling } = AppDeploymentsCommands.useCancel({
        onSuccess: () => {
            toast.success("Deployment canceled");
        },
    });

    return (
        <section className={cn(listBox)}>
            <div className="flex flex-col gap-5">
                <TableActions search={{ value: search, onChange: setSearch }} />

                <div className="flex flex-col gap-4">
                    {isFetching && deployments.length === 0 ? (
                        <>
                            <DeploymentSummaryCardSkeleton />
                            <DeploymentSummaryCardSkeleton />
                            <DeploymentSummaryCardSkeleton />
                        </>
                    ) : deployments.length === 0 ? (
                        <div className="rounded-lg border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
                            No deployments found.
                        </div>
                    ) : (
                        deployments.map(deployment => (
                            <DeploymentSummaryCard
                                key={deployment.id}
                                deployment={deployment}
                                now={now}
                                isCancelling={isCancelling}
                                onClick={() => {
                                    navigate.modules(
                                        ROUTE.projects.single.apps.single.deployments.details.$route(
                                            projectId,
                                            appId,
                                            deployment.id,
                                        ),
                                    );
                                }}
                                onCancel={deploymentID => {
                                    cancelDeployment({
                                        projectID: projectId,
                                        appID: appId,
                                        deploymentID,
                                    });
                                }}
                            />
                        ))
                    )}
                </div>

                {meta.page.total > 0 && (
                    <TablePagination
                        pageIndex={pagination.page - 1}
                        pageSize={pagination.size}
                        pageCount={pageCount}
                        totalCount={meta.page.total}
                        showPageSizeSelector={false}
                        onPageChange={pageIndex => {
                            setPagination(prev => ({ ...prev, page: pageIndex + 1 }));
                        }}
                    />
                )}
            </div>
        </section>
    );
}

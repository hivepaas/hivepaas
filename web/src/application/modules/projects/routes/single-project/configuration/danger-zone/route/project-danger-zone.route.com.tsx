import type { ReactNode } from "react";

import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectsQueries } from "~/projects/data";
import {
    ProjectDangerAction,
    useConfirmProjectDangerActionDialog,
} from "~/projects/dialogs/confirm-project-danger-action";
import { ProjectPermissionTooltipAction } from "~/projects/module-shared/components";
import { EProjectStatus } from "~/projects/module-shared/enums";

import { AppLoader } from "@application/shared/components";
import { PageError } from "@application/shared/pages";

import { Button } from "@/components/ui/button";

export function ProjectDangerZoneRoute() {
    const { id: projectId } = useParams<{ id: string }>();
    const { actions: confirmAction } = useConfirmProjectDangerActionDialog();

    invariant(projectId, "projectId must be defined");

    const { data, isLoading, error, refetch } = ProjectsQueries.useFindOneById({
        projectID: projectId,
    });

    if (isLoading) {
        return <AppLoader />;
    }

    if (error) {
        return (
            <PageError
                error={error}
                onRetry={() => {
                    void refetch();
                }}
            />
        );
    }

    invariant(data, "project data must be defined");

    const project = data.data;
    const isDisabled = project.status === EProjectStatus.Disabled;
    const isDeleting = project.status === EProjectStatus.Deleting;
    const isStatusActionable = project.status === EProjectStatus.Active || isDisabled;
    const statusAction = isDisabled ? ProjectDangerAction.ReEnable : ProjectDangerAction.Disable;
    const statusButtonLabel = isDisabled ? "Re-enable Project" : "Disable Project";
    const statusButtonVariant = isDisabled ? "default" : "destructive";

    const target = {
        projectId,
        projectName: project.name,
        updateVer: project.updateVer,
    };

    return (
        <div className="flex flex-col gap-5">
            <DangerActionPanel>
                <p className="text-base font-semibold leading-7 text-foreground">
                    Disabling a project disables all apps in the project, so they will no longer consume system
                    resources such as CPU or memory. However, the project&apos;s information will still remain in the
                    system, and you can restore it at any time.
                </p>

                <ProjectPermissionTooltipAction
                    projectId={projectId}
                    action="write"
                >
                    {({ isDenied }) => (
                        <Button
                            variant={statusButtonVariant}
                            disabled={isDenied || !isStatusActionable}
                            className="min-w-[160px]"
                            onClick={() => {
                                if (isDenied || !isStatusActionable) {
                                    return;
                                }

                                confirmAction.open(statusAction, target);
                            }}
                        >
                            {statusButtonLabel}
                        </Button>
                    )}
                </ProjectPermissionTooltipAction>
            </DangerActionPanel>

            <DangerActionPanel>
                <p className="text-base font-semibold leading-7 text-foreground">
                    Deleting a project will remove all of its information, apps, and allocated system resources. You
                    will not be able to recover it after deletion.
                </p>

                <ProjectPermissionTooltipAction
                    projectId={projectId}
                    action="delete"
                >
                    {({ isDenied }) => (
                        <Button
                            variant="destructive"
                            disabled={isDenied || isDeleting}
                            className="min-w-[160px]"
                            onClick={() => {
                                if (isDenied || isDeleting) {
                                    return;
                                }

                                confirmAction.open(ProjectDangerAction.Delete, target);
                            }}
                        >
                            Delete Project
                        </Button>
                    )}
                </ProjectPermissionTooltipAction>
            </DangerActionPanel>
        </div>
    );
}

function DangerActionPanel({ children }: { children: ReactNode }) {
    return (
        <section className="flex flex-col items-start gap-6 rounded-lg border bg-background p-5">{children}</section>
    );
}

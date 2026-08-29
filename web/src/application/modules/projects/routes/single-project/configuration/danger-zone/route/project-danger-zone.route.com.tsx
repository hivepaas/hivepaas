import type { ReactNode } from "react";

import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectsQueries } from "~/projects/data";
import { EnvDangerAction, useConfirmEnvDangerActionDialog } from "~/projects/dialogs/confirm-env-danger-action";
import {
    ProjectDangerAction,
    useConfirmProjectDangerActionDialog,
} from "~/projects/dialogs/confirm-project-danger-action";
import type { ProjectEnvEntity } from "~/projects/domain";
import { ProjectPermissionTooltipAction } from "~/projects/module-shared/components";
import { EProjectEnvStatus, EProjectStatus } from "~/projects/module-shared/enums";

import { AppLoader } from "@application/shared/components";
import { PageError } from "@application/shared/pages";

import { Button } from "@/components/ui/button";

export function ProjectDangerZoneRoute() {
    const { id: projectId } = useParams<{ id: string }>();
    const { actions: confirmProjectAction } = useConfirmProjectDangerActionDialog();
    const { actions: confirmEnvAction } = useConfirmEnvDangerActionDialog();

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

    const { envs } = project;

    return (
        <div className="flex flex-col gap-5">
            {envs.length > 0 && (
                <>
                    <DangerActionPanel>
                        <p className="text-sm leading-6 text-foreground">
                            Disabling a project environment means setting the number of instances of all its
                            applications to 0, so the applications will no longer consume system resources such as CPU
                            or memory. However, the information about the project environment and its applications will
                            still remain in the system, and you can restore them at any time.
                        </p>

                        <div className="flex flex-wrap gap-2">
                            {envs.map(env => (
                                <EnvStatusActionButton
                                    key={env.name}
                                    projectId={projectId}
                                    env={env}
                                    onOpen={action => {
                                        confirmEnvAction.open(action, {
                                            projectId,
                                            envName: env.name,
                                            updateVer: env.updateVer,
                                        });
                                    }}
                                />
                            ))}
                        </div>
                    </DangerActionPanel>

                    <DangerActionPanel>
                        <p className="text-sm leading-6 text-foreground">
                            Deleting a project environment will remove all of its information and applications and
                            release all system resources allocated to it. You will not be able to recover it after
                            deletion.
                        </p>

                        <div className="flex flex-wrap gap-2">
                            {envs.map(env => (
                                <EnvDeleteActionButton
                                    key={env.name}
                                    projectId={projectId}
                                    env={env}
                                    onOpen={() => {
                                        confirmEnvAction.open(EnvDangerAction.Delete, {
                                            projectId,
                                            envName: env.name,
                                            updateVer: env.updateVer,
                                        });
                                    }}
                                />
                            ))}
                        </div>
                    </DangerActionPanel>
                </>
            )}

            <DangerActionPanel>
                <p className="text-sm leading-6 text-foreground">
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
                            className="min-w-[120px]"
                            onClick={() => {
                                if (isDenied || !isStatusActionable) {
                                    return;
                                }

                                confirmProjectAction.open(statusAction, target);
                            }}
                        >
                            {statusButtonLabel}
                        </Button>
                    )}
                </ProjectPermissionTooltipAction>
            </DangerActionPanel>

            <DangerActionPanel>
                <p className="text-sm leading-6 text-foreground">
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
                            className="min-w-[120px]"
                            onClick={() => {
                                if (isDenied || isDeleting) {
                                    return;
                                }

                                confirmProjectAction.open(ProjectDangerAction.Delete, target);
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

function EnvStatusActionButton({
    projectId,
    env,
    onOpen,
}: {
    projectId: string;
    env: ProjectEnvEntity;
    onOpen: (action: EnvDangerAction) => void;
}) {
    const envStatus = env.status ?? EProjectEnvStatus.Active;
    const isEnvDisabled = envStatus === EProjectEnvStatus.Disabled;
    const isEnvDeleting = envStatus === EProjectEnvStatus.Deleting;
    const isEnvStatusActionable = envStatus === EProjectEnvStatus.Active || isEnvDisabled;
    const action = isEnvDisabled ? EnvDangerAction.ReEnable : EnvDangerAction.Disable;
    const buttonLabel = isEnvDisabled ? `Re-enable Env: ${env.name}` : `Disable Env: ${env.name}`;
    const buttonVariant = isEnvDisabled ? "default" : "destructive";

    return (
        <ProjectPermissionTooltipAction
            projectId={projectId}
            action="write"
        >
            {({ isDenied }) => (
                <Button
                    variant={buttonVariant}
                    disabled={isDenied || !isEnvStatusActionable || isEnvDeleting}
                    className="min-w-[120px]"
                    onClick={() => {
                        if (isDenied || !isEnvStatusActionable || isEnvDeleting) {
                            return;
                        }

                        onOpen(action);
                    }}
                >
                    {buttonLabel}
                </Button>
            )}
        </ProjectPermissionTooltipAction>
    );
}

function EnvDeleteActionButton({
    projectId,
    env,
    onOpen,
}: {
    projectId: string;
    env: ProjectEnvEntity;
    onOpen: () => void;
}) {
    const envStatus = env.status ?? EProjectEnvStatus.Active;
    const isEnvDeleting = envStatus === EProjectEnvStatus.Deleting;

    return (
        <ProjectPermissionTooltipAction
            projectId={projectId}
            action="delete"
        >
            {({ isDenied }) => (
                <Button
                    variant="destructive"
                    disabled={isDenied || isEnvDeleting}
                    className="min-w-[120px]"
                    onClick={() => {
                        if (isDenied || isEnvDeleting) {
                            return;
                        }

                        onOpen();
                    }}
                >
                    {`Delete Env: ${env.name}`}
                </Button>
            )}
        </ProjectPermissionTooltipAction>
    );
}

function DangerActionPanel({ children }: { children: ReactNode }) {
    return (
        <section className="flex flex-col items-start gap-6 rounded-lg border bg-background p-5">{children}</section>
    );
}

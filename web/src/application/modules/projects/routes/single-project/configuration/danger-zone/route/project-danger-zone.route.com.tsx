import type { ReactNode } from "react";

import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectsQueries } from "~/projects/data";
import { ProjectsCommands } from "~/projects/data/commands";
import {
    ProjectDangerAction,
    useConfirmProjectDangerActionDialog,
} from "~/projects/dialogs/confirm-project-danger-action";
import type { ProjectEnvEntity } from "~/projects/domain";
import { ProjectPermissionTooltipAction } from "~/projects/module-shared/components";
import { EProjectEnvStatus, EProjectStatus } from "~/projects/module-shared/enums";

import { PopConfirm } from "@application/shared/components";
import { AppLoader } from "@application/shared/components";
import { PageError } from "@application/shared/pages";

import { Button } from "@/components/ui/button";

const DEFAULT_ENV_UPDATE_VER = 0;

export function ProjectDangerZoneRoute() {
    const { id: projectId } = useParams<{ id: string }>();
    const { actions: confirmAction } = useConfirmProjectDangerActionDialog();

    invariant(projectId, "projectId must be defined");

    const { data, isLoading, error, refetch } = ProjectsQueries.useFindOneById({
        projectID: projectId,
    });

    const { mutate: updateEnvStatus, isPending: isUpdatingEnvStatus } = ProjectsCommands.useUpdateEnvStatus({
        onSuccess: (_response, request) => {
            const isReEnable = request.payload.status === EProjectEnvStatus.Active;
            toast.success(isReEnable ? "Environment re-enabled" : "Environment disabled");
        },
    });

    const { mutate: deleteEnv, isPending: isDeletingEnv } = ProjectsCommands.useDeleteEnv({
        onSuccess: () => {
            toast.success("Environment deleted");
        },
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
    const isEnvActionPending = isUpdatingEnvStatus || isDeletingEnv;

    return (
        <div className="flex flex-col gap-5">
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
                                    isPending={isEnvActionPending}
                                    onConfirm={nextStatus => {
                                        console.log("nextStatus", nextStatus);
                                        console.log("env.updateVer", env.updateVer);
                                        console.log("DEFAULT_ENV_UPDATE_VER", DEFAULT_ENV_UPDATE_VER);
                                        updateEnvStatus({
                                            projectID: projectId,
                                            envName: env.name,
                                            payload: {
                                                updateVer: env.updateVer ?? undefined,
                                                status: nextStatus,
                                            },
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
                                    isPending={isEnvActionPending}
                                    onConfirm={() => {
                                        deleteEnv({
                                            projectID: projectId,
                                            envName: env.name,
                                        });
                                    }}
                                />
                            ))}
                        </div>
                    </DangerActionPanel>
                </>
            )}
        </div>
    );
}

function EnvStatusActionButton({
    projectId,
    env,
    isPending,
    onConfirm,
}: {
    projectId: string;
    env: ProjectEnvEntity;
    isPending: boolean;
    onConfirm: (status: EProjectEnvStatus) => void;
}) {
    const envStatus = env.status ?? EProjectEnvStatus.Active;
    const isEnvDisabled = envStatus === EProjectEnvStatus.Disabled;
    const isEnvDeleting = envStatus === EProjectEnvStatus.Deleting;
    const isEnvStatusActionable = envStatus === EProjectEnvStatus.Active || isEnvDisabled;
    const nextStatus = isEnvDisabled ? EProjectEnvStatus.Active : EProjectEnvStatus.Disabled;
    const buttonLabel = isEnvDisabled ? `Re-enable Env: ${env.name}` : `Disable Env: ${env.name}`;
    const buttonVariant = isEnvDisabled ? "default" : "destructive";
    const confirmTitle = isEnvDisabled ? "Re-enable environment" : "Disable environment";
    const confirmDescription = isEnvDisabled
        ? `Re-enable the "${env.name}" environment and restore its applications?`
        : `Disable the "${env.name}" environment and stop all its applications?`;

    return (
        <ProjectPermissionTooltipAction
            projectId={projectId}
            action="write"
        >
            {({ isDenied }) => (
                <PopConfirm
                    title={confirmTitle}
                    description={confirmDescription}
                    confirmText={isEnvDisabled ? "Re-enable" : "Disable"}
                    cancelText="Cancel"
                    variant={buttonVariant}
                    onConfirm={() => {
                        onConfirm(nextStatus);
                    }}
                >
                    <Button
                        variant={buttonVariant}
                        disabled={isDenied || isPending || !isEnvStatusActionable || isEnvDeleting}
                        className="min-w-[160px]"
                    >
                        {buttonLabel}
                    </Button>
                </PopConfirm>
            )}
        </ProjectPermissionTooltipAction>
    );
}

function EnvDeleteActionButton({
    projectId,
    env,
    isPending,
    onConfirm,
}: {
    projectId: string;
    env: ProjectEnvEntity;
    isPending: boolean;
    onConfirm: () => void;
}) {
    const envStatus = env.status ?? EProjectEnvStatus.Active;
    const isEnvDeleting = envStatus === EProjectEnvStatus.Deleting;

    return (
        <ProjectPermissionTooltipAction
            projectId={projectId}
            action="delete"
        >
            {({ isDenied }) => (
                <PopConfirm
                    title="Delete environment"
                    description={`Delete the "${env.name}" environment and all of its applications? This action cannot be undone.`}
                    confirmText="Delete"
                    cancelText="Cancel"
                    variant="destructive"
                    onConfirm={onConfirm}
                >
                    <Button
                        variant="destructive"
                        disabled={isDenied || isPending || isEnvDeleting}
                        className="min-w-[160px]"
                    >
                        {`Delete Env: ${env.name}`}
                    </Button>
                </PopConfirm>
            )}
        </ProjectPermissionTooltipAction>
    );
}

function DangerActionPanel({ children }: { children: ReactNode }) {
    return (
        <section className="flex flex-col items-start gap-6 rounded-lg border bg-background p-5">{children}</section>
    );
}

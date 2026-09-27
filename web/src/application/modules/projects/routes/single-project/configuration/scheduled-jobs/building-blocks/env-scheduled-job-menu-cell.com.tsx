import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { Cog, ExternalLink, MoreVertical, Power, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { EnvScheduledJobsCommands } from "~/projects/data/commands";
import type { EnvScheduledJob } from "~/projects/domain";

import { PopConfirm } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { ESettingStatus } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

/**
 * The actions on a job of an env's list. The env's own jobs are run, enabled,
 * disabled and removed here; an app's job is managed on its app's page.
 */
function View({ projectId, env, job }: Props) {
    const [open, setOpen] = useState(false);
    const { navigate } = useAppNavigate();
    const { canWrite, canDelete } = useConditionalModule({ id: MODULE_IDS.Project });

    const { mutate: runNow, isPending: isRunning } = EnvScheduledJobsCommands.useRunNow({
        onSuccess: response => {
            setOpen(false);
            toast.success("Job sequence started");
            navigate.modules(ROUTE.projects.single.operations.tasks.details.$route(projectId, response.data.task.id));
        },
    });
    const { mutate: updateStatus, isPending: isUpdatingStatus } = EnvScheduledJobsCommands.useUpdateStatus({
        onSuccess: () => {
            setOpen(false);
        },
    });
    const { mutate: deleteOne, isPending: isDeleting } = EnvScheduledJobsCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Job sequence deleted successfully");
            setOpen(false);
        },
    });

    const isEnvJob = job.scope === "project-env";
    const isActive = job.status === ESettingStatus.Active;
    const request = { projectID: projectId, env, scheduledJobID: job.id };

    return (
        <DropdownMenu
            open={open}
            onOpenChange={setOpen}
        >
            <DropdownMenuTrigger asChild>
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                >
                    <MoreVertical className="size-4" />
                    <span className="sr-only">Actions menu</span>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <div className="flex flex-col gap-0">
                    {!isEnvJob && job.ownerApp && (
                        <Button
                            className="justify-start py-1.5"
                            variant="ghost"
                            onClick={() => {
                                navigate.modules(
                                    ROUTE.projects.single.apps.single.configuration.scheduledJobs.$route(
                                        projectId,
                                        env,
                                        job.ownerApp?.id ?? "",
                                    ),
                                );
                            }}
                        >
                            <ExternalLink className="mr-2 size-4" />
                            Open in {job.ownerApp.name}
                        </Button>
                    )}
                    {isEnvJob && (
                        <>
                            <Button
                                className="justify-start py-1.5"
                                variant="ghost"
                                disabled={!canWrite || isRunning}
                                onClick={() => {
                                    runNow(request);
                                }}
                            >
                                <Cog className="mr-2 size-4" />
                                Run Now
                            </Button>
                            <Button
                                className="justify-start py-1.5"
                                variant="ghost"
                                disabled={!canWrite || isUpdatingStatus}
                                onClick={() => {
                                    updateStatus({
                                        ...request,
                                        payload: {
                                            updateVer: job.updateVer,
                                            status: isActive ? ESettingStatus.Disabled : ESettingStatus.Active,
                                            expireAt: job.expireAt,
                                            inheritable: job.inheritable,
                                            default: job.default,
                                        },
                                    });
                                }}
                            >
                                <Power className="mr-2 size-4" />
                                {isActive ? "Disable" : "Enable"}
                            </Button>
                            <PopConfirm
                                title="Remove Job Sequence"
                                variant="destructive"
                                confirmText="Remove"
                                cancelText="Cancel"
                                description="Are you sure you want to remove this job sequence?"
                                onConfirm={() => {
                                    deleteOne(request);
                                }}
                            >
                                <Button
                                    className="justify-start py-1.5"
                                    variant="ghost"
                                    disabled={!canDelete || isDeleting}
                                >
                                    <Trash2Icon className="mr-2 size-4" />
                                    Remove
                                </Button>
                            </PopConfirm>
                        </>
                    )}
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    projectId: string;
    env: string;
    job: EnvScheduledJob;
}

export const EnvScheduledJobMenuCell = memo(View);

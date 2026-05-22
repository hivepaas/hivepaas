import React, { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectAppsCommands } from "~/projects/data/commands";
import { ProjectsQueries } from "~/projects/data/queries";

import { CreateProjectAppForm } from "../form";
import { useCreateProjectAppDialogState } from "../hooks";
import { type CreateProjectAppFormOutput } from "../schemas";

export function CreateProjectAppDialog() {
    const { state, props, ...actions } = useCreateProjectAppDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const open = state.mode !== "closed";
    const { projectId } = state;
    const { data: projectData } = ProjectsQueries.useFindOneById(
        { projectID: projectId ?? "" },
        { enabled: Boolean(projectId) },
    );
    const envs = projectData?.data.envs ?? [];

    const { mutate: createProjectApp, isPending } = ProjectAppsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project app created successfully");
            actions.close();
        },
    });

    // Reset hasChanges when dialog closes
    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
        }
    }, [state.mode]);

    function onSubmit(values: CreateProjectAppFormOutput) {
        if (!projectId) {
            return;
        }

        createProjectApp({
            projectID: projectId,
            name: values.name,
            env: values.env,
            note: values.note,
            tags: values.tags,
        });
    }

    function handleClose(): void {
        if (hasChanges) {
            const userConfirmed: boolean = window.confirm("Are you sure you want to close without saving changes?");
            if (!userConfirmed) {
                return;
            }
        }

        setHasChanges(false);
        actions.close();
    }

    if (!projectId) {
        return null;
    }

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[390px] w-[650px]">
                <DialogHeader>
                    <DialogTitle>Create App</DialogTitle>
                </DialogHeader>
                <CreateProjectAppForm
                    envs={envs}
                    isPending={isPending}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                />
            </DialogContent>
        </Dialog>
    );
}

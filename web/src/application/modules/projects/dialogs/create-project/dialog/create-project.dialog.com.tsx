import React, { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectsCommands } from "~/projects/data/commands";

import { CreateProjectForm } from "../form";
import { useCreateProjectDialogState } from "../hooks";
import { type CreateProjectFormOutput } from "../schemas";

export function CreateProjectDialog() {
    const { state, props, ...actions } = useCreateProjectDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const open = state.mode !== "closed";

    const { mutate: createProject, isPending } = ProjectsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project created successfully");
            actions.close();
        },
    });

    // Reset hasChanges when dialog closes
    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
        }
    }, [state.mode]);

    function onSubmit(values: CreateProjectFormOutput) {
        createProject({
            name: values.name,
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

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[390px] w-[650px]">
                <DialogHeader>
                    <DialogTitle>Create Project</DialogTitle>
                </DialogHeader>
                <CreateProjectForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                />
            </DialogContent>
        </Dialog>
    );
}

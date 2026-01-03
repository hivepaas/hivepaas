import React from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectsCommands } from "~/projects/data/commands";
import { EProjectStatus } from "~/projects/module-shared/enums";

import { CreateProjectForm } from "../form";
import { useCreateProjectDialogState } from "../hooks";
import { type CreateProjectFormOutput } from "../schemas";

export function CreateProjectDialog() {
    const { state, props, ...actions } = useCreateProjectDialogState();

    const open = state.mode !== "closed";

    const { mutate: createProject, isPending } = ProjectsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project created successfully");
            actions.close();
        },
    });

    function onSubmit(values: CreateProjectFormOutput) {
        createProject({
            name: values.name,
            note: values.note,
            tags: values.tags,
        });
    }

    function handleClose(): void {
        actions.close();
    }

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[500px] max-w-[600px]">
                <DialogHeader>
                    <DialogTitle>Create Project</DialogTitle>
                </DialogHeader>
                <CreateProjectForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                />
            </DialogContent>
        </Dialog>
    );
}

import React, { useEffect, useState } from "react";

import { Dialog, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { Separator } from "@components/ui/separator";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import { ProjectAppsCommands } from "~/projects/data/commands";
import { ProjectsQueries } from "~/projects/data/queries";

import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { CreateFunctionForm } from "../form";
import { useCreateFunctionDialogState } from "../hooks";
import { type CreateFunctionFormOutput, createFunctionSource } from "../schemas";

export function CreateFunctionDialog() {
    const navigate = useNavigate();
    const { state, props: dialogOptions, ...actions } = useCreateFunctionDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    const open = state.mode !== "closed";
    const { projectId } = state;
    const { data: projectData } = ProjectsQueries.useFindOneById(
        { projectID: projectId ?? "" },
        { enabled: Boolean(projectId) },
    );
    const envs = projectData?.data.envs ?? [];

    const { mutate: createFunction, isPending } = ProjectAppsCommands.useCreateFunction({
        onSuccess: (response, request) => {
            toast.success("Function created, its first deployment has started");
            actions.close();
            // A function's own page opens on its code.
            void navigate(
                ROUTE.projects.single.apps.single.code.$route(request.projectID, request.env, response.data.id),
            );
        },
    });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
        }
    }, [state.mode]);

    function onSubmit(values: CreateFunctionFormOutput) {
        if (!projectId || !canWrite) {
            return;
        }

        createFunction({
            projectID: projectId,
            name: values.name,
            env: values.env,
            note: "",
            tags: [],
            source: createFunctionSource(values),
            ...(values.expose ? { domain: values.domain } : {}),
        });
    }

    function handleClose(): void {
        if (canWrite && hasChanges) {
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
            <DialogFixedContent className="sm:max-w-[700px]">
                <DialogHeader>
                    <DialogTitle>Create Function</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                <CreateFunctionForm
                    projectId={projectId}
                    envs={envs}
                    initialEnv={dialogOptions?.initialEnv}
                    isPending={isPending}
                    readOnly={!canWrite}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                />
            </DialogFixedContent>
        </Dialog>
    );
}

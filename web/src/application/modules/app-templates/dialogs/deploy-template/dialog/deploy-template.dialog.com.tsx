import React, { useState } from "react";

import { MODULE_IDS, ROUTE } from "@/application/shared/constants";
import { useConditionalModule } from "@/application/shared/permissions";
import { Boxes, Loader2 } from "lucide-react";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import { ProjectClusterVolumesQueries, ProjectsQueries } from "~/projects/data/queries";

import { Dialog, DialogFixedContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";

import { type CreateAppFromTemplateReq, type CreateAppFromTemplateResp } from "../../../api";
import { useCreateAppFromTemplate, useGetAppTemplate } from "../../../data";
import { DeployTemplateForm } from "../form";
import { useDeployTemplateDialogState } from "../hooks";

export function DeployTemplateDialog() {
    const { state, props: dialogOptions, ...actions } = useDeployTemplateDialogState();
    const navigate = useNavigate();
    const [imageError, setImageError] = useState(false);

    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const open = state.mode !== "closed";
    const { projectId, templateName } = state;

    // 1. Fetch template detail
    const { data: template, isLoading: isTemplateLoading } = useGetAppTemplate(templateName ?? undefined);

    // 2. Fetch project environments
    const { data: projectData } = ProjectsQueries.useFindOneById(
        { projectID: projectId ?? "" },
        { enabled: Boolean(projectId) },
    );
    const envs = projectData?.data.envs ?? [];

    // 3. Fetch cluster volumes
    const { data: clusterVolumesData } = ProjectClusterVolumesQueries.useFindManyPaginated(
        { projectID: projectId ?? "" },
        { enabled: Boolean(projectId) },
    );
    const clusterVolumes =
        clusterVolumesData?.data.map(v => ({
            id: v.id,
            name: v.name,
        })) ?? [];

    // 4. Create App From Template Mutation
    const { mutate: createAppFromTemplate, isPending } = useCreateAppFromTemplate({
        onSuccess: (res: CreateAppFromTemplateResp, variables: CreateAppFromTemplateReq) => {
            toast.success(`Application "${variables.name}" deployed successfully!`);
            actions.close();
            if (projectId && variables.projectEnv && res.data.app.id) {
                // Navigate to newly created app
                void navigate(
                    ROUTE.projects.single.apps.single.deployments.$route(
                        projectId,
                        variables.projectEnv,
                        res.data.app.id,
                    ),
                );
            }
        },
        onError: (err: Error) => {
            toast.error(err.message !== "" ? err.message : "Failed to deploy application from template");
        },
    });

    const handleSubmit = (values: CreateAppFromTemplateReq) => {
        if (!projectId || !canWrite || isPending) return;
        createAppFromTemplate(values);
    };

    const handleClose = () => {
        actions.close();
    };

    if (!projectId || !templateName) {
        return null;
    }

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="sm:max-w-[800px] w-full">
                <DialogHeader className="border-b border-border/50 px-3.5 py-4">
                    <div className="flex items-center gap-3.5">
                        <div className="flex size-10 shrink-0 items-center justify-center rounded-lg border border-border/80 bg-muted/30 p-2 overflow-hidden shadow-2xs">
                            {!imageError && template?.iconUrl ? (
                                <img
                                    src={template.iconUrl}
                                    alt={template.title}
                                    onError={() => {
                                        setImageError(true);
                                    }}
                                    className="size-full object-contain"
                                />
                            ) : (
                                <Boxes className="size-5 text-amber-500" />
                            )}
                        </div>

                        <div className="min-w-0 flex-1">
                            <DialogTitle className="text-base font-semibold text-foreground tracking-tight">
                                Deploy Template: {template?.title ?? templateName}
                            </DialogTitle>
                            {template?.tagline && (
                                <p className="text-xs text-muted-foreground mt-0.5 truncate">{template.tagline}</p>
                            )}
                        </div>
                    </div>
                </DialogHeader>

                {isTemplateLoading || !template ? (
                    <div className="px-3.5 py-6 space-y-4">
                        <div className="flex items-center justify-center py-8 gap-3 text-sm text-muted-foreground">
                            <Loader2 className="size-5 animate-spin text-amber-500" />
                            <span>Loading template configuration...</span>
                        </div>
                        <Skeleton className="h-20 w-full rounded-xl" />
                        <Skeleton className="h-40 w-full rounded-xl" />
                    </div>
                ) : (
                    <DeployTemplateForm
                        template={template}
                        projectId={projectId}
                        envs={envs}
                        clusterVolumes={clusterVolumes}
                        initialEnv={dialogOptions.initialEnv}
                        initialVersion={dialogOptions.initialVersion}
                        initialVariant={dialogOptions.initialVariant}
                        isPending={isPending}
                        readOnly={!canWrite}
                        onSubmit={handleSubmit}
                        onCancel={handleClose}
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}

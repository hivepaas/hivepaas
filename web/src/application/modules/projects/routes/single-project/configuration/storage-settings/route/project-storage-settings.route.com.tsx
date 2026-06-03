import { useRef } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectStorageSettingsCommands, ProjectStorageSettingsQueries } from "~/projects/data";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { ProjectStorageSettingsForm } from "../form";
import type { ProjectStorageSettingsFormSchemaOutput } from "../schemas";
import type { ProjectStorageSettingsFormRef } from "../types";

function NoteBox({ children }: { children: React.ReactNode }) {
    return (
        <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
            <span className="text-orange-500">Note: </span>
            {children}
        </div>
    );
}

export function ProjectStorageSettingsRoute() {
    const { id: projectId } = useParams<{ id: string }>();
    const formRef = useRef<ProjectStorageSettingsFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    invariant(projectId, "projectId must be defined");
    const resolvedProjectId = projectId;

    const settingsQuery = ProjectStorageSettingsQueries.useFindOne({ projectID: resolvedProjectId });

    const { mutate: update, isPending: isUpdating } = ProjectStorageSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Storage settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });

    function handleSubmit(values: ProjectStorageSettingsFormSchemaOutput) {
        if (!canWrite) {
            return;
        }

        const settings = settingsQuery.data?.data;
        invariant(settings, "storage settings must be defined");

        update({
            projectID: resolvedProjectId,
            payload: {
                updateVer: settings.updateVer,
                bindSettings: {
                    enabled: values.bindSettings.enabled,
                    baseDirs: values.bindSettings.baseDirs.map(item => item.value).filter(Boolean),
                    subpathTemplate: values.bindSettings.subpathTemplate,
                },
                volumeSettings: {
                    enabled: values.volumeSettings.enabled,
                    volumes: values.volumeSettings.volumes.map(volume => ({ id: volume.id })),
                    subpathTemplate: values.volumeSettings.subpathTemplate,
                },
                clusterVolumeSettings: {
                    enabled: values.clusterVolumeSettings.enabled,
                    volumes: values.clusterVolumeSettings.volumes.map(volume => ({ id: volume.id })),
                    subpathTemplate: values.clusterVolumeSettings.subpathTemplate,
                },
                tmpfsSettings: {
                    enabled: values.tmpfsSettings.enabled,
                    maxSize: values.tmpfsSettings.maxSize,
                },
            },
        });
    }

    if (settingsQuery.isLoading) {
        return <AppLoader />;
    }

    if (settingsQuery.error) {
        return (
            <PageError
                error={settingsQuery.error}
                onRetry={settingsQuery.refetch}
            />
        );
    }

    invariant(settingsQuery.data, "storage settings data must be defined");

    return (
        <ProjectStorageSettingsForm
            ref={formRef}
            defaultValues={settingsQuery.data.data}
            onSubmit={handleSubmit}
            readOnly={!canWrite}
            note={
                <NoteBox>
                    The configurations for persistent storage here are constraints or mandatory requirements that all
                    apps in the project must comply with. Additionally, if your cluster has multiple nodes, make sure
                    that the persistent storages are accessible from all of those nodes; otherwise, your apps will not
                    function properly.
                </NoteBox>
            }
        >
            <div className="flex justify-end mt-4">
                <ProjectPermissionSubmitButton isPending={isUpdating} />
            </div>
        </ProjectStorageSettingsForm>
    );
}

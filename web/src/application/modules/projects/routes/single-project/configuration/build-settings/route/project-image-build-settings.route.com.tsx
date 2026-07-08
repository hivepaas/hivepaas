import { useRef, useState } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectImageBuildSettingsCommands, ProjectImageBuildSettingsQueries } from "~/projects/data";
import type { ProjectImageBuildRepoCacheClearResult } from "~/projects/domain";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { ClearRepoCacheResultDialog } from "../building-blocks";
import { ProjectImageBuildSettingsForm } from "../form";
import type { ProjectImageBuildSettingsFormSchemaOutput } from "../schemas";
import type { ProjectImageBuildSettingsFormRef } from "../types";

function NoteBox({ children }: { children: React.ReactNode }) {
    return (
        <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
            <span className="text-orange-500">Note: </span>
            {children}
        </div>
    );
}

export function ProjectImageBuildSettingsRoute() {
    const { id: projectId } = useParams<{ id: string }>();
    const formRef = useRef<ProjectImageBuildSettingsFormRef>(null);
    const [hasQueriedCache, setHasQueriedCache] = useState(false);
    const [clearRepoCacheResult, setClearRepoCacheResult] = useState<ProjectImageBuildRepoCacheClearResult | null>(
        null,
    );
    const { canWrite, canExecute } = useConditionalModule({ id: MODULE_IDS.Project });

    invariant(projectId, "projectId must be defined");
    const resolvedProjectId = projectId;

    const settingsQuery = ProjectImageBuildSettingsQueries.useFindOne({ projectID: resolvedProjectId });
    const repoCacheQuery = ProjectImageBuildSettingsQueries.useFindRepoCache(
        { projectID: resolvedProjectId },
        {
            enabled: false,
        },
    );

    const { mutate: update, isPending: isUpdating } = ProjectImageBuildSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Image build settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });

    const { mutate: clearRepoCache, isPending: isClearingRepoCache } =
        ProjectImageBuildSettingsCommands.useClearRepoCache({
            onSuccess: response => {
                setClearRepoCacheResult(response.data);
                setHasQueriedCache(true);
                void repoCacheQuery.refetch();
            },
        });

    function handleSubmit(values: ProjectImageBuildSettingsFormSchemaOutput) {
        if (!canWrite) {
            return;
        }

        const settings = settingsQuery.data?.data;
        invariant(settings, "image build settings must be defined");

        update({
            projectID: resolvedProjectId,
            payload: {
                updateVer: settings.updateVer,
                resources: values.resources,
                sources: values.sources,
                noCache: values.noCache,
                noVerbose: values.noVerbose,
            },
        });
    }

    function handleQueryRepoCache() {
        void repoCacheQuery.refetch().then(result => {
            if (result.data) {
                setHasQueriedCache(true);
            }
        });
    }

    function handleClearRepoCache() {
        if (!canExecute) {
            return;
        }

        clearRepoCache({ projectID: resolvedProjectId });
    }

    function handleClearRepoCacheResultOpenChange(open: boolean) {
        if (!open) {
            setClearRepoCacheResult(null);
        }
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

    invariant(settingsQuery.data, "image build settings data must be defined");

    return (
        <>
            <ProjectImageBuildSettingsForm
                ref={formRef}
                defaultValues={settingsQuery.data.data}
                onSubmit={handleSubmit}
                readOnly={!canWrite}
                cacheInfo={repoCacheQuery.data?.data}
                cacheInfoControls={{
                    hasQueried: hasQueriedCache,
                    isQuerying: repoCacheQuery.isFetching,
                    isClearing: isClearingRepoCache,
                    readOnly: !canExecute,
                    note: (
                        <NoteBox>
                            Enabling the cache feature can significantly reduce the application deployment time if your
                            repository is large. However, this will consume space on your hard drive.
                        </NoteBox>
                    ),
                    footer: (
                        <FormActionBar>
                            <ProjectPermissionSubmitButton isPending={isUpdating} />
                        </FormActionBar>
                    ),
                    onQuery: handleQueryRepoCache,
                    onClear: handleClearRepoCache,
                }}
            >
                <NoteBox>
                    The image build process can consume a large amount of your system&apos;s resources and may affect
                    running applications. It is recommended that you set resource limits for this process. The system
                    uses a default configuration, but you can reconfigure it here.
                </NoteBox>
            </ProjectImageBuildSettingsForm>

            <ClearRepoCacheResultDialog
                open={Boolean(clearRepoCacheResult)}
                result={clearRepoCacheResult}
                onOpenChange={handleClearRepoCacheResultOpenChange}
            />
        </>
    );
}

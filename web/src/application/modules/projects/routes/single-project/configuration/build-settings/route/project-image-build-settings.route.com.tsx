import { useRef, useState } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectImageBuildSettingsCommands, ProjectImageBuildSettingsQueries } from "~/projects/data";
import type { ProjectImageBuildRepoCacheClearResult } from "~/projects/domain";

import { AppLink, AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { useConditionalModule } from "@application/shared/permissions";

import { Button } from "@/components/ui";

import { ClearRepoCacheResultDialog } from "../building-blocks";
import { ProjectImageBuildSettingsForm } from "../form";
import type { ProjectImageBuildSettingsFormRef } from "../types";

function NoteBox({ children }: { children: React.ReactNode }) {
    return (
        <div className={cn(dashedBorderBox, "text-sm leading-6")}>
            <span className="font-semibold text-orange-500">Note: </span>
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

    const { mutate: clearRepoCache, isPending: isClearingRepoCache } =
        ProjectImageBuildSettingsCommands.useClearRepoCache({
            onSuccess: response => {
                setClearRepoCacheResult(response.data);
                setHasQueriedCache(true);
                void repoCacheQuery.refetch();
            },
        });

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
                readOnly={!canWrite}
                headerNote={
                    <NoteBox>
                        Image build configuration is only available at the global scope,{" "}
                        <AppLink.Basic
                            to={ROUTE.appSettings.imageBuild.$route}
                            className="font-medium text-primary underline-offset-4 hover:underline"
                        >
                            view here
                        </AppLink.Basic>
                        .
                    </NoteBox>
                }
                cacheInfo={repoCacheQuery.data?.data}
                cacheInfoControls={{
                    hasQueried: hasQueriedCache,
                    isQuerying: repoCacheQuery.isFetching,
                    isClearing: isClearingRepoCache,
                    readOnly: !canExecute,
                    note: <NoteBox>The repository sources cache information below is only for this project.</NoteBox>,
                    footer: (
                        <FormActionBar>
                            <Button
                                type="button"
                                className="min-w-[100px]"
                                disabled
                            >
                                Save
                            </Button>
                        </FormActionBar>
                    ),
                    onQuery: handleQueryRepoCache,
                    onClear: handleClearRepoCache,
                }}
            />

            <ClearRepoCacheResultDialog
                open={Boolean(clearRepoCacheResult)}
                result={clearRepoCacheResult}
                onOpenChange={handleClearRepoCacheResultOpenChange}
            />
        </>
    );
}

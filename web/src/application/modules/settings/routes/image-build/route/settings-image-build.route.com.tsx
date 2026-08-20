import { useRef, useState } from "react";

import { Button } from "@components/ui";
import { dashedBorderBox, listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ImageBuildSettingsCommands, ImageBuildSettingsQueries } from "~/settings/data";
import type { ImageBuildRepoCacheClearResult } from "~/settings/domain";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { ClearRepoCacheResultDialog } from "../building-blocks";
import { SettingsImageBuildForm } from "../form";
import type { SettingsImageBuildFormSchemaOutput } from "../schemas";
import type { SettingsImageBuildFormRef } from "../types";

function NoteBox({ children }: { children: React.ReactNode }) {
    return (
        <div className={cn(dashedBorderBox, "text-sm leading-6")}>
            <span className="text-orange-500">Note: </span>
            {children}
        </div>
    );
}

export function SettingsImageBuildRoute() {
    const formRef = useRef<SettingsImageBuildFormRef>(null);
    const [hasQueriedCache, setHasQueriedCache] = useState(false);
    const [clearRepoCacheResult, setClearRepoCacheResult] = useState<ImageBuildRepoCacheClearResult | null>(null);
    const { canWrite, canExecute } = useConditionalModule({ id: MODULE_IDS.Settings });

    const settingsQuery = ImageBuildSettingsQueries.useFindOne();
    const repoCacheQuery = ImageBuildSettingsQueries.useFindRepoCache({
        enabled: false,
    });

    const { mutate: update, isPending: isUpdating } = ImageBuildSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Image build settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });

    const { mutate: clearRepoCache, isPending: isClearingRepoCache } = ImageBuildSettingsCommands.useClearRepoCache({
        onSuccess: response => {
            setClearRepoCacheResult(response.data);
            setHasQueriedCache(true);
            void repoCacheQuery.refetch();
        },
    });

    function handleSubmit(values: SettingsImageBuildFormSchemaOutput) {
        if (!canWrite) {
            return;
        }

        const settings = settingsQuery.data?.data;
        invariant(settings, "image build settings must be defined");

        update({
            payload: {
                updateVer: settings.updateVer,
                inheritable: settings.inheritable ?? true,
                default: settings.default ?? true,
                workers: {
                    nodes: values.workers.nodes.map(node => ({ id: node.id })),
                    nodeLabels: values.workers.nodeLabels,
                    maxParallelism: values.workers.maxParallelism,
                },
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

        clearRepoCache({});
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
        <div className={cn(listBox)}>
            <SettingsImageBuildForm
                ref={formRef}
                defaultValues={settingsQuery.data.data}
                onSubmit={handleSubmit}
                readOnly={!canWrite}
                cacheInfo={repoCacheQuery.data?.data}
                workerNote={
                    <NoteBox>
                        Image builds consume significant CPU and RAM. We recommend configuring dedicated build nodes to
                        keep your manager and application workloads running smoothly.
                    </NoteBox>
                }
                cacheInfoControls={{
                    hasQueried: hasQueriedCache,
                    isQuerying: repoCacheQuery.isFetching,
                    isClearing: isClearingRepoCache,
                    readOnly: !canExecute,
                    note: (
                        <NoteBox>
                            Enabling the cache feature can significantly reduce the application deployment time if your
                            repository is large. However, this will consume space on your storage.
                        </NoteBox>
                    ),
                    footer: (
                        <FormActionBar>
                            <PermissionTooltipAction
                                id={MODULE_IDS.Settings}
                                action="write"
                            >
                                {({ isDenied }) => (
                                    <Button
                                        type="submit"
                                        className="min-w-[100px]"
                                        disabled={isUpdating || isDenied}
                                        isLoading={isUpdating}
                                    >
                                        Save
                                    </Button>
                                )}
                            </PermissionTooltipAction>
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
        </div>
    );
}

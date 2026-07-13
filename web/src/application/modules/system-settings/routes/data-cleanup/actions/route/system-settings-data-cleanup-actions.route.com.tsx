import { type ReactNode, useState } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import { SystemCleanupCommands, SystemCleanupQueries } from "~/system-settings/data";
import type { SystemCleanupBuildCacheClearResult, SystemCleanupRepoCacheClearResult } from "~/system-settings/domain";
import { ActionExecutePanel } from "~/system-settings/module-shared";

import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";
import { getFriendlyDataSize } from "@application/shared/utils/data-size";

import { ClearCacheResultDialog, type ClearCacheResultRow } from "../building-blocks";

type ClearCacheResult =
    | {
          type: "repo";
          data: SystemCleanupRepoCacheClearResult;
      }
    | {
          type: "build";
          data: SystemCleanupBuildCacheClearResult;
      };

function formatDataSize(value?: number): string {
    return getFriendlyDataSize(value) || "0 B";
}

function getResultDialogProps(result: ClearCacheResult | null): { title: string; rows: ClearCacheResultRow[] } {
    if (result?.type === "repo") {
        return {
            title: "Repo cache cleared",
            rows: [
                { label: "Files deleted", value: result.data.filesDeleted },
                { label: "Space Reclaimed", value: formatDataSize(result.data.spaceReclaimed) },
            ],
        };
    }

    if (result?.type === "build") {
        return {
            title: "Build cache cleared",
            rows: [
                { label: "Caches deleted", value: result.data.cachesDeleted },
                { label: "Space Reclaimed", value: formatDataSize(result.data.spaceReclaimed) },
            ],
        };
    }

    return {
        title: "",
        rows: [],
    };
}

function CacheActionPanel({ children }: { children: ReactNode }) {
    return <div className="rounded-lg border bg-background p-4">{children}</div>;
}

export function SystemSettingsDataCleanupActionsRoute() {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const [hasQueriedRepoCache, setHasQueriedRepoCache] = useState(false);
    const [clearCacheResult, setClearCacheResult] = useState<ClearCacheResult | null>(null);

    const repoCacheQuery = SystemCleanupQueries.useFindRepoCache(
        {},
        {
            enabled: false,
        },
    );

    const { mutate: execute, isPending } = SystemCleanupCommands.useExecute({
        onSuccess: () => {
            toast.success("Cleanup started");
        },
    });

    const { mutate: clearRepoCache, isPending: isClearingRepoCache } = SystemCleanupCommands.useClearRepoCache({
        onSuccess: response => {
            setClearCacheResult({ type: "repo", data: response.data });
            setHasQueriedRepoCache(true);
            void repoCacheQuery.refetch();
        },
    });

    const { mutate: clearBuildCache, isPending: isClearingBuildCache } = SystemCleanupCommands.useClearBuildCache({
        onSuccess: response => {
            setClearCacheResult({ type: "build", data: response.data });
        },
    });

    const totalFiles = repoCacheQuery.data?.data.totalFiles ?? 0;
    const totalSize = getFriendlyDataSize(repoCacheQuery.data?.data.totalSizeBytes) || "-";
    const canClearRepoCache = hasQueriedRepoCache && totalFiles > 0;
    const dialogProps = getResultDialogProps(clearCacheResult);

    function handleQueryRepoCache() {
        void repoCacheQuery.refetch().then(result => {
            if (result.data) {
                setHasQueriedRepoCache(true);
            }
        });
    }

    function handleClearRepoCache() {
        if (!canWrite) {
            return;
        }

        clearRepoCache({});
    }

    function handleClearBuildCache() {
        if (!canWrite) {
            return;
        }

        clearBuildCache({});
    }

    function handleResultDialogOpenChange(open: boolean) {
        if (!open) {
            setClearCacheResult(null);
        }
    }

    return (
        <>
            <div className="flex flex-col gap-6">
                <ActionExecutePanel
                    message="Make sure you have enabled the cleanup job before performing this action."
                    buttonLabel="Run Cleanup Now"
                    isLoading={isPending}
                    permissionModuleId={MODULE_IDS.System}
                    onExecute={() => {
                        if (!canWrite) {
                            return;
                        }

                        execute({});
                    }}
                />

                <CacheActionPanel>
                    <div className="flex flex-col items-start gap-6">
                        <p className="text-base text-foreground">
                            Using the repository cache can significantly reduce your application deployment time.
                            However, it will consume system storage space.
                        </p>

                        <div className="flex min-h-9 flex-wrap items-center gap-x-10 gap-y-4">
                            {hasQueriedRepoCache && (
                                <>
                                    <span>Total Files: {totalFiles}</span>
                                    <span>Total Size: {totalSize}</span>
                                </>
                            )}

                            {canClearRepoCache && (
                                <PermissionTooltipAction
                                    id={MODULE_IDS.System}
                                    action="write"
                                >
                                    {({ isDenied }) => (
                                        <Button
                                            type="button"
                                            className="min-w-[220px]"
                                            disabled={isDenied}
                                            isLoading={isClearingRepoCache}
                                            onClick={handleClearRepoCache}
                                        >
                                            Force Clear Repo Cache
                                        </Button>
                                    )}
                                </PermissionTooltipAction>
                            )}

                            <Button
                                type="button"
                                variant="link"
                                className="px-0 text-primary underline-offset-4 hover:underline"
                                isLoading={repoCacheQuery.isFetching}
                                onClick={handleQueryRepoCache}
                            >
                                Query Repo Cache Info
                            </Button>
                        </div>
                    </div>
                </CacheActionPanel>

                <CacheActionPanel>
                    <div className="flex flex-col items-start gap-6">
                        <p className="text-base text-foreground">
                            Docker uses the build cache to reduce image build time for subsequent builds. You can clear
                            this cache to reclaim storage space.
                        </p>

                        <PermissionTooltipAction
                            id={MODULE_IDS.System}
                            action="write"
                        >
                            {({ isDenied }) => (
                                <Button
                                    type="button"
                                    className="min-w-[220px]"
                                    disabled={isDenied}
                                    isLoading={isClearingBuildCache}
                                    onClick={handleClearBuildCache}
                                >
                                    Force Clear Build Cache
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    </div>
                </CacheActionPanel>
            </div>

            <ClearCacheResultDialog
                open={Boolean(clearCacheResult)}
                title={dialogProps.title}
                rows={dialogProps.rows}
                onOpenChange={handleResultDialogOpenChange}
            />
        </>
    );
}

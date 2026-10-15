import { useState } from "react";

import { dashedBorderBox, inlineLink, listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { AlertTriangle, ScanSearch, ShieldCheck } from "lucide-react";
import { Link, useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import type { VulnFinding, VulnIgnoreScope } from "~/projects/api/services";
import { AppImageScanCommands, AppImageScanQueries } from "~/projects/data";
import {
    FindingsTable,
    IgnoreFindingDialog,
    IgnoredFindings,
    ScannedImages,
    SeverityCounters,
} from "~/projects/module-shared/components";

import { AppLoader, BetaBadge } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Button } from "@/components/ui/button";

/** Saves a downloaded file under its name. */
function saveBlob(blob: Blob, filename: string) {
    const url = window.URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    anchor.click();
    window.URL.revokeObjectURL(url);
}

// How often the tab asks again while a scan of the app is queued or running.
const PENDING_SCAN_REFETCH_INTERVAL_MS = 3_000;

/**
 * What the app's images carry: the vulnerabilities Trivy found in the images
 * its tasks run, after the ignores of its scopes. Scanned after each
 * deployment, every day, and when asked.
 */
export function AppSecurityRoute() {
    const { id: projectID, env, appId: appID } = useParams<{ id: string; env: string; appId: string }>();
    invariant(projectID, "projectID must be defined");
    invariant(env, "env must be defined");
    invariant(appID, "appID must be defined");
    const appRef = { projectID, env, appID };

    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { canWrite: canWriteSettings } = useConditionalModule({ id: MODULE_IDS.Settings });
    const { canRead: canReadSystem } = useConditionalModule({ id: MODULE_IDS.System });
    // The scopes the user may set or remove an ignore at, as the server checks:
    // a global one is the Settings module's, the others the project's.
    const writableScopes: VulnIgnoreScope[] = [
        ...(canWrite ? (["app", "project-env", "project"] as const) : []),
        ...(canWriteSettings ? (["global"] as const) : []),
    ];
    const [ignoring, setIgnoring] = useState<VulnFinding | null>(null);

    const { data, isLoading } = AppImageScanQueries.useFindOne(appRef, {
        refetchInterval: query => (query.state.data?.data.task ? PENDING_SCAN_REFETCH_INTERVAL_MS : false),
    });
    const { mutate: scan, isPending: isStarting } = AppImageScanCommands.useScan({
        onSuccess: () => {
            toast.success("Scan queued");
        },
    });
    const { mutate: createIgnore, isPending: isIgnoring } = AppImageScanCommands.useCreateIgnore({
        onSuccess: () => {
            toast.success("Vulnerability ignored");
            setIgnoring(null);
        },
    });
    const {
        mutate: downloadSBOM,
        isPending: isDownloading,
        variables: downloadingSBOM,
    } = AppImageScanCommands.useDownloadSBOM({
        onSuccess: res => {
            saveBlob(res.data.blob, res.data.filename ?? "sbom.cdx.json");
        },
    });
    const { mutate: deleteIgnore } = AppImageScanCommands.useDeleteIgnore({
        onSuccess: () => {
            toast.success("Ignore removed");
        },
    });

    if (isLoading || !data) {
        return <AppLoader />;
    }
    const result = data.data;
    const scanning = result.task !== null;

    return (
        <div className={cn(listBox, "flex flex-col gap-5")}>
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex flex-col gap-0.5">
                    <h3 className="flex items-center gap-2 text-base font-semibold">
                        <ShieldCheck className="size-4.5 text-muted-foreground" />
                        Image vulnerabilities
                        <BetaBadge />
                    </h3>
                    <span className="text-[13px] text-muted-foreground">
                        {scanning
                            ? "A scan is running…"
                            : result.scannedAt
                              ? `Last looked at ${timeAgoFormatter.format(result.scannedAt)}`
                              : "Not scanned yet"}
                    </span>
                </div>
                {result.enabled && (
                    <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        className="h-8 self-start sm:self-auto"
                        isLoading={isStarting || scanning}
                        disabled={!canWrite || isStarting || scanning}
                        onClick={() => {
                            scan(appRef);
                        }}
                    >
                        <ScanSearch className="size-3.5" />
                        {scanning ? "Scanning" : "Scan now"}
                    </Button>
                )}
            </div>

            {!result.enabled && (
                <div className={cn(dashedBorderBox, "flex items-start gap-2")}>
                    <AlertTriangle className="mt-0.5 size-4 shrink-0 text-amber-600 dark:text-amber-500" />
                    <span>
                        Image scanning is switched off: nothing is scanned
                        {result.scannedAt ? ", and what is shown is from the last scan" : ""}.{" "}
                        {canReadSystem ? (
                            <Link
                                to={ROUTE.systemSettings.imageScanning.configuration.$route}
                                className={inlineLink}
                            >
                                Switch it on in System &gt; Image Scanning
                            </Link>
                        ) : (
                            "An administrator switches it on in System > Image Scanning."
                        )}
                    </span>
                </div>
            )}

            {result.error && (
                <div className="flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/5 px-3 py-2 text-[13px] text-destructive">
                    <AlertTriangle className="mt-0.5 size-4 shrink-0" />
                    <span className="break-words">
                        {result.images.length > 0 ? "Last scan incomplete" : "Not scanned"}: {result.error}
                    </span>
                </div>
            )}

            {result.images.length === 0 ? (
                result.enabled &&
                !result.error && (
                    <div className="rounded-md border px-4 py-6 text-center text-sm text-muted-foreground">
                        {scanning
                            ? "The app's images are being scanned."
                            : "The app's images have not been scanned yet: they are, after its next deployment or with Scan now."}
                    </div>
                )
            ) : (
                <>
                    <SeverityCounters summary={result.summary} />
                    <ScannedImages
                        images={result.images}
                        onDownloadSBOM={
                            result.enabled
                                ? image => {
                                      downloadSBOM({ ...appRef, imageId: image.imageId });
                                  }
                                : undefined
                        }
                        downloading={isDownloading ? downloadingSBOM.imageId : undefined}
                    />
                    <FindingsTable
                        findings={result.findings}
                        onIgnore={writableScopes.length > 0 ? setIgnoring : undefined}
                    />
                    <IgnoredFindings
                        ignored={result.ignored}
                        canRemove={ignore => writableScopes.includes(ignore.scope)}
                        onRemove={
                            writableScopes.length > 0
                                ? ignore => {
                                      // An app-scope ignore may be the parent app's, which the
                                      // app inherits: it is removed where it is set.
                                      const ownerAppID =
                                          ignore.scope === "app" && ignore.objectId ? ignore.objectId : appRef.appID;
                                      deleteIgnore({
                                          ...appRef,
                                          appID: ownerAppID,
                                          scope: ignore.scope,
                                          ignoreID: ignore.id,
                                      });
                                  }
                                : undefined
                        }
                    />
                </>
            )}

            <IgnoreFindingDialog
                finding={ignoring}
                scopes={writableScopes}
                isPending={isIgnoring}
                onOpenChange={open => {
                    if (!open) {
                        setIgnoring(null);
                    }
                }}
                onConfirm={(scope, payload) => {
                    createIgnore({ ...appRef, scope, payload });
                }}
            />
        </div>
    );
}

import { useState } from "react";

import { AlertTriangle, ArrowLeft } from "lucide-react";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import type { VulnFinding } from "~/projects/api/services";
import { AppImageScanCommands } from "~/projects/data";
import {
    FindingsTable,
    IgnoreFindingDialog,
    IgnoredFindings,
    ScannedImages,
    SeverityCounters,
} from "~/projects/module-shared/components";
import { SystemImageScanQueries } from "~/system-settings/data";

import { AppLink, AppLoader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

// An ignore of HivePaaS's own images is global: it names no app.
const NO_APP = { projectID: "", env: "", appID: "" };

/**
 * What a part of HivePaaS ran when last scanned, as an app's Security tab
 * shows it: a finding is set aside by a global ignore, for HivePaaS alone or
 * for every app too.
 */
export function SystemSettingsImageScanningHivePaaSComponentRoute() {
    const { component } = useParams<{ component: string }>();
    invariant(component, "component must be defined");
    // Its ignores are global: the Settings module's to write, as the server checks.
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Settings });
    const [ignoring, setIgnoring] = useState<VulnFinding | null>(null);
    const { data, isLoading, refetch } = SystemImageScanQueries.useFindComponent({ component });
    const { mutate: createIgnore, isPending: isIgnoring } = AppImageScanCommands.useCreateIgnore({
        onSuccess: () => {
            toast.success("Vulnerability ignored");
            setIgnoring(null);
            void refetch();
        },
    });
    const { mutate: deleteIgnore } = AppImageScanCommands.useDeleteIgnore({
        onSuccess: () => {
            toast.success("Ignore removed");
            void refetch();
        },
    });

    if (isLoading || !data) {
        return <AppLoader />;
    }
    const { scan } = data.data;

    return (
        <div className="flex flex-col gap-5">
            <div className="flex flex-col gap-1">
                <AppLink.Basic
                    to={ROUTE.systemSettings.imageScanning.hivepaas.$route}
                    className="inline-flex w-fit items-center gap-1 text-[13px] text-muted-foreground hover:text-foreground"
                >
                    <ArrowLeft className="size-3.5" />
                    HivePaaS
                </AppLink.Basic>
                <h3 className="text-base font-semibold">{data.data.component.name}</h3>
                <span className="text-[13px] text-muted-foreground">
                    {scan.scannedAt
                        ? `Last looked at ${timeAgoFormatter.format(scan.scannedAt)}`
                        : "Not scanned yet: the daily scan scans it, and Actions runs it now."}
                </span>
            </div>

            {scan.error && (
                <div className="flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/5 px-3 py-2 text-[13px] text-destructive">
                    <AlertTriangle className="mt-0.5 size-4 shrink-0" />
                    <span className="break-words">
                        {scan.images.length > 0 ? "Last scan incomplete" : "Not scanned"}: {scan.error}
                    </span>
                </div>
            )}

            {scan.images.length > 0 && (
                <>
                    <SeverityCounters summary={scan.summary} />
                    <ScannedImages images={scan.images} />
                    <FindingsTable
                        findings={scan.findings}
                        onIgnore={canWrite ? setIgnoring : undefined}
                    />
                    <IgnoredFindings
                        ignored={scan.ignored}
                        onRemove={
                            canWrite
                                ? ignore => {
                                      deleteIgnore({ ...NO_APP, scope: ignore.scope, ignoreID: ignore.id });
                                  }
                                : undefined
                        }
                    />
                </>
            )}

            <IgnoreFindingDialog
                finding={ignoring}
                variant="hivepaas"
                isPending={isIgnoring}
                onOpenChange={open => {
                    if (!open) {
                        setIgnoring(null);
                    }
                }}
                onConfirm={(scope, payload) => {
                    createIgnore({ ...NO_APP, scope, payload });
                }}
            />
        </div>
    );
}

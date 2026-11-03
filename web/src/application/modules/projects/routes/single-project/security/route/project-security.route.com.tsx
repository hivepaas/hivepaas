import { dashedBorderBox, inlineLink, listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { AlertTriangle, ShieldCheck } from "lucide-react";
import { Link, useParams } from "react-router";
import invariant from "tiny-invariant";
import { AppImageScanQueries } from "~/projects/data";
import { ScannedAppsTable } from "~/projects/module-shared/components";

import { AppLoader, BetaBadge } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

/**
 * What the images of the project's apps carry, the worst first: those of the
 * envs and the apps the person may read. Each app opens its Security tab.
 */
export function ProjectSecurityRoute() {
    const { id: projectID } = useParams<{ id: string }>();
    invariant(projectID, "projectID must be defined");
    const { canRead: canReadSystem } = useConditionalModule({ id: MODULE_IDS.System });
    const { data, isLoading } = AppImageScanQueries.useFindProject({ projectID });

    if (isLoading || !data) {
        return <AppLoader />;
    }
    const { enabled, apps } = data.data;

    return (
        <div
            className={cn(listBox, "flex flex-col gap-5")}
            data-testid="project-image-scan"
        >
            <div className="flex flex-col gap-0.5">
                <h3 className="flex items-center gap-2 text-base font-semibold">
                    <ShieldCheck className="size-4.5 text-muted-foreground" />
                    Image vulnerabilities
                    <BetaBadge />
                </h3>
                <span className="text-[13px] text-muted-foreground">
                    The project&apos;s apps, the worst first. Counts are fixable / all, after ignores.
                </span>
            </div>

            {!enabled && (
                <div className={cn(dashedBorderBox, "flex items-start gap-2")}>
                    <AlertTriangle className="mt-0.5 size-4 shrink-0 text-amber-600 dark:text-amber-500" />
                    <span>
                        Image scanning is switched off: nothing is scanned
                        {apps.length > 0 ? ", and what is shown is from the last scans" : ""}.{" "}
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

            {apps.length === 0 ? (
                <div className="rounded-md border px-4 py-8 text-center text-sm text-muted-foreground">
                    No app of this project has been scanned yet. An app is scanned after each deployment, every day, and
                    when asked on its Security tab.
                </div>
            ) : (
                <ScannedAppsTable items={apps} />
            )}
        </div>
    );
}

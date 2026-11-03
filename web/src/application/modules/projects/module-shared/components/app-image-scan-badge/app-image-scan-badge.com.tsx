import { memo } from "react";

import { Badge } from "@components/ui/badge";
import { ShieldAlert } from "lucide-react";
import { AppImageScanQueries } from "~/projects/data";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

// The header asks again now and then: a scan after a deployment ends after it.
const IMAGE_SCAN_REFETCH_INTERVAL_MS = 60_000;

/**
 * What the app's images carry, beside its status in the header: red with the
 * count of fixable CRITICAL vulnerabilities, amber with fixable HIGH ones when
 * there is no CRITICAL, grey when its scan failed; nothing otherwise, nor while
 * image scanning is off. It opens the Security tab.
 */
function View({ projectId, env, appId }: Props) {
    const { data } = AppImageScanQueries.useFindOne(
        { projectID: projectId, env, appID: appId },
        { refetchInterval: IMAGE_SCAN_REFETCH_INTERVAL_MS, retry: false },
    );
    const result = data?.data;
    if (!result?.enabled) {
        return null;
    }

    const { critical, high } = result.summary;
    const failed = result.error !== "" || result.images.some(image => image.status === "failed");
    let badge: { tone: "red" | "amber" | "neutral"; label: string; title: string };
    if (critical.fixable > 0) {
        badge = {
            tone: "red",
            label: `${critical.fixable} critical`,
            title: `${critical.fixable} fixable CRITICAL vulnerabilities in the app's images`,
        };
    } else if (high.fixable > 0) {
        badge = {
            tone: "amber",
            label: `${high.fixable} high`,
            title: `${high.fixable} fixable HIGH vulnerabilities in the app's images`,
        };
    } else if (failed) {
        badge = { tone: "neutral", label: "Scan failed", title: "The app's images could not be scanned" };
    } else {
        return null;
    }

    return (
        <AppLink.Basic
            to={ROUTE.projects.single.apps.single.security.$route(projectId, env, appId)}
            title={`${badge.title}: open the Security tab`}
            className="rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            data-testid="app-image-scan-badge"
        >
            <Badge
                tone={badge.tone}
                className="h-6 gap-1 hover:opacity-90"
            >
                <ShieldAlert
                    className="size-3"
                    aria-hidden
                />
                {badge.label}
            </Badge>
        </AppLink.Basic>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
}

export const AppImageScanBadge = memo(View);

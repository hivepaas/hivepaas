import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge, statusClassName, statusDotClassName } from "@components/ui/badge";
import { Clock, Loader2 } from "lucide-react";
import type { AppActiveDeployment } from "~/projects/domain";
import { EAppDeploymentStatus } from "~/projects/module-shared/enums";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

// What the deployment is doing, in the words and colors of the deployments
// list: Deploying (progress, purple) while it runs, Queued (waiting, blue)
// while it waits for one before it.
function describe(deployment: AppActiveDeployment) {
    const isRunning = deployment.status === EAppDeploymentStatus.InProgress;
    return {
        isRunning,
        label: isRunning ? "Deploying" : "Queued",
        meaning: isRunning ? ("progress" as const) : ("waiting" as const),
    };
}

/**
 * The app's deployment that has not ended, beside its status in the header. It
 * opens that deployment, and its log.
 */
function BadgeView({ projectId, env, appId, deployment }: BadgeProps) {
    const { isRunning, label, meaning } = describe(deployment);
    const Icon = isRunning ? Loader2 : Clock;

    return (
        <AppLink.Basic
            to={ROUTE.projects.single.apps.single.deployments.details.$route(projectId, env, appId, deployment.id)}
            title={`${label}: open the deployment`}
            className="rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
            <Badge className={cn(statusClassName(meaning), "h-6 gap-1 hover:opacity-90")}>
                <Icon
                    className={cn("size-3", isRunning && "animate-spin")}
                    aria-hidden
                />
                {label}
            </Badge>
        </AppLink.Basic>
    );
}

/** The Deployments tab's mark while a deployment of the app has not ended. */
function DotView({ deployment }: DotProps) {
    const { isRunning, label, meaning } = describe(deployment);

    return (
        <span
            className={cn("size-2 rounded-full", statusDotClassName(meaning), isRunning && "animate-pulse")}
            role="img"
            aria-label={label}
            title={label}
        />
    );
}

interface BadgeProps {
    projectId: string;
    env: string;
    appId: string;
    deployment: AppActiveDeployment;
}

interface DotProps {
    deployment: AppActiveDeployment;
}

export const AppActiveDeploymentBadge = memo(BadgeView);
export const AppActiveDeploymentDot = memo(DotView);

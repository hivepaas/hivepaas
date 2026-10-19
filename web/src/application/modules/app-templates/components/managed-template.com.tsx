import { DatabaseBackup, ExternalLink, Lock } from "lucide-react";
import { managedEngineOf } from "~/projects/module-shared/utils";

import { LICENSE_DOCS } from "@application/shared/constants";
import { useReadableLicense } from "@application/shared/hooks";
import { licenseFeatureCopy } from "@application/shared/utils";

import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

/** A managed template's mark: HivePaaS goes on looking after what it creates. */
export function ManagedTemplateBadge({ engine }: { engine: string }) {
    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <Badge
                    variant="secondary"
                    className="shrink-0 text-[11px] font-medium bg-violet-500/10 text-violet-700 dark:text-violet-300 border border-violet-500/20"
                >
                    <DatabaseBackup className="size-3 mr-1 inline" />
                    Managed
                </Badge>
            </TooltipTrigger>
            <TooltipContent className="max-w-[260px]">
                HivePaaS looks after this {managedEngineOf(engine).name}: it archives every change, takes nightly
                backups, tests that they restore, and restores it to any moment into a new app.
            </TooltipContent>
        </Tooltip>
    );
}

/** The features a template needs that the plan does not include, on a card: one line. */
export function MissingFeaturesLine({ features }: { features: string[] }) {
    if (features.length === 0) {
        return null;
    }
    return (
        <div className="mt-3 flex items-center gap-1.5 rounded-md bg-violet-500/10 px-2 py-1 text-[12px] text-violet-700 dark:text-violet-300 border border-violet-500/20">
            <Lock className="size-3.5 shrink-0" />
            <span className="truncate">Needs a paid plan</span>
        </div>
    );
}

/**
 * The features a template needs that the plan does not include, in full: what they are, the plans that grant them,
 * and where to get one - in place of a form the server would refuse.
 */
export function MissingFeaturesNotice({ features }: { features: string[] }) {
    const { canRead, license } = useReadableLicense({ enabled: features.length > 0 });
    if (features.length === 0) {
        return null;
    }
    const pricingUrl = license?.portal?.pricingUrl ?? "";
    return (
        <div
            role="note"
            className="rounded-xl border border-violet-500/30 bg-violet-500/5 p-4 text-sm"
        >
            <div className="flex items-start gap-3">
                <Lock className="mt-0.5 size-4 shrink-0 text-violet-700 dark:text-violet-300" />
                <div className="flex flex-col gap-1.5">
                    {features.map(feature => {
                        const copy = licenseFeatureCopy(feature);
                        return (
                            <p key={feature}>
                                This installation&apos;s plan does not include {copy.name}. They come with {copy.plans}.
                            </p>
                        );
                    })}
                    {!canRead && (
                        <p className="text-muted-foreground">
                            Ask an administrator to upgrade this installation&apos;s plan.
                        </p>
                    )}
                    <div className="flex flex-wrap gap-4">
                        {canRead && pricingUrl !== "" && (
                            <a
                                href={pricingUrl}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="inline-flex items-center gap-1 font-medium underline-offset-4 hover:underline"
                            >
                                Upgrade
                                <ExternalLink className="size-3.5 shrink-0" />
                            </a>
                        )}
                        <a
                            href={LICENSE_DOCS.plansAndLimits}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="inline-flex items-center gap-1 underline-offset-4 hover:underline"
                        >
                            Plans and limits
                            <ExternalLink className="size-3.5 shrink-0" />
                        </a>
                    </div>
                </div>
            </div>
        </div>
    );
}

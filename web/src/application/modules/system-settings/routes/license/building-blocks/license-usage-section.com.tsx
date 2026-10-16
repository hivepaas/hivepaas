import { type ReactNode, useId } from "react";

import { cn } from "@lib/utils";
import { SectionHeader } from "~/system-settings/module-shared";

import { InfoBlock } from "@application/shared/components";
import type { License, LicenseLimitUsage } from "@application/shared/entities";
import { ELicenseLimit } from "@application/shared/enums";
import { licenseLimitCopy, licenseLimitNoun, licensePlanName } from "@application/shared/utils";

/**
 * How many there are against the most the plan allows, as a bar: the primary
 * color below it, amber at it - none more can be added - and red past it, as
 * after a key ran out. The bar never runs past full; the numbers say by how much.
 */
function LicenseUsageMeter({ usage, max }: { usage: LicenseLimitUsage; max: number }) {
    const { label } = licenseLimitCopy(usage.limit);
    const { used } = usage;
    const text = `${used} of ${max} ${licenseLimitNoun(usage.limit, max)}`;
    const isOver = used > max;
    const isFull = used === max;
    const percent = Math.min(100, (used / max) * 100);

    return (
        <div className="flex flex-col gap-1.5">
            <div className="flex flex-wrap items-baseline justify-between gap-x-3 text-sm">
                <span className="tabular-nums">{text}</span>
                {isOver && (
                    <span className="text-xs text-red-700 dark:text-red-400">Over the limit: none can be added</span>
                )}
                {isFull && <span className="text-xs text-amber-700 dark:text-amber-400">Full</span>}
            </div>
            <div
                role="meter"
                aria-label={label}
                aria-valuemin={0}
                aria-valuemax={max}
                // Within the range, as a meter's value must be; the text says how far over.
                aria-valuenow={Math.min(used, max)}
                aria-valuetext={text}
                className="h-1.5 w-full overflow-hidden rounded-full bg-muted"
            >
                <div
                    className={cn(
                        "h-full transition-[width]",
                        isOver ? "bg-red-500" : isFull ? "bg-amber-500" : "bg-primary",
                    )}
                    style={{ width: `${percent}%` }}
                />
            </div>
        </div>
    );
}

/** The meter, or what the plan says of the kind when there is nothing to measure against. */
function usageValue(usage: LicenseLimitUsage, plan: string): ReactNode {
    const { used, max } = usage;

    if (max === null) {
        return (
            <span className="text-sm tabular-nums">
                {used} {licenseLimitNoun(usage.limit, used)} · no limit
            </span>
        );
    }

    // Nothing to fill: the plan does not include the kind at all.
    if (max === 0) {
        return (
            <span className="text-sm">
                Not included on {plan}
                {used > 0 && ` · ${used} already there`}
            </span>
        );
    }

    return (
        <LicenseUsageMeter
            usage={usage}
            max={max}
        />
    );
}

/** One limited kind, with what counts towards it. */
function LicenseUsageRow({ usage, plan }: { usage: LicenseLimitUsage; plan: string }) {
    const copy = licenseLimitCopy(usage.limit);

    return (
        <InfoBlock
            titleWidth={220}
            title={copy.label}
        >
            <div className="flex max-w-[640px] flex-col gap-1.5">
                {usageValue(usage, plan)}
                {copy.counts && <p className="text-xs text-muted-foreground">{copy.counts}</p>}
            </div>
        </InfoBlock>
    );
}

/**
 * Google, Microsoft and OpenID Connect sign-in is not a count: Bee includes none
 * (max 0) and every paid plan all (max null). On Bee the ones set up before keep
 * signing people in; none can be added or enabled.
 */
function SignInProvidersRow({ usage, plan }: { usage: LicenseLimitUsage; plan: string }) {
    const { used } = usage;

    return (
        <InfoBlock
            titleWidth={220}
            title={licenseLimitCopy(usage.limit).label}
        >
            <span className="text-sm">
                {usage.max === null
                    ? "Included"
                    : `Not included on ${plan}${used > 0 ? ` · ${used} already set up ${used === 1 ? "keeps" : "keep"} signing people in` : ""}`}
            </span>
        </InfoBlock>
    );
}

interface Props {
    license: License;
}

/** How much of each limited kind there is, against what the plan in force allows. */
export function LicenseUsageSection({ license }: Props) {
    const headingId = useId();
    const plan = licensePlanName(license.plan);

    return (
        <section
            aria-labelledby={headingId}
            className="flex flex-col gap-4"
        >
            <SectionHeader>
                <h2 id={headingId}>Usage</h2>
            </SectionHeader>
            <div className="flex flex-col gap-4 px-3">
                {license.limits.map(usage =>
                    // Worded on its own while it is all or none; a number for it
                    // would be measured like any other kind.
                    usage.limit === ELicenseLimit.SignInProviders && (usage.max === null || usage.max === 0) ? (
                        <SignInProvidersRow
                            key={usage.limit}
                            usage={usage}
                            plan={plan}
                        />
                    ) : (
                        <LicenseUsageRow
                            key={usage.limit}
                            usage={usage}
                            plan={plan}
                        />
                    ),
                )}
            </div>
        </section>
    );
}

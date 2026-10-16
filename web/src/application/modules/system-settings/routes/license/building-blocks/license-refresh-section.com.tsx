import { useId } from "react";

import { toast } from "sonner";
import { SectionHeader } from "~/system-settings/module-shared";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { LicenseCommands } from "@application/shared/data/commands";
import type { License } from "@application/shared/entities";
import { ELicenseRefreshResult, ELicenseStatus } from "@application/shared/enums";
import { PermissionTooltipAction } from "@application/shared/permissions";
import { formatLicenseDateTime } from "@application/shared/utils";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Button, Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui";

import { licenseInForceSummary, refreshResultText } from "./license-labels";

/**
 * Says what a check now came to. One that failed never gets here: the API hook
 * toasts the server's reason, such as "The license portal could not be reached".
 */
function announce(result: string, license: License) {
    switch (result) {
        case ELicenseRefreshResult.Newer:
            toast.success(`A newer key is installed: ${licenseInForceSummary(license)}`);
            return;
        case ELicenseRefreshResult.Moved:
            toast.warning(refreshResultText(result));
            return;
        case ELicenseRefreshResult.Current:
            toast.success(refreshResultText(result));
            return;
        default:
            toast.info(refreshResultText(result));
    }
}

interface Props {
    license: License;
}

/**
 * The daily check with the portal for a newer key - a renewal, a plan change, a
 * paid key in place of a trial - how the last one went, and asking now. Asking
 * now asks whether or not the daily check is on.
 */
export function LicenseRefreshSection({ license }: Props) {
    const headingId = useId();
    const { refresh } = license;
    // The key is what the portal is asked with: with none, there is nothing to ask.
    const hasNoKey = license.status === ELicenseStatus.None;
    const lastCheckedAt = refresh?.lastCheckedAt ?? null;

    const { mutate: check, isPending } = LicenseCommands.useRefresh({
        onSuccess: response => {
            announce(response.data.result, response.data.license);
        },
    });

    return (
        <section
            aria-labelledby={headingId}
            className="flex flex-col gap-4"
        >
            <SectionHeader>
                <h2 id={headingId}>Online renewal</h2>
            </SectionHeader>
            <div className="flex flex-col gap-6 px-3 text-sm">
                {refresh && (
                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Daily check"
                                content="Once a day HivePaaS sends the key and this installation's ID to the license portal, and installs any newer key it answers with. It sends nothing else. HP_LICENSE_REFRESH_DISABLED switches it off."
                            />
                        }
                    >
                        <span>{refresh.enabled ? "On" : "Off"}</span>
                    </InfoBlock>
                )}
                <InfoBlock
                    titleWidth={220}
                    title="Last check"
                >
                    <div className="flex flex-col items-start gap-2">
                        {lastCheckedAt ? (
                            <span>
                                <time
                                    dateTime={lastCheckedAt.toISOString()}
                                    title={formatLicenseDateTime(lastCheckedAt)}
                                >
                                    {timeAgoFormatter.format(lastCheckedAt)}
                                </time>
                                {refresh?.lastResult && (
                                    <span className="text-muted-foreground">
                                        {" "}
                                        · {refreshResultText(refresh.lastResult)}
                                    </span>
                                )}
                            </span>
                        ) : (
                            <span className="text-muted-foreground">Never</span>
                        )}
                        {refresh?.lastError && (
                            <span className="whitespace-pre-wrap break-words text-amber-700 dark:text-amber-400">
                                {refresh.lastError}
                            </span>
                        )}
                        <PermissionTooltipAction
                            id={MODULE_IDS.System}
                            action="write"
                        >
                            {({ isDenied }) => {
                                const button = (
                                    <Button
                                        type="button"
                                        variant="outline"
                                        size="sm"
                                        disabled={isDenied || hasNoKey || isPending}
                                        isLoading={isPending}
                                        onClick={() => {
                                            check();
                                        }}
                                    >
                                        Check for a newer key
                                    </Button>
                                );

                                if (!hasNoKey || isDenied) {
                                    return button;
                                }

                                // A disabled button takes no pointer events, so the
                                // tooltip hangs on a wrapper, as for a denied action.
                                return (
                                    <Tooltip>
                                        <TooltipTrigger asChild>
                                            <span className="inline-flex">{button}</span>
                                        </TooltipTrigger>
                                        <TooltipContent side="top">
                                            No key to ask about yet: paste one first
                                        </TooltipContent>
                                    </Tooltip>
                                );
                            }}
                        </PermissionTooltipAction>
                    </div>
                </InfoBlock>
            </div>
        </section>
    );
}

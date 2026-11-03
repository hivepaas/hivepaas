import { Button } from "@components/ui";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogDescription,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { Separator } from "@components/ui/separator";
import { ExternalLinkIcon } from "lucide-react";

import { AppLink } from "@application/shared/components/navigation";
import { LICENSE_DOCS, ROUTE } from "@application/shared/constants";
import { useReadableLicense } from "@application/shared/hooks";

import { usePlanLimitReachedDialogState } from "../hooks";
import type { PlanLimitRefusal } from "../types";
import { planLimitReachedCopy, planLimitUsageLine } from "../utils";

/** Shown only while the dialog closes after destroy() cleared it. */
const NO_REFUSAL: PlanLimitRefusal = { limit: "", max: null, plan: "", message: "" };

/**
 * Says why a creation was refused: the plan in force allows no more of its kind.
 * It names the limit, says what counts towards it and what makes room, and leads
 * to the plan, and to the docs' page on plans for more than a dialog has room for.
 *
 * It is opened from useApiErrorNotifications rather than from the creations:
 * around a dozen of them - apps, functions, previews, clones, users, nodes,
 * sign-in providers - funnel their failures through that one hook, and the
 * refusal reads the same wherever it comes from. A creation that does not go
 * through it calls openPlanLimitReachedDialog.
 *
 * Only someone who can read System is shown the plan, and offered to change it:
 * anyone else is asked to go to an administrator.
 */
export function PlanLimitReachedDialog() {
    const {
        mode,
        props: { refusal = NO_REFUSAL },
        close,
    } = usePlanLimitReachedDialogState();

    const open = mode === "open";
    // Fetched each time the dialog opens, so the usage it shows counts what is
    // there now; the banners, which read the same query, catch up with it too.
    const { canRead, license } = useReadableLicense({ enabled: open });

    const { title, body } = planLimitReachedCopy(refusal);
    const usageLine = planLimitUsageLine(refusal, license);
    const pricingUrl = license?.portal?.pricingUrl ?? "";

    return (
        <Dialog
            open={open}
            onOpenChange={nextOpen => {
                if (!nextOpen) {
                    close();
                }
            }}
        >
            <DialogFixedContent className="sm:max-w-[560px]">
                <DialogHeader>
                    <DialogTitle className="pr-6 leading-snug">{title}</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>

                <DialogBody className="flex flex-col gap-3">
                    <DialogDescription className="text-sm leading-6 text-foreground">{body}</DialogDescription>

                    {usageLine !== undefined && <p className="text-sm text-muted-foreground">{usageLine}</p>}

                    {!canRead && (
                        <p className="text-sm text-muted-foreground">
                            Ask an administrator to upgrade this installation&apos;s plan.
                        </p>
                    )}

                    <a
                        href={LICENSE_DOCS.plansAndLimits}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex items-center gap-1 self-start text-sm underline-offset-4 hover:underline"
                    >
                        Plans and limits
                        <ExternalLinkIcon className="size-3.5 shrink-0" />
                    </a>
                </DialogBody>

                <DialogActionFooter className="flex flex-wrap justify-end">
                    <Button
                        type="button"
                        variant="outline"
                        onClick={close}
                    >
                        Close
                    </Button>

                    {/* Closes the dialog itself rather than leave it to the route change:
                        the destroy() that comes with it would empty the dialog while it
                        fades out. */}
                    {canRead && (
                        <Button
                            asChild
                            variant="outline"
                            className="min-w-[100px]"
                        >
                            <AppLink.Basic
                                to={ROUTE.systemSettings.license.$route}
                                onClick={close}
                            >
                                View license
                            </AppLink.Basic>
                        </Button>
                    )}

                    {pricingUrl !== "" && (
                        <Button
                            asChild
                            className="min-w-[100px]"
                        >
                            <a
                                href={pricingUrl}
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                Upgrade
                                <ExternalLinkIcon />
                            </a>
                        </Button>
                    )}
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}

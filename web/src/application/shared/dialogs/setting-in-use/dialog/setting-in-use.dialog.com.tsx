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
import { useQuery } from "@tanstack/react-query";
import { AlertTriangle, ExternalLinkIcon, Loader2Icon } from "lucide-react";

import { useSettingUsageApi } from "@application/shared/api/hooks";
import { type SettingUsage, settingUsageHref, settingUsageLabel } from "@application/shared/utils";

import { useSettingInUseDialogState } from "../hooks";

/**
 * Says what is still using a setting the operator just tried to delete.
 *
 * It is opened from useApiErrorNotifications rather than from any delete call
 * site. There are around fifty of those, they all funnel their failures through
 * that one hook, and the refusal is the same wherever it comes from - so handling
 * it there means every setting type gets this without touching any of them.
 *
 * The list is fetched rather than read out of the error. The delete answers only
 * whether it happened; describing the blockers is a list resource of its own,
 * which is why the failed request's URL is carried on the exception - the usages
 * of a setting hang off the same path the delete was sent to.
 */
export function SettingInUseDialog() {
    const { mode, props: { settingName = "", requestUrl = "" } = {}, close } = useSettingInUseDialogState();
    const { queries } = useSettingUsageApi();

    const open = mode === "open";
    const usages = useQuery({
        queryKey: ["setting-usages", requestUrl],
        queryFn: async ({ signal }) => {
            const response = await queries.findMany({ deleteUrl: requestUrl }, signal);
            return response.data;
        },
        enabled: open && requestUrl !== "",
        retry: false,
        gcTime: 0,
    });

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
                    <DialogTitle>
                        {settingName === "" ? "This setting is still in use" : `'${settingName}' is still in use`}
                    </DialogTitle>
                </DialogHeader>
                <DialogDescription className="sr-only">
                    The objects that still reference this setting, and where to edit them.
                </DialogDescription>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>

                <DialogBody className="flex flex-col gap-4">
                    <div className="flex items-center gap-2.5 rounded-md border border-destructive bg-destructive/10 px-3.5 py-2.5 text-sm text-destructive font-medium">
                        <AlertTriangle className="size-4 shrink-0 text-destructive" />
                        <span>It cannot be deleted while anything still points at it.</span>
                    </div>

                    <p className="text-sm leading-6 text-foreground">
                        Deleting it would leave these pointing at something that no longer exists, and each of them
                        would fail the next time it is applied. Remove the reference from them first.
                    </p>

                    {usages.isPending && (
                        <p className="flex items-center gap-2 text-sm text-muted-foreground">
                            <Loader2Icon className="size-3.5 shrink-0 animate-spin" />
                            Looking up what uses it...
                        </p>
                    )}

                    {usages.isError && (
                        <p className="text-sm text-muted-foreground">The list of what uses it could not be loaded.</p>
                    )}

                    {usages.data != null && usages.data.length > 0 && (
                        <ul className="flex max-h-[280px] flex-col gap-1.5 overflow-y-auto rounded-md border bg-muted/30 px-3.5 py-2.5">
                            {usages.data.map((usage: SettingUsage) => (
                                <li
                                    key={`${usage.type}:${usage.id}`}
                                    className="text-sm leading-6"
                                >
                                    <SettingUsageRow usage={usage} />
                                </li>
                            ))}
                        </ul>
                    )}
                </DialogBody>

                <DialogActionFooter className="flex justify-end">
                    <Button
                        type="button"
                        variant="outline"
                        onClick={close}
                        className="min-w-[120px]"
                    >
                        Close
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}

function SettingUsageRow({ usage }: { usage: SettingUsage }) {
    const href = settingUsageHref(usage);
    const label = settingUsageLabel(usage);

    // No link when the setting type is one this does not know how to place. The
    // row still appears: naming the thing is most of the value, and a guessed URL
    // that lands on a 404 is worse than none.
    if (href == null) {
        return (
            <span className="text-muted-foreground">
                {label}
                {usage.settingType != null && usage.settingType !== "" && (
                    <span className="ml-2 font-mono text-xs opacity-70">{usage.settingType}</span>
                )}
            </span>
        );
    }

    return (
        <a
            href={href}
            target="_blank"
            rel="noopener noreferrer"
            className="text-link inline-flex items-center gap-1 font-medium underline underline-offset-4 hover:opacity-80"
        >
            {label}
            <ExternalLinkIcon className="size-3 shrink-0" />
        </a>
    );
}

import { useEffect, useState } from "react";

import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { Textarea } from "@components/ui/textarea";
import type { VulnFinding, VulnIgnorePayload, VulnIgnoreScope } from "~/projects/api/services";

import { InfoBlock } from "@application/shared/components";

import {
    Button,
    Checkbox,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
    Separator,
} from "@/components/ui";

import { SCOPE_LABELS } from "./severity";

const SCOPES: VulnIgnoreScope[] = ["app", "project-env", "project", "global"];

/** How long an ignore lasts: at most a year, so each is looked at again. */
const EXPIRIES = [
    { value: "30", label: "30 days" },
    { value: "90", label: "90 days" },
    { value: "180", label: "180 days" },
    { value: "365", label: "1 year" },
    { value: "never", label: "Never" },
] as const;

const REASON_MAX_LENGTH = 1000;

interface Props {
    finding: VulnFinding | null;
    isPending: boolean;
    onOpenChange: (open: boolean) => void;
    onConfirm: (scope: VulnIgnoreScope, payload: VulnIgnorePayload) => void;
}

/**
 * Sets a finding aside, for the apps of a scope: a team found it does not
 * apply to them. It needs a reason, and lasts until it expires.
 */
export function IgnoreFindingDialog({ finding, isPending, onOpenChange, onConfirm }: Props) {
    const [scope, setScope] = useState<VulnIgnoreScope>("app");
    const [packageOnly, setPackageOnly] = useState(true);
    const [reason, setReason] = useState("");
    const [expiry, setExpiry] = useState<string>("90");

    useEffect(() => {
        if (finding) {
            setScope("app");
            setPackageOnly(true);
            setReason("");
            setExpiry("90");
        }
    }, [finding]);

    if (!finding) {
        return null;
    }

    const trimmed = reason.trim();

    return (
        <Dialog
            open
            onOpenChange={open => {
                if (!isPending) {
                    onOpenChange(open);
                }
            }}
        >
            <DialogFixedContent className="sm:max-w-[620px]">
                <DialogHeader>
                    <DialogTitle>Ignore {finding.vulnId}</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                <DialogBody className="flex flex-col gap-4">
                    <p className="text-sm leading-6 text-muted-foreground">
                        An ignored vulnerability is left out of the counts, the app&apos;s badge and the daily report,
                        for every app of the scope chosen. It is still listed under Ignored.
                    </p>
                    <InfoBlock
                        titleWidth={140}
                        title="For"
                    >
                        <Select
                            value={scope}
                            onValueChange={value => {
                                setScope(value as VulnIgnoreScope);
                            }}
                        >
                            <SelectTrigger
                                className="w-full max-w-[320px]"
                                aria-label="Scope"
                            >
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                {SCOPES.map(value => (
                                    <SelectItem
                                        key={value}
                                        value={value}
                                    >
                                        {SCOPE_LABELS[value]}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </InfoBlock>
                    <InfoBlock
                        titleWidth={140}
                        title="Package"
                    >
                        <label
                            htmlFor="ignore-package-only"
                            className="flex cursor-pointer select-none items-center gap-2 text-sm"
                        >
                            <Checkbox
                                id="ignore-package-only"
                                checked={packageOnly}
                                onCheckedChange={checked => {
                                    setPackageOnly(checked === true);
                                }}
                            />
                            <span>
                                Only in <span className="font-mono">{finding.pkgName}</span>
                            </span>
                        </label>
                    </InfoBlock>
                    <InfoBlock
                        titleWidth={140}
                        title="Reason"
                    >
                        <Textarea
                            value={reason}
                            maxLength={REASON_MAX_LENGTH}
                            rows={3}
                            placeholder="Why it does not apply: the code that has it never runs, the input never reaches it..."
                            aria-label="Reason"
                            onChange={event => {
                                setReason(event.target.value);
                            }}
                        />
                    </InfoBlock>
                    <InfoBlock
                        titleWidth={140}
                        title="Expires after"
                    >
                        <Select
                            value={expiry}
                            onValueChange={setExpiry}
                        >
                            <SelectTrigger
                                className="w-full max-w-[200px]"
                                aria-label="Expires after"
                            >
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                {EXPIRIES.map(item => (
                                    <SelectItem
                                        key={item.value}
                                        value={item.value}
                                    >
                                        {item.label}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </InfoBlock>
                </DialogBody>
                <DialogActionFooter>
                    <Button
                        variant="outline"
                        disabled={isPending}
                        onClick={() => {
                            onOpenChange(false);
                        }}
                    >
                        Cancel
                    </Button>
                    <Button
                        disabled={trimmed === "" || isPending}
                        isLoading={isPending}
                        onClick={() => {
                            onConfirm(scope, {
                                vulnId: finding.vulnId,
                                package: packageOnly ? finding.pkgName : "",
                                reason: trimmed,
                                expireAt:
                                    expiry === "never"
                                        ? null
                                        : new Date(Date.now() + Number(expiry) * 24 * 60 * 60 * 1000),
                                inheritable: true,
                            });
                        }}
                    >
                        Ignore
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}

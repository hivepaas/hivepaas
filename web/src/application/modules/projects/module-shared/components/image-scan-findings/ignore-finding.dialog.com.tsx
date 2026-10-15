import { useEffect, useMemo, useState } from "react";

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

/** Whom an ignore is for: a scope, and for a global one whether the apps have it. */
interface ScopeChoice {
    value: string;
    label: string;
    scope: VulnIgnoreScope;
    inheritable: boolean;
}

type ScopeChoices = readonly [ScopeChoice, ...ScopeChoice[]];

function appChoice(scope: VulnIgnoreScope): ScopeChoice {
    return { value: scope, label: SCOPE_LABELS[scope], scope, inheritable: true };
}

/** An app's: its own scope or one it inherits from. */
const APP_CHOICES: ScopeChoices = [
    appChoice("app"),
    appChoice("project-env"),
    appChoice("project"),
    appChoice("global"),
];

/** HivePaaS's own parts are of no project: a global ignore, for them alone or for every app too. */
const HIVEPAAS_CHOICES: ScopeChoices = [
    { value: "hivepaas", label: "HivePaaS only", scope: "global", inheritable: false },
    { value: "everywhere", label: "HivePaaS and every project", scope: "global", inheritable: true },
];

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
    /** hivepaas sets aside a finding of HivePaaS's own images: a global ignore. */
    variant?: "app" | "hivepaas";
    /** The scopes the user may set an ignore at; every one of the variant's when left out. */
    scopes?: readonly VulnIgnoreScope[];
    isPending: boolean;
    onOpenChange: (open: boolean) => void;
    onConfirm: (scope: VulnIgnoreScope, payload: VulnIgnorePayload) => void;
}

/**
 * Sets a finding aside, for the apps of a scope: a team found it does not
 * apply to them. It needs a reason, and lasts until it expires.
 */
export function IgnoreFindingDialog({ finding, variant = "app", scopes, isPending, onOpenChange, onConfirm }: Props) {
    // Only what the server would take from the user: a global ignore is the
    // Settings module's, the others the project's.
    const scopesKey = scopes?.join(",");
    const choices = useMemo((): readonly ScopeChoice[] => {
        const all = variant === "hivepaas" ? HIVEPAAS_CHOICES : APP_CHOICES;
        const allowed = scopesKey?.split(",");
        return allowed ? all.filter(item => allowed.includes(item.scope)) : all;
    }, [variant, scopesKey]);
    const firstValue = choices[0]?.value ?? "";
    const [choiceValue, setChoiceValue] = useState(firstValue);
    const [packageOnly, setPackageOnly] = useState(true);
    const [reason, setReason] = useState("");
    const [expiry, setExpiry] = useState<string>("90");

    useEffect(() => {
        if (finding) {
            setChoiceValue(firstValue);
            setPackageOnly(true);
            setReason("");
            setExpiry("90");
        }
    }, [finding, firstValue]);

    // None at all: the user may set no ignore, and the dialog is not offered.
    const choice = choices.find(item => item.value === choiceValue) ?? choices[0];
    if (!finding || !choice) {
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
                        {variant === "hivepaas"
                            ? "An ignored vulnerability is left out of the counts of HivePaaS's own images - and of every app's too, when it is for every project. It is still listed under Ignored."
                            : "An ignored vulnerability is left out of the counts, the app's badge and the daily report, for every app of the scope chosen. It is still listed under Ignored."}
                    </p>
                    <InfoBlock
                        titleWidth={140}
                        title="For"
                    >
                        <Select
                            value={choice.value}
                            onValueChange={setChoiceValue}
                        >
                            <SelectTrigger
                                className="w-full max-w-[320px]"
                                aria-label="Scope"
                            >
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                {choices.map(item => (
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
                            onConfirm(choice.scope, {
                                vulnId: finding.vulnId,
                                package: packageOnly ? finding.pkgName : "",
                                reason: trimmed,
                                expireAt:
                                    expiry === "never"
                                        ? null
                                        : new Date(Date.now() + Number(expiry) * 24 * 60 * 60 * 1000),
                                inheritable: choice.inheritable,
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

import { useEffect, useMemo, useRef, useState } from "react";

import {
    Button,
    Checkbox,
    Field,
    FieldGroup,
    Input,
    Label,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@components/ui";
import { PasswordInput } from "@components/ui/input-password";
import { FileUpIcon, OctagonXIcon } from "lucide-react";
import { specImportErrorCode } from "~/operations/api/hooks";
import { SpecImportCommands } from "~/operations/data";
import type {
    SpecImportExisting,
    SpecImportOptions,
    SpecImportPlan,
    SpecImportResult,
    SpecImportScope,
    SpecImportSelection,
} from "~/operations/domain";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { type ModuleId, PermissionTooltipAction } from "@application/shared/permissions";

import { SpecImportPlanTree } from "./spec-import-plan-tree.com";
import { SpecImportResultPanel } from "./spec-import-result.com";
import { buildImportTree, leafPaths, selectionOf } from "./spec-import.tree";

const EXISTING_OPTIONS: { value: SpecImportExisting; label: string; hint: string }[] = [
    {
        value: "update",
        label: "Make it match the bundle",
        hint: "What this installation already has is changed to what the bundle says. A running app whose configuration changes restarts.",
    },
    {
        value: "keep",
        label: "Leave it as it is",
        hint: "Only what is missing is created. Nothing that exists changes, and nothing restarts.",
    },
];

const DEFAULT_OPTIONS: SpecImportOptions = { existing: "update", deployCreated: true, deployChangedSource: false };

/** How long the plan waits for the operator to stop clicking before it is asked for again. */
const REVALIDATE_DELAY_MS = 400;

const EVERYTHING: SpecImportSelection = { include: [], exclude: [] };

function formatDate(date: Date): string {
    return date.toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" });
}

export function SpecImportPanel({ scope, scopeLabel, permissionModuleId = MODULE_IDS.System }: Props) {
    const fileInputRef = useRef<HTMLInputElement>(null);
    const [file, setFile] = useState<File | undefined>();
    const [passphrase, setPassphrase] = useState("");
    const [submittedPassphrase, setSubmittedPassphrase] = useState<string | undefined>();
    const [needsPassphrase, setNeedsPassphrase] = useState(false);
    const [options, setOptions] = useState<SpecImportOptions>(DEFAULT_OPTIONS);
    const [checked, setChecked] = useState<Set<string> | undefined>();
    const [plan, setPlan] = useState<SpecImportPlan | undefined>();
    const [planKey, setPlanKey] = useState<string | undefined>();
    const [validateError, setValidateError] = useState<Error | undefined>();
    const [planChanged, setPlanChanged] = useState(false);
    const [result, setResult] = useState<SpecImportResult | undefined>();
    const [revision, setRevision] = useState(0);

    const roots = useMemo(() => (plan ? buildImportTree(plan.nodes) : []), [plan]);
    const selection = useMemo(() => (checked ? selectionOf(roots, checked) : EVERYTHING), [roots, checked]);
    // What a plan was made for. A plan answers one selection, one set of options,
    // one passphrase; importing it for another would be refused anyway.
    const requestKey = JSON.stringify({ selection, options, submittedPassphrase });
    const selectionRef = useRef(selection);
    selectionRef.current = selection;

    const { mutateAsync: validate, isPending: isValidating } = SpecImportCommands.useValidateImport();
    const { mutate: apply, isPending: isApplying } = SpecImportCommands.useApplyImport({
        onSuccess: response => {
            setResult(response.data);
        },
        onError: error => {
            if (specImportErrorCode(error) === "ERR_SPEC_IMPORT_PLAN_CHANGED") {
                setPlanChanged(true);
                askAgain();
            }
        },
    });

    // The plan is asked for again whenever what it answers changes, once the
    // operator stops changing it. Only the latest answer is kept.
    const latest = useRef(0);
    useEffect(() => {
        if (!file || result || (needsPassphrase && submittedPassphrase === undefined)) {
            return;
        }
        const { current } = selectionRef;
        if (!current) {
            return;
        }
        const requestID = ++latest.current;
        const key = requestKey;
        const timer = window.setTimeout(() => {
            validate({ scope, bundle: file, passphrase: submittedPassphrase, selection: current, options })
                .then(response => {
                    if (requestID !== latest.current) {
                        return;
                    }
                    setPlan(response.data);
                    setPlanKey(key);
                    setValidateError(undefined);
                    setChecked(previous => previous ?? new Set(leafPaths(buildImportTree(response.data.nodes))));
                })
                .catch((error: unknown) => {
                    if (requestID !== latest.current) {
                        return;
                    }
                    const code = specImportErrorCode(error);
                    if (code === "ERR_SPEC_PASSPHRASE_REQUIRED" || code === "ERR_SPEC_PASSPHRASE_INVALID") {
                        setNeedsPassphrase(true);
                        if (code === "ERR_SPEC_PASSPHRASE_REQUIRED") {
                            setSubmittedPassphrase(undefined);
                            return;
                        }
                    }
                    setValidateError(error instanceof Error ? error : new Error(String(error)));
                });
        }, REVALIDATE_DELAY_MS);

        return () => {
            window.clearTimeout(timer);
        };
    }, [file, result, needsPassphrase, submittedPassphrase, requestKey, revision, scope, options, validate]);

    // The plan shown no longer answers what is asked until the next one arrives,
    // so importing waits for it.
    function askAgain() {
        setPlanKey(undefined);
        setRevision(value => value + 1);
    }

    const chooseFile = (next: File | undefined) => {
        latest.current++;
        setFile(next);
        setPassphrase("");
        setSubmittedPassphrase(undefined);
        setNeedsPassphrase(next?.name.endsWith(".age") ?? false);
        setChecked(undefined);
        setPlan(undefined);
        setPlanKey(undefined);
        setValidateError(undefined);
        setPlanChanged(false);
        setResult(undefined);
    };

    const waitingForPassphrase = needsPassphrase && submittedPassphrase === undefined;
    const summary = plan?.summary ?? {};
    const blocked = summary["blocked"] ?? 0;
    const accepted = (summary["skipped"] ?? 0) + (summary["fixable"] ?? 0) + (summary["warning"] ?? 0);
    const isCurrent = planKey === requestKey && !isValidating;
    const nothingChosen = checked !== undefined && selection === undefined;
    const canImport = plan !== undefined && isCurrent && blocked === 0 && !nothingChosen && !isApplying;
    const existingHint = EXISTING_OPTIONS.find(option => option.value === options.existing)?.hint;

    return (
        <div className="flex flex-col gap-6">
            <div className="rounded-lg border bg-background p-4">
                <div className="flex flex-col items-start gap-6">
                    <p className="text-sm font-medium text-foreground">
                        Import a spec exported from HivePaaS into {scopeLabel}. Nothing is written until you confirm:
                        first you see what the import would do - what it creates, what it changes, what restarts, and
                        what it cannot bring over as it is.
                    </p>

                    <div className="w-full">
                        <InfoBlock
                            titleWidth={220}
                            title={
                                <LabelWithInfo
                                    label="Bundle"
                                    content="The .tar.gz or .tar.gz.age file an export downloaded. It stays in this page until you import it; reloading the page means choosing it again."
                                />
                            }
                        >
                            <div className="flex items-center gap-3">
                                <Button
                                    type="button"
                                    variant="outline"
                                    onClick={() => fileInputRef.current?.click()}
                                >
                                    <FileUpIcon className="size-4" />
                                    Choose file
                                </Button>
                                <span className="truncate text-sm text-muted-foreground">
                                    {file?.name ?? "No file chosen"}
                                </span>
                            </div>
                            <Input
                                ref={fileInputRef}
                                type="file"
                                accept=".gz,.tgz,.age"
                                className="hidden"
                                onChange={event => {
                                    chooseFile(event.target.files?.[0]);
                                    // The same file chosen again is a change too: the operator
                                    // may have exported it again in between.
                                    if (fileInputRef.current) {
                                        fileInputRef.current.value = "";
                                    }
                                }}
                            />
                            {file && !plan && !validateError && !waitingForPassphrase && (
                                <p className="mt-2 text-xs text-muted-foreground">Reading the bundle…</p>
                            )}
                            {validateError && (
                                <p className="mt-2 max-w-[560px] text-xs text-destructive">{validateError.message}</p>
                            )}
                        </InfoBlock>
                    </div>

                    {file && needsPassphrase && (
                        <div className="w-full">
                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Passphrase"
                                        content="The one the bundle was exported with. It is sent with each request and never stored."
                                    />
                                }
                            >
                                <form
                                    className="flex items-center gap-3"
                                    onSubmit={event => {
                                        event.preventDefault();
                                        setValidateError(undefined);
                                        setSubmittedPassphrase(passphrase);
                                        askAgain();
                                    }}
                                >
                                    <PasswordInput
                                        value={passphrase}
                                        onChange={event => {
                                            setPassphrase(event.target.value);
                                        }}
                                        placeholder="Required"
                                        className="max-w-[280px]"
                                    />
                                    <Button
                                        type="submit"
                                        variant="outline"
                                        disabled={passphrase.length === 0}
                                    >
                                        Open
                                    </Button>
                                </form>
                            </InfoBlock>
                        </div>
                    )}

                    {plan && !result && (
                        <>
                            <div className="w-full">
                                <InfoBlock
                                    titleWidth={220}
                                    title={<LabelWithInfo label="Exported" />}
                                >
                                    <p className="text-sm text-foreground">
                                        {plan.bundle.scope === "global" ? "A whole installation" : "One project"}, on{" "}
                                        {formatDate(plan.bundle.exportedAt)} by HivePaaS{" "}
                                        {plan.bundle.sourceAppVersion || "(version unknown)"}. Secrets:{" "}
                                        {plan.bundle.secretsMode === "omit"
                                            ? "not included"
                                            : `included, ${plan.bundle.secretsMode}`}
                                        .
                                    </p>
                                </InfoBlock>
                            </div>

                            <div className="w-full">
                                <InfoBlock
                                    titleWidth={220}
                                    title={
                                        <LabelWithInfo
                                            label="What exists here"
                                            content="Projects, environments, apps and settings this installation already has."
                                        />
                                    }
                                >
                                    <FieldGroup>
                                        <Field>
                                            <Select
                                                value={options.existing}
                                                onValueChange={value => {
                                                    setOptions(previous => ({
                                                        ...previous,
                                                        existing: value as SpecImportExisting,
                                                    }));
                                                }}
                                            >
                                                <SelectTrigger className="max-w-[280px]">
                                                    <SelectValue />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {EXISTING_OPTIONS.map(option => (
                                                        <SelectItem
                                                            key={option.value}
                                                            value={option.value}
                                                        >
                                                            {option.label}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                            {existingHint && (
                                                <p className="mt-2 max-w-[560px] text-xs text-muted-foreground">
                                                    {existingHint}
                                                </p>
                                            )}
                                        </Field>
                                    </FieldGroup>
                                </InfoBlock>
                            </div>

                            <div className="w-full">
                                <InfoBlock
                                    titleWidth={220}
                                    title={<LabelWithInfo label="Deploy" />}
                                >
                                    <div className="flex flex-col gap-3">
                                        <OptionCheckbox
                                            id="spec-import-deploy-created"
                                            label="Deploy the apps it creates"
                                            hint="Otherwise they start on a placeholder image, as an app created by hand does."
                                            checked={options.deployCreated}
                                            onChange={value => {
                                                setOptions(previous => ({ ...previous, deployCreated: value }));
                                            }}
                                        />
                                        <OptionCheckbox
                                            id="spec-import-deploy-changed-source"
                                            label="Deploy the apps whose source changes"
                                            hint="Otherwise they keep running their current image with the new configuration until someone deploys them."
                                            checked={options.deployChangedSource}
                                            onChange={value => {
                                                setOptions(previous => ({ ...previous, deployChangedSource: value }));
                                            }}
                                        />
                                    </div>
                                </InfoBlock>
                            </div>

                            <div className="flex w-full flex-col gap-2">
                                <div className="flex flex-wrap items-baseline justify-between gap-2">
                                    <p className="text-sm font-medium text-foreground">What the import would do</p>
                                    <p className="text-xs text-muted-foreground">
                                        {isValidating ? "Checking…" : describeSummary(summary)}
                                    </p>
                                </div>
                                {checked && (
                                    <SpecImportPlanTree
                                        roots={roots}
                                        checked={checked}
                                        onChange={setChecked}
                                    />
                                )}
                            </div>

                            {planChanged && (
                                <p className="w-full rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-400">
                                    Something changed on this installation since the plan was made, and so did the plan.
                                    Review it, and import again.
                                </p>
                            )}

                            <div className="flex w-full flex-col items-start gap-2">
                                {blocked > 0 && (
                                    <p className="flex items-center gap-2 text-sm text-destructive">
                                        <OctagonXIcon className="size-4" />
                                        {blocked} {blocked === 1 ? "issue blocks" : "issues block"} this import. Nothing
                                        can be written while {blocked === 1 ? "it stands" : "they stand"}.
                                    </p>
                                )}
                                {blocked === 0 && accepted > 0 && (
                                    <p className="max-w-[720px] text-sm text-muted-foreground">
                                        Importing accepts {accepted} {accepted === 1 ? "issue" : "issues"}: what is left
                                        out is not written, what is cleared is written without it, and a warning changes
                                        nothing but may not be what you meant.
                                    </p>
                                )}
                                {nothingChosen && (
                                    <p className="text-sm text-muted-foreground">Choose something to import.</p>
                                )}
                                <PermissionTooltipAction
                                    id={permissionModuleId}
                                    action="write"
                                >
                                    {({ isDenied }) => (
                                        <Button
                                            type="button"
                                            className="min-w-[120px]"
                                            disabled={!canImport || isDenied}
                                            isLoading={isApplying}
                                            onClick={() => {
                                                if (isDenied || !file || !selection) {
                                                    return;
                                                }
                                                setPlanChanged(false);
                                                apply({
                                                    scope,
                                                    bundle: file,
                                                    passphrase: submittedPassphrase,
                                                    selection,
                                                    options,
                                                    planHash: plan.planHash,
                                                    acceptIssues: accepted > 0,
                                                });
                                            }}
                                        >
                                            {accepted > 0 && blocked === 0
                                                ? `Import and accept ${accepted} ${accepted === 1 ? "issue" : "issues"}`
                                                : "Import"}
                                        </Button>
                                    )}
                                </PermissionTooltipAction>
                            </div>
                        </>
                    )}
                </div>
            </div>

            {result && (
                <SpecImportResultPanel
                    result={result}
                    onImportAgain={() => {
                        setResult(undefined);
                        askAgain();
                    }}
                    onChooseAnother={() => {
                        chooseFile(undefined);
                    }}
                />
            )}
        </div>
    );
}

function OptionCheckbox({ id, label, hint, checked, onChange }: OptionCheckboxProps) {
    return (
        <div className="flex items-start gap-2">
            <Checkbox
                id={id}
                className="mt-0.5"
                checked={checked}
                onCheckedChange={value => {
                    onChange(value === true);
                }}
            />
            <div className="flex flex-col">
                <Label
                    htmlFor={id}
                    className="text-sm font-normal"
                >
                    {label}
                </Label>
                <p className="text-xs text-muted-foreground">{hint}</p>
            </div>
        </div>
    );
}

interface OptionCheckboxProps {
    id: string;
    label: string;
    hint: string;
    checked: boolean;
    onChange: (checked: boolean) => void;
}

/** The plan's counts, in a line: what it creates and changes, and what it finds. */
function describeSummary(summary: Record<string, number>): string {
    const parts = [
        [summary["create"], "to create"],
        [summary["update"], "to update"],
        [summary["unchanged"], "unchanged"],
        [summary["keep"], "kept as is"],
        [summary["skip"], "left out"],
        [summary["restart"], "restart"],
        [summary["deploy"], "to deploy"],
    ] as const;
    const counted = parts.filter(([count]) => (count ?? 0) > 0).map(([count, label]) => `${count} ${label}`);
    const issues = ["blocked", "skipped", "fixable", "warning"]
        .filter(severity => (summary[severity] ?? 0) > 0)
        .map(severity => `${summary[severity]} ${severity}`);
    if (issues.length > 0) {
        counted.push(`issues: ${issues.join(", ")}`);
    }

    return counted.length > 0 ? counted.join(" · ") : "Nothing chosen";
}

interface Props {
    scope: SpecImportScope;
    /** Read in a sentence: "into {scopeLabel}". */
    scopeLabel: string;
    /** The module whose write permission importing takes. */
    permissionModuleId?: ModuleId;
}

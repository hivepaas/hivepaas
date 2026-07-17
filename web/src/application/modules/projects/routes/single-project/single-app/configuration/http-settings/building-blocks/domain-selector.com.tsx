import { useMemo, useState } from "react";

import { flushSync } from "react-dom";

import { Button, Checkbox, FieldError, Input } from "@components/ui";
import { cn } from "@lib/utils";
import { ArrowBigLeftDash, Check, EyeIcon, Plus, X } from "lucide-react";
import { useController, useFieldArray, useFormContext, useWatch } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { PopConfirm } from "@application/shared/components/pop-confirm";
import { isValidDomain } from "@application/shared/utils/domain";

import {
    type AppConfigHttpSettingsFormSchemaInput,
    type AppConfigHttpSettingsFormSchemaOutput,
    createDefaultCompressionConfig,
    emptyDomain,
} from "../schemas";

interface DomainChipProps {
    domain: string;
    isFirst: boolean;
    isActive: boolean;
    readOnly: boolean;
    onClick: () => void;
    onMoveLeft: () => void;
    onView: () => void;
    onRemove: () => void;
}

function DomainChip({ domain, isFirst, isActive, readOnly, onClick, onMoveLeft, onView, onRemove }: DomainChipProps) {
    return (
        <div
            role="button"
            tabIndex={0}
            className={cn(
                "flex items-center gap-1.5 rounded-md border px-2 py-1.25 cursor-pointer select-none",
                isActive ? "border-primary bg-primary/10 text-primary" : "text-muted-foreground",
            )}
            onClick={onClick}
        >
            <span className="max-w-[160px] truncate">{domain || "(empty)"}</span>
            {!isFirst && !readOnly && (
                <button
                    type="button"
                    title="Move left"
                    className="flex items-center rounded hover:text-primary"
                    onClick={e => {
                        e.stopPropagation();
                        onMoveLeft();
                    }}
                >
                    <ArrowBigLeftDash className="size-4 cursor-pointer" />
                </button>
            )}
            <button
                type="button"
                title="View domain"
                className="flex items-center rounded hover:text-primary"
                onClick={e => {
                    e.stopPropagation();
                    onView();
                }}
            >
                <EyeIcon className="size-4 cursor-pointer" />
            </button>
            {!readOnly && (
                <PopConfirm
                    title="Remove domain"
                    description="Confirm deletion of this domain?"
                    confirmText="Remove"
                    cancelText="Cancel"
                    variant="destructive"
                    onConfirm={onRemove}
                >
                    <button
                        type="button"
                        title="Remove domain"
                        className="flex items-center rounded hover:text-destructive"
                        onClick={e => {
                            e.stopPropagation();
                        }}
                    >
                        <X className="size-4 cursor-pointer" />
                    </button>
                </PopConfirm>
            )}
        </div>
    );
}

interface DomainSelectorProps {
    activeDomainIndex: number;
    setActiveDomainIndex: (index: number) => void;
    internalEndpoints: string[];
    domainSuggestion: string;
    readOnly?: boolean;
}

export function DomainSelector({
    activeDomainIndex,
    setActiveDomainIndex,
    internalEndpoints,
    domainSuggestion,
    readOnly = false,
}: DomainSelectorProps) {
    const { control, clearErrors, getValues, reset } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const { fields, append, move } = useFieldArray({ control, name: "domains" });
    const { field: exposePublicly } = useController({ control, name: "exposePublicly" });
    const domainValues = useWatch({ control, name: "domains", defaultValue: [] });
    const normalizeDomain = useMemo(() => (value: string) => value.trim().toLowerCase(), []);
    const suggestedDomain = domainSuggestion.trim();

    const [isAdding, setIsAdding] = useState(false);
    const [newDomainDraft, setNewDomainDraft] = useState(suggestedDomain);
    const [domainInputError, setDomainInputError] = useState<string | null>(null);

    function handleMoveLeft(index: number) {
        move(index, index - 1);
        setActiveDomainIndex(index - 1);
    }

    function handleRemoveDomain(index: number) {
        if (readOnly) {
            return;
        }

        const values = getValues();
        // RHF can leave null holes after unregister; strip them when compacting.
        const rawDomains = values.domains as ((typeof values.domains)[number] | null)[];
        const nextDomains = rawDomains.filter((d, i): d is (typeof values.domains)[number] => i !== index && d != null);

        let nextActive = -1;
        if (nextDomains.length > 0) {
            if (activeDomainIndex === index) {
                // Deleted active → select neighbor (next at same index, or previous if last).
                nextActive = Math.min(index, nextDomains.length - 1);
            } else if (activeDomainIndex > index) {
                nextActive = activeDomainIndex - 1;
            } else {
                nextActive = activeDomainIndex;
            }
        }

        // Unmount domain detail controllers before reset so RHF does not leave null/phantom holes.
        flushSync(() => {
            setActiveDomainIndex(-1);
        });
        reset({ ...values, domains: nextDomains }, { keepDefaultValues: true, keepDirty: true, keepTouched: true });
        clearErrors();
        setActiveDomainIndex(nextActive);
    }

    function handleConfirmAdd() {
        if (readOnly) return;

        const firstPort = domainValues[0]?.containerPort;
        const containerPort =
            typeof firstPort === "number" && Number.isInteger(firstPort) && firstPort >= 1 && firstPort <= 65535
                ? firstPort
                : emptyDomain.containerPort;

        const trimmedDraft = newDomainDraft.trim();
        const isFirstDomain = domainValues.length === 0;
        const firstDomainDefaults = isFirstDomain ? { compressionConfig: createDefaultCompressionConfig() } : {};

        if (trimmedDraft.length === 0) {
            setDomainInputError("Domain is required");
            return;
        }
        if (!isValidDomain(trimmedDraft, { maxLength: 100 })) {
            setDomainInputError("Enter a valid domain (e.g. app.example.com)");
            return;
        }
        const hasDuplicateDomain = domainValues.some(
            domain => normalizeDomain(domain.domain) === normalizeDomain(trimmedDraft),
        );
        if (hasDuplicateDomain) {
            setDomainInputError("Domain already exists.");
            return;
        }
        append({ ...emptyDomain, domain: trimmedDraft, containerPort, ...firstDomainDefaults });
        setActiveDomainIndex(domainValues.length);

        setIsAdding(false);
        setNewDomainDraft("");
        setDomainInputError(null);
    }

    function handleCancelAdd() {
        setIsAdding(false);
        setNewDomainDraft("");
        setDomainInputError(null);
    }

    function handleViewDomain(domain: string) {
        const url = /^https?:\/\//i.test(domain) ? domain : `https://${domain}`;
        window.open(url, "_blank", "noopener,noreferrer");
    }

    return (
        <div className="flex flex-col gap-4">
            <InfoBlock title="Project Internal Endpoint(s)">
                <div className="flex flex-wrap gap-2">
                    {internalEndpoints.map(ep => (
                        <span
                            key={ep}
                            className="font-mono text-xs text-red-500"
                        >
                            {ep}
                        </span>
                    ))}
                </div>
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Expose The App To The Internet"
                        content="Allow external access to this app via HTTP/HTTPS"
                    />
                }
            >
                <Checkbox
                    checked={exposePublicly.value}
                    onCheckedChange={value => {
                        if (readOnly) return;
                        exposePublicly.onChange(value);
                    }}
                    disabled={readOnly}
                />
            </InfoBlock>

            {exposePublicly.value && (
                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="Domains"
                            content="Manage domains for this app. Click a domain to edit its settings."
                        />
                    }
                >
                    <div className="flex flex-wrap items-center gap-2">
                        {fields.map((field, i) => (
                            <DomainChip
                                key={field.id}
                                domain={domainValues[i]?.domain ?? ""}
                                isFirst={i === 0}
                                isActive={i === activeDomainIndex}
                                readOnly={readOnly}
                                onClick={() => {
                                    setActiveDomainIndex(i);
                                }}
                                onMoveLeft={() => {
                                    handleMoveLeft(i);
                                }}
                                onView={() => {
                                    handleViewDomain(domainValues[i]?.domain ?? "");
                                }}
                                onRemove={() => {
                                    handleRemoveDomain(i);
                                }}
                            />
                        ))}

                        {isAdding ? (
                            <div className="flex items-center gap-1">
                                <Input
                                    value={newDomainDraft}
                                    onChange={e => {
                                        setNewDomainDraft(e.target.value);
                                        setDomainInputError(null);
                                    }}
                                    placeholder="e.g. app.example.com"
                                    className="h-8 w-48 text-xs"
                                    onKeyDown={e => {
                                        if (e.key === "Enter") {
                                            e.preventDefault();
                                            handleConfirmAdd();
                                        }
                                        if (e.key === "Escape") {
                                            handleCancelAdd();
                                        }
                                    }}
                                />
                                <Button
                                    type="button"
                                    size="icon"
                                    className="h-7 w-7"
                                    title="Confirm"
                                    onClick={handleConfirmAdd}
                                >
                                    <Check className="size-3" />
                                </Button>
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="h-7 w-7"
                                    title="Cancel"
                                    onClick={handleCancelAdd}
                                >
                                    <X className="size-3" />
                                </Button>
                            </div>
                        ) : (
                            !readOnly && (
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="icon"
                                    className="h-8 w-8"
                                    title="Add domain"
                                    onClick={() => {
                                        setIsAdding(true);
                                    }}
                                >
                                    <Plus className="size-4" />
                                </Button>
                            )
                        )}
                    </div>
                    {domainInputError && <FieldError errors={[{ message: domainInputError }]} />}
                </InfoBlock>
            )}
        </div>
    );
}

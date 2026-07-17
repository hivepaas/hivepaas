import { useMemo, useState } from "react";

import { flushSync } from "react-dom";

import { Button, FieldError, Input } from "@components/ui";
import { cn } from "@lib/utils";
import { Check, EyeIcon, Plus, X } from "lucide-react";
import { useFieldArray, useFormContext, useWatch } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { PopConfirm } from "@application/shared/components/pop-confirm";
import { isValidDomain } from "@application/shared/utils/domain";

import {
    type HivePaaSHttpSettingsFormInput,
    type HivePaaSHttpSettingsFormOutput,
    MAX_HIVEPAAS_HTTP_DOMAINS,
    emptyDomain,
} from "../schemas";

interface DomainChipProps {
    domain: string;
    isFirst: boolean;
    isActive: boolean;
    readOnly: boolean;
    onClick: () => void;
    onMakeDefault: () => void;
    onView: () => void;
    onRemove: () => void;
}

function DomainChip({
    domain,
    isFirst,
    isActive,
    readOnly,
    onClick,
    onMakeDefault,
    onView,
    onRemove,
}: DomainChipProps) {
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
                    title="Make Default"
                    className="text-xs text-primary hover:underline whitespace-nowrap"
                    onClick={e => {
                        e.stopPropagation();
                        onMakeDefault();
                    }}
                >
                    Make Default
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
            {!isFirst && !readOnly && (
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
    readOnly?: boolean;
}

export function DomainSelector({ activeDomainIndex, setActiveDomainIndex, readOnly = false }: DomainSelectorProps) {
    const { control, clearErrors, getValues, reset } = useFormContext<
        HivePaaSHttpSettingsFormInput,
        unknown,
        HivePaaSHttpSettingsFormOutput
    >();

    const { fields, append, move } = useFieldArray({ control, name: "domains" });
    const domainValues = useWatch({ control, name: "domains", defaultValue: [] });
    const normalizeDomain = useMemo(() => (value: string) => value.trim().toLowerCase(), []);

    const [isAdding, setIsAdding] = useState(false);
    const [newDomainDraft, setNewDomainDraft] = useState("");
    const [domainInputError, setDomainInputError] = useState<string | null>(null);

    const canAddDomain = domainValues.length < MAX_HIVEPAAS_HTTP_DOMAINS;

    function handleMakeDefault(index: number) {
        if (readOnly || index === 0) {
            return;
        }
        move(index, 0);
        setActiveDomainIndex(0);
    }

    function handleRemoveDomain(index: number) {
        if (readOnly || index === 0) {
            return;
        }

        const values = getValues();
        const rawDomains = values.domains as ((typeof values.domains)[number] | null)[];
        const nextDomains = rawDomains.filter((d, i): d is (typeof values.domains)[number] => i !== index && d != null);

        let nextActive = -1;
        if (nextDomains.length > 0) {
            if (activeDomainIndex === index) {
                nextActive = Math.min(index, nextDomains.length - 1);
            } else if (activeDomainIndex > index) {
                nextActive = activeDomainIndex - 1;
            } else {
                nextActive = activeDomainIndex;
            }
        }

        flushSync(() => {
            setActiveDomainIndex(-1);
        });
        reset({ ...values, domains: nextDomains }, { keepDefaultValues: true, keepDirty: true, keepTouched: true });
        clearErrors();
        setActiveDomainIndex(nextActive);
    }

    function handleConfirmAdd() {
        if (readOnly || !canAddDomain) {
            return;
        }

        const trimmedDraft = newDomainDraft.trim();

        if (trimmedDraft.length > 0) {
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
            append({ ...emptyDomain, domain: trimmedDraft });
            setActiveDomainIndex(domainValues.length);
        } else {
            append({ ...emptyDomain });
            setActiveDomainIndex(domainValues.length);
        }

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
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Domains"
                    content="Manage domains for HivePaaS. The first domain is the default. Maximum 2 domains."
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
                        onMakeDefault={() => {
                            handleMakeDefault(i);
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
                    !readOnly &&
                    canAddDomain && (
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
    );
}

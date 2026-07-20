import React, { useEffect, useMemo, useRef, useState } from "react";

import { cn } from "@/lib/utils";
import { Button } from "@components/ui/button";
import { Checkbox } from "@components/ui/checkbox";
import { Field, FieldError } from "@components/ui/field";
import { Input } from "@components/ui/input";
import { Textarea } from "@components/ui/textarea";
import { Tooltip, TooltipContent, TooltipTrigger } from "@components/ui/tooltip";
import { Plus, Trash2, WrapText } from "lucide-react";
import { type Path, get, useFieldArray, useFormContext, useWatch } from "react-hook-form";

import { MULTILINE_ENV_SEPARATOR, parseEnvVars, stringifyEnvVars } from "@application/shared/utils/env-vars";

type EnvVarRecord = {
    key: string;
    value: string;
    isLiteral: boolean;
};

const MERGE_VIEW_PLACEHOLDER = `KEY_1=VALUE_1
${MULTILINE_ENV_SEPARATOR}
KEY_2=
value line
${MULTILINE_ENV_SEPARATOR}`;

function View<T>({ name, search = "", viewMode = "individual", isRevealed = false, readOnly = false }: Props<T>) {
    const {
        control,
        register,
        formState: { errors },
    } = useFormContext<Record<string, EnvVarRecord[]>>();

    const { fields, append, remove, replace, update } = useFieldArray({
        control,
        name: name as string,
    });

    const fieldValuesWatch = useWatch({
        control,
        name: name as string,
    });

    const fieldValues = useMemo(() => {
        return fieldValuesWatch;
    }, [fieldValuesWatch]);

    const previousViewModeRef = useRef<"merge" | "individual">(viewMode);
    const [mergeText, setMergeText] = useState<string | null>(null);
    const [multilineFieldIds, setMultilineFieldIds] = useState<Set<string>>(() => new Set());

    useEffect(() => {
        setMultilineFieldIds(previousIds => {
            const nextIds = new Set(previousIds);
            let changed = false;
            const currentFieldIds = new Set(fields.map(field => field.id));

            for (const fieldId of Array.from(nextIds)) {
                if (!currentFieldIds.has(fieldId)) {
                    nextIds.delete(fieldId);
                    changed = true;
                }
            }

            fields.forEach((field, index) => {
                const record = fieldValues[index];

                if (record?.value.includes("\n") && !nextIds.has(field.id)) {
                    nextIds.add(field.id);
                    changed = true;
                }
            });

            return changed ? nextIds : previousIds;
        });
    }, [fields, fieldValues]);

    // Filter records based on search query
    const filteredFields = useMemo(() => {
        if (!search || search.trim() === "") {
            return fields.map((_field, index) => index);
        }

        const searchLower = search.toLowerCase().trim();
        return fields
            .map((_field, index) => index)
            .filter(index => {
                const record = fieldValues[index];
                if (!record) return false;

                const keyMatch = record.key.toLowerCase().includes(searchLower);
                const valueMatch = record.value.toLowerCase().includes(searchLower);
                return keyMatch || valueMatch;
            });
    }, [fields, fieldValues, search]);

    // Handle view mode switching
    useEffect(() => {
        // When switching from individual to merge
        if (previousViewModeRef.current === "individual" && viewMode === "merge") {
            const text = stringifyEnvVars(fieldValues);
            setMergeText(text);
        }

        // When switching from merge to individual
        if (previousViewModeRef.current === "merge" && viewMode === "individual") {
            if (readOnly) {
                previousViewModeRef.current = viewMode;
                return;
            }

            const text = mergeText ?? stringifyEnvVars(fieldValues);
            const parsedArray = parseEnvVars(text);
            replace(parsedArray);
        }

        previousViewModeRef.current = viewMode;
    }, [viewMode, fieldValues, replace, mergeText, readOnly]);

    // Handle merge textarea changes
    const handleMergeTextChange = (value: string) => {
        if (readOnly) {
            return;
        }

        setMergeText(value);
    };

    const syncMergeTextToFields = () => {
        if (readOnly) {
            return;
        }

        const text = mergeText ?? "";
        const parsedArray = parseEnvVars(text);
        replace(parsedArray);
    };

    // Get merge text value
    const mergeTextValue = useMemo(() => {
        if (viewMode === "merge") {
            if (mergeText !== null) {
                return mergeText;
            }
            return stringifyEnvVars(fieldValues);
        }
        return "";
    }, [viewMode, fieldValues, mergeText]);

    const handleAdd = () => {
        if (readOnly) {
            return;
        }

        append({ key: "", value: "", isLiteral: false });
    };

    const handleMultilineToggle = (fieldId: string, value: string) => {
        if (readOnly) {
            return;
        }

        const isMultiline = multilineFieldIds.has(fieldId) || value.includes("\n");

        if (isMultiline) {
            if (value.includes("\n")) {
                return;
            }

            setMultilineFieldIds(previousIds => {
                const nextIds = new Set(previousIds);
                nextIds.delete(fieldId);
                return nextIds;
            });
            return;
        }

        setMultilineFieldIds(previousIds => {
            const nextIds = new Set(previousIds);
            nextIds.add(fieldId);
            return nextIds;
        });
    };

    return viewMode === "individual" ? (
        <div className="flex flex-col gap-4">
            {/* Individual Records List */}
            {filteredFields.length > 0 ? (
                <div className="space-y-3">
                    {filteredFields.map(index => {
                        const record = fieldValues[index];
                        const field = fields[index];
                        if (!field || !record) return null;

                        const hasMultilineValue = record.value.includes("\n");
                        const isMultiline = multilineFieldIds.has(field.id) || hasMultilineValue;
                        const cannotDisableMultiline = isMultiline && hasMultilineValue;

                        return (
                            <div
                                key={field.id}
                                className="flex items-start gap-3"
                            >
                                {/* Key Input */}
                                <div className="min-w-0 flex-1">
                                    <Field>
                                        <Input
                                            id={`${name}-${index}-key`}
                                            {...register(`${name}.${index}.key`)}
                                            placeholder="Key"
                                            aria-invalid={!!get(errors, `${name}.${index}.key`)}
                                            disabled={readOnly}
                                        />
                                        <FieldError errors={[get(errors, `${name}.${index}.key`)]} />
                                    </Field>
                                </div>

                                {/* Value Input */}
                                <div className="min-w-0 flex-1">
                                    <Field>
                                        {isMultiline ? (
                                            <Textarea
                                                id={`${name}-${index}-value`}
                                                {...register(`${name}.${index}.value`)}
                                                placeholder="Value"
                                                aria-invalid={!!get(errors, `${name}.${index}.value`)}
                                                disabled={readOnly}
                                                minRows={4}
                                                maxRows={0}
                                                className={cn(
                                                    "min-h-24 resize-y",
                                                    !isRevealed && "[-webkit-text-security:disc]",
                                                )}
                                            />
                                        ) : (
                                            <Input
                                                id={`${name}-${index}-value`}
                                                type={isRevealed ? "text" : "password"}
                                                {...register(`${name}.${index}.value`)}
                                                placeholder="Value"
                                                aria-invalid={!!get(errors, `${name}.${index}.value`)}
                                                disabled={readOnly}
                                            />
                                        )}
                                        <FieldError errors={[get(errors, `${name}.${index}.value`)]} />
                                    </Field>
                                </div>

                                {/* Multi-line Toggle */}
                                <Tooltip>
                                    <TooltipTrigger asChild>
                                        <Button
                                            type="button"
                                            variant={isMultiline ? "secondary" : "ghost"}
                                            size="icon"
                                            onClick={() => {
                                                handleMultilineToggle(field.id, record.value);
                                            }}
                                            disabled={readOnly}
                                            aria-label="Toggle multi-line mode"
                                            aria-pressed={isMultiline}
                                            aria-disabled={readOnly || cannotDisableMultiline || undefined}
                                            className={cn(
                                                "h-8 w-8 text-muted-foreground",
                                                isMultiline && "text-foreground",
                                                cannotDisableMultiline && "cursor-not-allowed opacity-50",
                                            )}
                                        >
                                            <WrapText className="size-4" />
                                        </Button>
                                    </TooltipTrigger>
                                    <TooltipContent side="top">Toggle multi-line mode</TooltipContent>
                                </Tooltip>

                                {/* Literal Checkbox */}
                                <div className="flex h-9 items-center gap-2">
                                    <Checkbox
                                        id={`${name}-${index}-literal`}
                                        checked={record.isLiteral}
                                        onCheckedChange={checked => {
                                            if (readOnly) {
                                                return;
                                            }

                                            update(index, {
                                                ...record,
                                                isLiteral: checked === true,
                                            });
                                        }}
                                        disabled={readOnly}
                                    />
                                    <label
                                        htmlFor={`${name}-${index}-literal`}
                                        className="text-sm cursor-pointer"
                                    >
                                        Literal
                                    </label>
                                </div>

                                {/* Delete Button */}
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    onClick={() => {
                                        if (readOnly) {
                                            return;
                                        }

                                        remove(index);
                                    }}
                                    disabled={readOnly}
                                    className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
                                >
                                    <Trash2 className="size-4" />
                                </Button>
                            </div>
                        );
                    })}
                </div>
            ) : (
                <div className="text-sm text-muted-foreground py-4 text-center">
                    {search ? "No records match your search" : "No results"}
                </div>
            )}

            {/* Add Button */}
            <Button
                type="button"
                variant="outline"
                onClick={handleAdd}
                disabled={readOnly}
                className="w-fit"
            >
                <Plus className="size-4" />
                Add
            </Button>
        </div>
    ) : (
        <div className="flex flex-col gap-2">
            {/* Merge View Textarea */}
            <Textarea
                value={mergeTextValue}
                onChange={e => {
                    handleMergeTextChange(e.target.value);
                }}
                onBlur={() => {
                    syncMergeTextToFields();
                }}
                placeholder={MERGE_VIEW_PLACEHOLDER}
                rows={10}
                className="font-mono text-sm min-h-[500px]"
                disabled={readOnly}
            />
            <p className="text-xs text-muted-foreground">
                Enter environment variables in key=value format, one per line
            </p>
        </div>
    );
}

type Props<T> = {
    name: Path<T>;
    search?: string;
    viewMode?: "merge" | "individual";
    isRevealed?: boolean;
    readOnly?: boolean;
};

export const ConfigVariables = React.memo(View) as typeof View;

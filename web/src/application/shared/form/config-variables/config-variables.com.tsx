import React, { useEffect, useMemo, useRef, useState } from "react";

import { Button } from "@components/ui/button";
import { Checkbox } from "@components/ui/checkbox";
import { Field, FieldError } from "@components/ui/field";
import { Input } from "@components/ui/input";
import { Textarea } from "@components/ui/textarea";
import { Plus, Trash2 } from "lucide-react";
import { type Path, get, useFieldArray, useFormContext, useWatch } from "react-hook-form";

import { parseEnvVars, stringifyEnvVars } from "@application/shared/utils/env-vars";

type EnvVarRecord = {
    key: string;
    value: string;
    isLiteral: boolean;
};

function View<T>({ name, search = "", viewMode = "individual", isRevealed = false }: Props<T>) {
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
            const text = mergeText ?? stringifyEnvVars(fieldValues);
            const parsedArray = parseEnvVars(text);
            replace(parsedArray);
        }

        previousViewModeRef.current = viewMode;
    }, [viewMode, fieldValues, replace, mergeText]);

    // Handle merge textarea changes
    const handleMergeTextChange = (value: string) => {
        setMergeText(value);
    };

    const syncMergeTextToFields = () => {
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
        append({ key: "", value: "", isLiteral: false });
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

                        return (
                            <div
                                key={field.id}
                                className="flex items-center gap-3"
                            >
                                {/* Key Input */}
                                <div className="flex-1">
                                    <Field>
                                        <Input
                                            id={`${name}-${index}-key`}
                                            {...register(`${name}.${index}.key`)}
                                            placeholder="Key"
                                            aria-invalid={!!get(errors, `${name}.${index}.key`)}
                                        />
                                        <FieldError errors={[get(errors, `${name}.${index}.key`)]} />
                                    </Field>
                                </div>

                                {/* Value Input */}
                                <div className="flex-1">
                                    <Field>
                                        <Input
                                            type={isRevealed ? "text" : "password"}
                                            {...register(`${name}.${index}.value`)}
                                            placeholder="Value"
                                            aria-invalid={!!get(errors, `${name}.${index}.value`)}
                                        />
                                        <FieldError errors={[get(errors, `${name}.${index}.value`)]} />
                                    </Field>
                                </div>

                                {/* Literal Checkbox */}
                                <div className="flex items-center gap-2">
                                    <Checkbox
                                        id={`${name}-${index}-literal`}
                                        checked={record.isLiteral}
                                        onCheckedChange={checked => {
                                            update(index, {
                                                ...record,
                                                isLiteral: checked === true,
                                            });
                                        }}
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
                                        remove(index);
                                    }}
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
                    {search ? "No records match your search" : "No environment variables"}
                </div>
            )}

            {/* Add Button */}
            <Button
                type="button"
                variant="outline"
                onClick={handleAdd}
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
                placeholder="KEY_1=VALUE_1&#10;KEY_2=VALUE_2"
                rows={10}
                className="font-mono text-sm"
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
};

export const ConfigVariables = React.memo(View) as typeof View;

import React, { useState } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { ChevronDown, ChevronRight, Plus, Trash2, X } from "lucide-react";
import { get, useController, useFieldArray, useFormContext, useWatch } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { EAppScheduledJobArgSeparator } from "~/projects/module-shared/enums";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import {
    Button,
    Checkbox,
    Collapsible,
    CollapsibleContent,
    CollapsibleTrigger,
    Field,
    FieldError,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui";
import { Textarea } from "@/components/ui/textarea";

import { createDefaultCommandArg, createDefaultCommandArgGroup } from "./command-arg-groups-section.helpers";
import type { CommandArgGroupsFormValue } from "./command-arg-groups-section.types";

type SchemaInput = CommandArgGroupsFormValue;
type SchemaOutput = CommandArgGroupsFormValue;

function isQuoted(value: unknown): boolean {
    if (typeof value !== "string" || value.length < 2) {
        return false;
    }

    const firstChar = value[0];
    const lastChar = value[value.length - 1];

    return (firstChar === "'" || firstChar === '"') && firstChar === lastChar; // eslint-disable-line quotes
}

function quotePreviewValue(value: string) {
    if (isQuoted(value)) {
        return value;
    }

    return `"${value.replace(/\\/g, "\\\\").replace(/"/g, '\\"')}"`; // eslint-disable-line quotes
}

function buildArgGroupPreview(group: SchemaInput["argGroups"][number] | undefined) {
    if (!group?.enabled || !group.exportEnv.trim()) {
        return "";
    }

    const args = group.args
        .filter(arg => arg.use && arg.name.trim())
        .map(arg => {
            if (!arg.value) {
                return arg.name.trim();
            }

            return `${arg.name.trim()}${group.separator}${quotePreviewValue(arg.value)}`;
        });

    if (args.length === 0) {
        return `${group.exportEnv}=`;
    }

    return `${group.exportEnv}=${args.join(" ")}`;
}

function ArgItem({
    groupIndex,
    argIndex,
    separator = "=",
    onRemove,
    readOnly = false,
    fieldName = "argGroups",
}: ArgItemProps) {
    const {
        control,
        formState: { errors },
    } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const basePath = `${fieldName}.${groupIndex}.args.${argIndex}`;

    const { field: useArg } = useController({ control, name: `${basePath}.use` as never });
    const {
        field: name,
        fieldState: { error: nameError },
    } = useController({ control, name: `${basePath}.name` as never });
    const { field: value } = useController({ control, name: `${basePath}.value` as never });
    const nameValue = typeof name.value === "string" ? name.value : "";
    const argValue = typeof value.value === "string" ? value.value : "";

    return (
        <div className="flex w-full flex-col gap-1">
            <div
                className={cn(
                    "flex w-full items-start gap-1.5 sm:gap-2 rounded-md border bg-background px-2.5 py-1 transition-colors",
                    useArg.value === true && "border-green-200 bg-green-50 dark:bg-green-950/20",
                )}
            >
                <div className="flex h-8 shrink-0 items-center">
                    <Checkbox
                        checked={useArg.value === true}
                        onCheckedChange={checked => {
                            if (readOnly) {
                                return;
                            }

                            useArg.onChange(checked === true);
                        }}
                        disabled={readOnly}
                    />
                </div>
                <Input
                    {...name}
                    value={nameValue}
                    onChange={name.onChange}
                    placeholder="--arg"
                    className="h-8 w-[140px] shrink-0 px-2 font-mono text-xs sm:w-[220px] sm:text-sm"
                    aria-invalid={!!get(errors, `${basePath}.name`)}
                    disabled={readOnly}
                />
                <div className="flex h-8 shrink-0 items-center">
                    <span className="text-xs sm:text-sm text-muted-foreground font-mono">{separator || "="}</span>
                </div>
                <Textarea
                    {...value}
                    value={argValue}
                    onChange={value.onChange}
                    placeholder="value"
                    minRows={1}
                    maxRows={1}
                    className="min-h-8 w-full min-w-0 flex-1 resize-y overflow-y-auto py-1 px-2 text-xs sm:text-sm leading-normal font-mono"
                    disabled={readOnly}
                />
                <div className="flex h-8 shrink-0 items-center gap-0.5 ml-auto sm:ml-0">
                    <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="size-6 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
                        onClick={onRemove}
                        disabled={readOnly}
                    >
                        <X className="size-3.5" />
                        <span className="sr-only">Remove arg</span>
                    </Button>
                </div>
            </div>
            <FieldError errors={[nameError]} />
        </div>
    );
}

function ArgGroupRow({ groupIndex, onRemove, readOnly = false, fieldName = "argGroups" }: ArgGroupRowProps) {
    const [open, setOpen] = useState(true);

    const {
        control,
        formState: { errors },
    } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const basePath = `${fieldName}.${groupIndex}`;

    const { field: enabled } = useController({ control, name: `${basePath}.enabled` as never });
    const {
        field: exportEnv,
        fieldState: { error: exportEnvError },
    } = useController({ control, name: `${basePath}.exportEnv` as never });
    const { field: separator } = useController({ control, name: `${basePath}.separator` as never });
    const { fields, append, remove } = useFieldArray({
        control,
        name: `${basePath}.args` as never,
    });
    const groupValue = useWatch({ control, name: basePath as never }) as SchemaInput["argGroups"][number] | undefined;
    const preview = buildArgGroupPreview(groupValue);
    const exportEnvValue = groupValue?.exportEnv ?? "";

    return (
        <Collapsible
            open={open}
            onOpenChange={setOpen}
        >
            <div className="flex items-center gap-2 rounded-md border px-2 py-0 bg-accent">
                <CollapsibleTrigger asChild>
                    <button
                        type="button"
                        className="flex flex-1 items-center gap-2 text-left"
                    >
                        {open ? (
                            <ChevronDown className="size-4 shrink-0" />
                        ) : (
                            <ChevronRight className="size-4 shrink-0" />
                        )}
                        <span className="font-mono text-sm text-red-500">
                            Arg Group: {exportEnvValue || `CMD_ARG_GROUP_${groupIndex + 1}`}
                        </span>
                    </button>
                </CollapsibleTrigger>
                <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    disabled={readOnly}
                    onClick={onRemove}
                >
                    <Trash2 className="size-3.5" />
                    <span className="sr-only">Remove arg group</span>
                </Button>
            </div>

            <CollapsibleContent>
                <div className="flex flex-col gap-4 border-l-2 border-accent pl-4 pt-3 pb-3 ml-3">
                    <InfoBlock
                        title="Enabled"
                        titleWidth={220}
                    >
                        <Checkbox
                            checked={enabled.value === true}
                            onCheckedChange={checked => {
                                if (readOnly) {
                                    return;
                                }

                                enabled.onChange(checked === true);
                            }}
                            disabled={readOnly}
                        />
                    </InfoBlock>

                    {enabled.value === true && (
                        <>
                            <InfoBlock
                                title={
                                    <LabelWithInfo
                                        label="Export Env"
                                        isRequired
                                    />
                                }
                                titleWidth={220}
                            >
                                <Field>
                                    <Input
                                        {...exportEnv}
                                        value={exportEnv.value}
                                        onChange={exportEnv.onChange}
                                        placeholder="CMD_ARG_GROUP_1"
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                        aria-invalid={!!get(errors, `${basePath}.exportEnv`)}
                                        disabled={readOnly}
                                    />
                                    <FieldError errors={[exportEnvError]} />
                                </Field>
                            </InfoBlock>

                            <InfoBlock
                                title="Arg Separator"
                                titleWidth={220}
                            >
                                <Select
                                    value={separator.value}
                                    onValueChange={value => {
                                        if (readOnly) {
                                            return;
                                        }

                                        separator.onChange(value);
                                    }}
                                    disabled={readOnly}
                                >
                                    <SelectTrigger className="max-w-[260px]">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value={EAppScheduledJobArgSeparator.Equal}>=</SelectItem>
                                        <SelectItem value={EAppScheduledJobArgSeparator.Whitespace}>
                                            whitespace
                                        </SelectItem>
                                    </SelectContent>
                                </Select>
                            </InfoBlock>

                            <InfoBlock
                                title=""
                                titleWidth={220}
                            >
                                <div className="flex w-full min-w-0 flex-1 flex-col gap-3">
                                    <div className="flex w-full flex-col gap-2.5">
                                        {fields.map((field, argIndex) => (
                                            <ArgItem
                                                key={field.id}
                                                groupIndex={groupIndex}
                                                argIndex={argIndex}
                                                separator={separator.value}
                                                onRemove={() => {
                                                    if (readOnly) {
                                                        return;
                                                    }

                                                    remove(argIndex);
                                                }}
                                                readOnly={readOnly}
                                                fieldName={fieldName}
                                            />
                                        ))}
                                    </div>
                                    <Button
                                        type="button"
                                        variant="outline"
                                        className="w-fit"
                                        disabled={readOnly}
                                        onClick={() => {
                                            append(createDefaultCommandArg() as never);
                                        }}
                                    >
                                        <Plus className="size-4" />
                                        Add Arg
                                    </Button>
                                </div>
                            </InfoBlock>

                            <div
                                className={cn(
                                    dashedBorderBox,
                                    "w-full text-xs sm:text-sm min-h-7 h-auto py-1.5 px-3 leading-normal",
                                )}
                            >
                                <div className="break-all">
                                    <span>(Illustration only) </span>
                                    <span className="text-orange-500">{exportEnvValue}</span>
                                    {preview ? preview.slice(exportEnvValue.length) : "="}
                                </div>
                            </div>
                        </>
                    )}
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}

function View({ readOnly = false, fieldName = "argGroups" }: Props) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { fields, append, remove } = useFieldArray({
        control,
        name: fieldName as never,
    });

    return (
        <div className="flex flex-col gap-3">
            {fields.map((field, index) => (
                <ArgGroupRow
                    key={field.id}
                    groupIndex={index}
                    onRemove={() => {
                        if (readOnly) {
                            return;
                        }

                        remove(index);
                    }}
                    readOnly={readOnly}
                    fieldName={fieldName}
                />
            ))}
            <Button
                type="button"
                variant="outline"
                className="w-fit"
                disabled={readOnly}
                onClick={() => {
                    append(createDefaultCommandArgGroup(fields.length));
                }}
            >
                <Plus className="size-4" />
                Add Arg Group
            </Button>
        </div>
    );
}

interface ArgItemProps {
    groupIndex: number;
    argIndex: number;
    separator?: string;
    onRemove: () => void;
    readOnly?: boolean;
    fieldName?: string;
}

interface ArgGroupRowProps {
    groupIndex: number;
    onRemove: () => void;
    readOnly?: boolean;
    fieldName?: string;
}

interface Props {
    readOnly?: boolean;
    fieldName?: string;
}

export const CommandArgGroupsSection = React.memo(View);

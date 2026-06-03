import React, { useState } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { ChevronDown, ChevronRight, Plus, Trash2, X } from "lucide-react";
import { get, useController, useFieldArray, useFormContext, useWatch } from "react-hook-form";
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

import type { CreateOrEditAppScheduledJobFormInput, CreateOrEditAppScheduledJobFormOutput } from "../../schemas";
import { createDefaultArg, createDefaultArgGroup } from "../create-or-edit-app-scheduled-job.form-mappers";

import { INFO_BLOCK_TITLE_WIDTH } from "./app-scheduled-job-form.constants";

type SchemaInput = CreateOrEditAppScheduledJobFormInput;
type SchemaOutput = CreateOrEditAppScheduledJobFormOutput;

const CONTENT_COLUMN_OFFSET = INFO_BLOCK_TITLE_WIDTH + 12;

function isQuoted(value: unknown): boolean {
    if (typeof value !== "string" || value.length < 2) {
        return false;
    }

    const firstChar = value[0];
    const lastChar = value[value.length - 1];

    return (firstChar === "'" || firstChar === '"') && firstChar === lastChar;
}

function quotePreviewValue(value: string) {
    if (isQuoted(value)) {
        return value;
    }

    return `"${value.replace(/\\/g, "\\\\").replace(/"/g, '\\"')}"`;
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

function getInputWidth(value: string, placeholder: string, min: number, max: number) {
    const length = Math.max(value.length, placeholder.length) + 2;
    const width = Math.min(Math.max(length, min), max);

    return `${width}ch`;
}

function ArgItem({ groupIndex, argIndex, onRemove, readOnly = false }: ArgItemProps) {
    const {
        control,
        formState: { errors },
    } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const basePath = `argGroups.${groupIndex}.args.${argIndex}`;

    const { field: useArg } = useController({ control, name: `${basePath}.use` as never });
    const {
        field: name,
        fieldState: { error: nameError },
    } = useController({ control, name: `${basePath}.name` as never });
    const { field: value } = useController({ control, name: `${basePath}.value` as never });
    const nameValue = name.value;
    const argValue = value.value;

    return (
        <div className="flex flex-col gap-1">
            <div
                className={cn(
                    "flex w-fit max-w-full items-center gap-2 rounded-md border bg-background px-2 py-1",
                    useArg.value === true && "border-green-200 bg-green-50 dark:bg-green-950/20",
                )}
            >
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
                <Input
                    {...name}
                    value={nameValue}
                    onChange={name.onChange}
                    placeholder="--arg"
                    className="h-8 min-w-[8ch] max-w-[28ch]"
                    style={{ width: getInputWidth(nameValue, "--arg", 8, 28) }}
                    aria-invalid={!!get(errors, `${basePath}.name`)}
                    disabled={readOnly}
                />
                <span className="text-sm text-muted-foreground">=</span>
                <Input
                    {...value}
                    value={argValue}
                    onChange={value.onChange}
                    placeholder="value"
                    className="h-8 min-w-[8ch] max-w-[34ch]"
                    style={{ width: getInputWidth(argValue, "value", 8, 34) }}
                    disabled={readOnly}
                />
                <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="h-4 w-4 shrink-0 text-muted-foreground"
                    onClick={onRemove}
                    disabled={readOnly}
                >
                    <X className="size-4" />
                    <span className="sr-only">Remove arg</span>
                </Button>
            </div>
            <FieldError errors={[nameError]} />
        </div>
    );
}

function ArgGroupRow({ groupIndex, onRemove, readOnly = false }: ArgGroupRowProps) {
    const [open, setOpen] = useState(true);
    const {
        control,
        formState: { errors },
    } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const basePath = `argGroups.${groupIndex}`;

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
                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
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
                                titleWidth={INFO_BLOCK_TITLE_WIDTH}
                            >
                                <Field>
                                    <Input
                                        {...exportEnv}
                                        value={exportEnv.value}
                                        onChange={exportEnv.onChange}
                                        placeholder="CMD_ARG_GROUP_1"
                                        className="max-w-[400px]"
                                        aria-invalid={!!get(errors, `${basePath}.exportEnv`)}
                                        disabled={readOnly}
                                    />
                                    <FieldError errors={[exportEnvError]} />
                                </Field>
                            </InfoBlock>

                            <InfoBlock
                                title="Arg Separator"
                                titleWidth={INFO_BLOCK_TITLE_WIDTH}
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

                            <div
                                className="flex max-w-[760px] flex-col gap-3"
                                style={{ marginLeft: CONTENT_COLUMN_OFFSET }}
                            >
                                <div className="flex flex-wrap gap-3">
                                    {fields.map((field, argIndex) => (
                                        <ArgItem
                                            key={field.id}
                                            groupIndex={groupIndex}
                                            argIndex={argIndex}
                                            onRemove={() => {
                                                if (readOnly) {
                                                    return;
                                                }

                                                remove(argIndex);
                                            }}
                                            readOnly={readOnly}
                                        />
                                    ))}
                                </div>
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="icon"
                                    className="h-8 w-8"
                                    disabled={readOnly}
                                    onClick={() => {
                                        append(createDefaultArg() as never);
                                    }}
                                >
                                    <Plus className="size-4" />
                                    <span className="sr-only">Add arg</span>
                                </Button>
                            </div>

                            <div
                                className={cn(dashedBorderBox, "max-w-[760px] text-center text-sm")}
                                style={{ marginLeft: CONTENT_COLUMN_OFFSET }}
                            >
                                <div className="break-all">
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

function View({ readOnly = false }: Props) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { fields, append, remove } = useFieldArray({
        control,
        name: "argGroups",
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
                />
            ))}
            <Button
                type="button"
                variant="outline"
                className="w-fit"
                disabled={readOnly}
                onClick={() => {
                    append(createDefaultArgGroup(fields.length));
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
    onRemove: () => void;
    readOnly?: boolean;
}

interface ArgGroupRowProps {
    groupIndex: number;
    onRemove: () => void;
    readOnly?: boolean;
}

interface Props {
    readOnly?: boolean;
}

export const ArgGroupsSection = React.memo(View);

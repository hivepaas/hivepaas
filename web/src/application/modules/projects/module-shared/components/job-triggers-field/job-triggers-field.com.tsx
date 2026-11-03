import { useState } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { ChevronDown, Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext, useWatch } from "react-hook-form";
import { ESchedJobTriggerEvent, SCHED_JOB_TRIGGER_EVENT_OPTIONS } from "~/projects/module-shared/enums";

import { Button, Checkbox, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

import { JOB_MAX_TRIGGERS, type JobTriggerFormValue, createDefaultJobTrigger } from "./job-triggers.helpers";

interface NamedApp {
    id: string;
    name: string;
}

/** A pick of several apps, as checkboxes in a popover. */
function AppsPicker({ apps, value, onChange, readOnly }: AppsPickerProps) {
    const [open, setOpen] = useState(false);
    const names = apps.filter(app => value.includes(app.id)).map(app => app.name);

    return (
        <Popover
            open={open}
            onOpenChange={setOpen}
        >
            <PopoverTrigger asChild>
                <Button
                    type="button"
                    variant="outline"
                    disabled={readOnly}
                    className="h-9 w-[240px] justify-between font-normal"
                >
                    <span className={cn("truncate", names.length === 0 && "text-muted-foreground")}>
                        {names.length > 0 ? names.join(", ") : "Pick the apps"}
                    </span>
                    <ChevronDown className="size-4 opacity-50" />
                </Button>
            </PopoverTrigger>
            <PopoverContent
                className="w-[240px] p-1"
                align="start"
            >
                {apps.length === 0 && <p className="p-2 text-sm text-muted-foreground">No app in this env</p>}
                {apps.map(app => {
                    const checked = value.includes(app.id);
                    const inputId = `trigger-app-${app.id}`;
                    return (
                        <div
                            key={app.id}
                            className="flex items-center gap-2 rounded px-2 py-1.5 text-sm hover:bg-accent"
                        >
                            <Checkbox
                                id={inputId}
                                checked={checked}
                                onCheckedChange={next => {
                                    onChange(next === true ? [...value, app.id] : value.filter(id => id !== app.id));
                                }}
                            />
                            <label
                                htmlFor={inputId}
                                className="grow cursor-pointer truncate"
                            >
                                {app.name}
                            </label>
                        </div>
                    );
                })}
            </PopoverContent>
        </Popover>
    );
}

interface AppsPickerProps {
    apps: NamedApp[];
    value: string[];
    onChange: (value: string[]) => void;
    readOnly: boolean;
}

/**
 * The events that run a job, in a form whose values have `triggers`. An app's
 * job listens to its own app (no `apps`); an env's job picks the apps of the env
 * it listens to.
 */
export function JobTriggersField({ apps, readOnly = false }: Props) {
    const { control, setValue } = useFormContext<{ triggers: JobTriggerFormValue[] }>();
    const { fields, append, remove } = useFieldArray({ control, name: "triggers" });
    const triggers = useWatch({ control, name: "triggers" });
    const isEnvJob = apps !== undefined;

    return (
        <div className="flex w-full max-w-[760px] flex-col gap-3">
            {fields.length === 0 && (
                <p className="text-sm text-muted-foreground">
                    No trigger: the job runs on its schedule, by hand, or as a step of a sequence.
                </p>
            )}

            {fields.map((field, index) => {
                const trigger = triggers[index] ?? field;
                return (
                    <div
                        key={field.id}
                        className="flex flex-wrap items-center gap-2 rounded-md border px-3 py-2"
                    >
                        <Select
                            value={trigger.event}
                            onValueChange={value => {
                                setValue(`triggers.${index}.event`, value as ESchedJobTriggerEvent, {
                                    shouldDirty: true,
                                });
                            }}
                            disabled={readOnly}
                        >
                            <SelectTrigger className="h-9 w-[210px]">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                {SCHED_JOB_TRIGGER_EVENT_OPTIONS.map(option => (
                                    <SelectItem
                                        key={option.value}
                                        value={option.value}
                                    >
                                        {option.label}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>

                        {isEnvJob ? (
                            <>
                                <span className="text-sm text-muted-foreground">of</span>
                                <AppsPicker
                                    apps={apps}
                                    value={trigger.appIds}
                                    onChange={value => {
                                        setValue(`triggers.${index}.appIds`, value, { shouldDirty: true });
                                    }}
                                    readOnly={readOnly}
                                />
                            </>
                        ) : (
                            <span className="text-sm text-muted-foreground">of this app</span>
                        )}

                        {trigger.event === ESchedJobTriggerEvent.PreDeploy && (
                            <div className="flex items-center gap-2 text-sm">
                                <Checkbox
                                    id={`${field.id}-wait`}
                                    checked={trigger.wait}
                                    onCheckedChange={checked => {
                                        setValue(`triggers.${index}.wait`, checked === true, { shouldDirty: true });
                                    }}
                                    disabled={readOnly}
                                />
                                <label
                                    htmlFor={`${field.id}-wait`}
                                    className="cursor-pointer"
                                >
                                    Deploy waits for this job
                                </label>
                            </div>
                        )}

                        {!readOnly && (
                            <Button
                                type="button"
                                variant="ghost"
                                size="icon-sm"
                                aria-label="Remove trigger"
                                className="ml-auto"
                                onClick={() => {
                                    remove(index);
                                }}
                            >
                                <Trash2 className="size-4" />
                            </Button>
                        )}

                        {!isEnvJob && trigger.event === ESchedJobTriggerEvent.AppDisabled && (
                            <p className={cn(dashedBorderBox, "basis-full text-sm")}>
                                <span className="font-semibold text-orange-500">Note:</span> a disabled app has no
                                container to run this job in. An env job can run one in another app.
                            </p>
                        )}
                    </div>
                );
            })}

            {!readOnly && (
                <Button
                    type="button"
                    variant="outline"
                    className="w-fit"
                    disabled={fields.length >= JOB_MAX_TRIGGERS}
                    onClick={() => {
                        append(createDefaultJobTrigger());
                    }}
                >
                    <Plus className="size-4" /> Add trigger
                </Button>
            )}
        </div>
    );
}

interface Props {
    /** The env's apps, for an env's job; undefined for an app's job. */
    apps?: NamedApp[];
    readOnly?: boolean;
}

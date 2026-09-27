import { useMemo, useState } from "react";

import { ArrowDown, ArrowUp, Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext, useFormState } from "react-hook-form";

import {
    Button,
    FieldError,
    Input,
    Select,
    SelectContent,
    SelectGroup,
    SelectItem,
    SelectLabel,
    SelectTrigger,
    SelectValue,
} from "@/components/ui";

import { JOB_SEQUENCE_MAX_STEPS, type JobSequenceFormInput } from "./job-sequence.form.schema";

/** A job a step can run: one of the scope's jobs, never a sequence. */
export interface JobSequenceCandidate {
    id: string;
    name: string;
    /** The app the job belongs to; empty for the scope's own. */
    appName: string;
    disabled: boolean;
}

function groupByApp(candidates: JobSequenceCandidate[]): [string, JobSequenceCandidate[]][] {
    const groups = new Map<string, JobSequenceCandidate[]>();
    for (const candidate of candidates) {
        const group = groups.get(candidate.appName) ?? [];
        group.push(candidate);
        groups.set(candidate.appName, group);
    }

    return [...groups.entries()].sort(([a], [b]) => a.localeCompare(b));
}

/**
 * The steps of a sequence, in order: added from the scope's jobs, moved up and
 * down, each with an optional label. A job may be added more than once.
 */
export function JobSequenceStepsField({ candidates, isLoadingCandidates, readOnly }: Props) {
    const { control } = useFormContext<JobSequenceFormInput>();
    const { fields, append, remove, swap } = useFieldArray({ control, name: "steps" });
    const { errors } = useFormState({ control, name: "steps" });
    const [picked, setPicked] = useState("");

    const groups = useMemo(() => groupByApp(candidates), [candidates]);
    const canAdd = !readOnly && fields.length < JOB_SEQUENCE_MAX_STEPS;

    function addPicked() {
        const candidate = candidates.find(item => item.id === picked);
        if (!candidate) {
            return;
        }

        append({ jobId: candidate.id, jobName: candidate.name, appName: candidate.appName, name: "" });
        setPicked("");
    }

    return (
        <div className="flex w-full max-w-[720px] flex-col gap-3">
            {fields.length === 0 && (
                <p className="text-sm text-muted-foreground">No steps yet: add the jobs to run, in order.</p>
            )}

            {fields.length > 0 && (
                <ol className="flex flex-col gap-2">
                    {fields.map((field, index) => {
                        const stepError = errors.steps?.[index];

                        return (
                            <li
                                key={field.id}
                                className="flex flex-wrap items-center gap-2 rounded-md border px-3 py-2"
                            >
                                <span className="w-6 shrink-0 text-sm font-semibold tabular-nums">{index + 1}.</span>
                                <div className="flex min-w-[160px] grow flex-col">
                                    <span className="text-sm font-medium">{field.jobName || field.jobId}</span>
                                    {field.appName && (
                                        <span className="text-xs text-muted-foreground">{field.appName}</span>
                                    )}
                                </div>
                                <div className="flex flex-col gap-1">
                                    <Input
                                        {...control.register(`steps.${index}.name`)}
                                        placeholder="label (optional)"
                                        className="h-8 w-[200px] text-sm"
                                        aria-invalid={Boolean(stepError?.name)}
                                        disabled={readOnly}
                                    />
                                    <FieldError errors={[stepError?.name, stepError?.jobId]} />
                                </div>
                                {!readOnly && (
                                    <div className="flex items-center">
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon-sm"
                                            aria-label="Move up"
                                            disabled={index === 0}
                                            onClick={() => {
                                                swap(index, index - 1);
                                            }}
                                        >
                                            <ArrowUp className="size-4" />
                                        </Button>
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon-sm"
                                            aria-label="Move down"
                                            disabled={index === fields.length - 1}
                                            onClick={() => {
                                                swap(index, index + 1);
                                            }}
                                        >
                                            <ArrowDown className="size-4" />
                                        </Button>
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon-sm"
                                            aria-label="Remove step"
                                            onClick={() => {
                                                remove(index);
                                            }}
                                        >
                                            <Trash2 className="size-4" />
                                        </Button>
                                    </div>
                                )}
                            </li>
                        );
                    })}
                </ol>
            )}

            {canAdd && (
                <div className="flex flex-wrap items-center gap-2">
                    <Select
                        value={picked}
                        onValueChange={setPicked}
                        disabled={isLoadingCandidates || candidates.length === 0}
                    >
                        <SelectTrigger className="w-[320px]">
                            <SelectValue
                                placeholder={
                                    isLoadingCandidates
                                        ? "Loading jobs..."
                                        : candidates.length === 0
                                          ? "No job to add"
                                          : "Pick a job"
                                }
                            />
                        </SelectTrigger>
                        <SelectContent>
                            {groups.map(([appName, jobs]) => (
                                <SelectGroup key={appName || "-"}>
                                    <SelectLabel>{appName || "This scope"}</SelectLabel>
                                    {jobs.map(job => (
                                        <SelectItem
                                            key={job.id}
                                            value={job.id}
                                        >
                                            {job.name}
                                            {job.disabled && <span className="text-muted-foreground"> (disabled)</span>}
                                        </SelectItem>
                                    ))}
                                </SelectGroup>
                            ))}
                        </SelectContent>
                    </Select>
                    <Button
                        type="button"
                        variant="outline"
                        disabled={!picked}
                        onClick={addPicked}
                    >
                        <Plus className="size-4" /> Add step
                    </Button>
                </div>
            )}

            <FieldError errors={[errors.steps?.root, errors.steps]} />
        </div>
    );
}

interface Props {
    candidates: JobSequenceCandidate[];
    isLoadingCandidates: boolean;
    readOnly: boolean;
}

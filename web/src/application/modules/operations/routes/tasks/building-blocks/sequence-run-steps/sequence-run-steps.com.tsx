import { Badge } from "@components/ui/badge";
import { Popover, PopoverContent, PopoverTrigger } from "@components/ui/popover";
import { cn } from "@lib/utils";
import type { SystemTaskSequenceRun, SystemTaskSequenceStep } from "~/operations/domain";
import { SystemTaskSequenceStepStatus } from "~/operations/domain";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui";

const STATUS_CLASS: Record<SystemTaskSequenceStep["status"], string> = {
    [SystemTaskSequenceStepStatus.Pending]: "bg-muted text-muted-foreground",
    [SystemTaskSequenceStepStatus.Running]: "bg-purple-500 text-white",
    [SystemTaskSequenceStepStatus.Done]: "bg-green-600 text-white",
    [SystemTaskSequenceStepStatus.Failed]: "bg-destructive text-white",
    [SystemTaskSequenceStepStatus.Skipped]: "bg-amber-500 text-white",
};

function formatStepDuration(step: SystemTaskSequenceStep, now: Date): string {
    if (!step.startedAt) {
        return "-";
    }

    const seconds = Math.max(0, Math.round(((step.endedAt ?? now).getTime() - step.startedAt.getTime()) / 1000));
    if (seconds < 60) {
        return `${seconds}s`;
    }

    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) {
        return `${minutes}m ${seconds % 60}s`;
    }

    return `${Math.floor(minutes / 60)}h ${minutes % 60}m`;
}

function StepOutputs({ outputs }: { outputs: Record<string, string> }) {
    const keys = Object.keys(outputs).sort();
    if (keys.length === 0) {
        return <span className="text-muted-foreground">-</span>;
    }

    return (
        <div className="flex flex-wrap gap-1">
            {keys.map(key => (
                <Popover key={key}>
                    <PopoverTrigger asChild>
                        <button
                            type="button"
                            className="rounded border px-1.5 py-0.5 font-mono text-xs hover:bg-accent"
                        >
                            {key}
                        </button>
                    </PopoverTrigger>
                    <PopoverContent className="max-w-[480px]">
                        <pre className="whitespace-pre-wrap break-all font-mono text-xs">{outputs[key]}</pre>
                    </PopoverContent>
                </Popover>
            ))}
        </div>
    );
}

/**
 * The steps of a job sequence's run: how each went, and what it output. The
 * log below has the steps' output, each under its own header.
 */
export function SequenceRunSteps({ run, now, className }: Props) {
    if (run.steps.length === 0) {
        return null;
    }

    return (
        <div className={cn("flex min-w-0 flex-col gap-2", className)}>
            <h3 className="text-sm font-semibold">Steps</h3>
            <div className="overflow-x-auto rounded-md border">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-10">#</TableHead>
                            <TableHead>Step</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead className="text-right">Attempts</TableHead>
                            <TableHead className="text-right">Duration</TableHead>
                            <TableHead className="text-right">Exit Code</TableHead>
                            <TableHead>Error</TableHead>
                            <TableHead>Outputs</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {run.steps.map((step, index) => (
                            <TableRow key={`${index}-${step.job.id}`}>
                                <TableCell className="tabular-nums">{index + 1}</TableCell>
                                <TableCell className="font-medium">{step.name || step.job.id}</TableCell>
                                <TableCell>
                                    <Badge className={cn("capitalize", STATUS_CLASS[step.status])}>{step.status}</Badge>
                                </TableCell>
                                <TableCell className="text-right tabular-nums">{step.attempts || "-"}</TableCell>
                                <TableCell className="text-right tabular-nums">
                                    {formatStepDuration(step, now)}
                                </TableCell>
                                <TableCell className="text-right tabular-nums">{step.exitCode ?? "-"}</TableCell>
                                <TableCell className="max-w-[320px] whitespace-normal break-words text-sm">
                                    {step.error || "-"}
                                </TableCell>
                                <TableCell>
                                    <StepOutputs outputs={step.outputs} />
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </div>
        </div>
    );
}

interface Props {
    run: SystemTaskSequenceRun;
    now: Date;
    className?: string;
}

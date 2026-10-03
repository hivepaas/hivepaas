import { type PropsWithChildren, useEffect } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Controller, FormProvider, useForm, useWatch } from "react-hook-form";
import type { LoggingPerformance, LoggingPerformanceNode } from "~/system-settings/domain";
import { SectionHeader } from "~/system-settings/module-shared";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Checkbox, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

import { CAPACITY_LABEL, PerformanceNodeStatus, capacityText, formatNodeMemory } from "../building-blocks";
import {
    type LoggingPerformanceFormInput,
    type LoggingPerformanceFormOutput,
    LoggingPerformanceFormSchema,
} from "../schemas";

import { toLoggingPerformanceFormInput } from "./logging-performance.form-mappers";

/** The rows of one section, indented under its header like every settings page. */
function SectionBody({ children }: PropsWithChildren) {
    return <div className="flex flex-col gap-6 px-3">{children}</div>;
}

/** About how much of a node's memory OBI takes at the capacity chosen for it. */
function memoryShare(node: LoggingPerformanceNode, capacity: string, settings: LoggingPerformance): string | null {
    const effective = capacity === "auto" ? node.recommended : capacity;
    const info = settings.capacities.find(item => item.capacity === effective);
    if (!info || node.memoryBytes <= 0) {
        return null;
    }
    const share = (info.memoryMiB * 1024 * 1024 * 100) / node.memoryBytes;
    return share < 1 ? "under 1% of its memory" : `about ${Math.round(share)}% of its memory`;
}

function NodeRow({
    node,
    index,
    settings,
}: {
    node: LoggingPerformanceNode;
    index: number;
    settings: LoggingPerformance;
}) {
    const enabled = useWatch<LoggingPerformanceFormInput, `nodes.${number}.enabled`>({
        name: `nodes.${index}.enabled`,
    });
    const capacity = useWatch<LoggingPerformanceFormInput, `nodes.${number}.capacity`>({
        name: `nodes.${index}.capacity`,
    });
    const recommended = settings.capacities.find(item => item.capacity === node.recommended);
    const share = enabled ? memoryShare(node, capacity, settings) : null;

    return (
        <TableRow>
            <TableCell>
                <Controller<LoggingPerformanceFormInput, `nodes.${number}.enabled`>
                    name={`nodes.${index}.enabled`}
                    render={({ field }) => (
                        <Checkbox
                            aria-label={`Run OBI on ${node.hostname || node.id}`}
                            checked={field.value}
                            onCheckedChange={checked => {
                                field.onChange(checked === true);
                            }}
                        />
                    )}
                />
            </TableCell>
            <TableCell>
                <span className="flex flex-col gap-1">
                    <span className="font-medium">{node.hostname || node.id}</span>
                    <span className="flex flex-wrap gap-1">
                        {node.role && <Badge tone="neutral">{node.role}</Badge>}
                        {node.state && node.state !== "ready" && <Badge tone="red">{node.state}</Badge>}
                        {node.availability && node.availability !== "active" && (
                            <Badge tone="amber">{node.availability}</Badge>
                        )}
                    </span>
                </span>
            </TableCell>
            <TableCell className="whitespace-nowrap">
                <span className="flex flex-col gap-1">
                    <span className="tabular-nums">{formatNodeMemory(node.memoryBytes)}</span>
                    {share && <span className="text-xs text-muted-foreground">OBI: {share}</span>}
                </span>
            </TableCell>
            <TableCell className="min-w-[240px]">
                <Controller<LoggingPerformanceFormInput, `nodes.${number}.capacity`>
                    name={`nodes.${index}.capacity`}
                    render={({ field }) => (
                        <Select
                            value={field.value}
                            onValueChange={field.onChange}
                            disabled={!enabled}
                        >
                            <SelectTrigger className="w-full">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="auto">
                                    Recommended ({CAPACITY_LABEL[node.recommended]}, {capacityText(recommended)})
                                </SelectItem>
                                {settings.capacities.map(item => (
                                    <SelectItem
                                        key={item.capacity}
                                        value={item.capacity}
                                    >
                                        {CAPACITY_LABEL[item.capacity]} ({capacityText(item)})
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    )}
                />
            </TableCell>
            <TableCell className="min-w-[200px]">
                <PerformanceNodeStatus
                    node={node}
                    statusReason={settings.statusReason}
                />
            </TableCell>
        </TableRow>
    );
}

export function LoggingPerformanceForm({ settings, readOnly, onSubmit, children }: Props) {
    const methods = useForm<LoggingPerformanceFormInput, unknown, LoggingPerformanceFormOutput>({
        defaultValues: toLoggingPerformanceFormInput(settings),
        resolver: zodResolver(LoggingPerformanceFormSchema),
        mode: "onSubmit",
    });
    const { control, reset } = methods;

    // After a save the query refetches; show what the server now holds.
    useEffect(() => {
        reset(toLoggingPerformanceFormInput(settings));
    }, [reset, settings]);

    const enabled = useWatch({ control, name: "enabled" });

    return (
        <div className="pt-2">
            <FormProvider {...methods}>
                <form
                    onSubmit={event => {
                        event.preventDefault();
                        if (readOnly) {
                            return;
                        }
                        void methods.handleSubmit(onSubmit)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    <fieldset
                        disabled={readOnly}
                        className="flex flex-col gap-6 border-0 p-0 m-0 min-w-0"
                    >
                        <div className={cn(dashedBorderBox, "space-y-2")}>
                            <p>
                                <span className="font-semibold text-orange-500">Note:</span> Routes and calls are
                                measured inside each app&apos;s containers by OBI, an eBPF program HivePaaS runs on the
                                nodes chosen below: every request an app answers - from Traefik or from another app of
                                its project - by route, and every call it makes to a database, another app or an outside
                                host. The apps are not changed, and no request or query text is kept.
                            </p>
                            <p>
                                An app is measured once its Feature Settings ask for it, on the nodes here that run OBI.
                                OBI needs Linux 5.8 or later on a VM or a dedicated server, not a container-based VPS,
                                and takes memory on each node, idle or not.
                            </p>
                        </div>

                        <SectionHeader>General</SectionHeader>
                        <SectionBody>
                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Enabled"
                                        content="Run OBI on the nodes chosen below, for the apps that ask for their routes and calls. Each node's agent applies a change within 30 seconds."
                                    />
                                }
                            >
                                <Controller
                                    control={control}
                                    name="enabled"
                                    render={({ field }) => (
                                        <Checkbox
                                            checked={field.value}
                                            onCheckedChange={checked => {
                                                field.onChange(checked === true);
                                            }}
                                        />
                                    )}
                                />
                            </InfoBlock>
                        </SectionBody>

                        {enabled && (
                            <>
                                <SectionHeader>Nodes</SectionHeader>
                                <SectionBody>
                                    <p className="text-sm text-muted-foreground">
                                        A node&apos;s capacity is how many requests and connections OBI tracks there at
                                        once. More takes more memory, idle or not; too little loses what does not fit,
                                        silently. The recommendation follows the node&apos;s memory: Small under 8 GB,
                                        Medium under 32 GB, Large from 32 GB.
                                    </p>
                                    {settings.statusReason === "unreadable" && (
                                        <p className="text-sm text-muted-foreground">
                                            The stored logs could not be read: the nodes&apos; statuses are not shown.
                                        </p>
                                    )}
                                    <div className="overflow-x-auto rounded-md border">
                                        <Table>
                                            <TableHeader>
                                                <TableRow>
                                                    <TableHead className="w-12">Run</TableHead>
                                                    <TableHead>Node</TableHead>
                                                    <TableHead>Memory</TableHead>
                                                    <TableHead>Capacity</TableHead>
                                                    <TableHead>Status</TableHead>
                                                </TableRow>
                                            </TableHeader>
                                            <TableBody>
                                                {settings.nodes.map((node, index) => (
                                                    <NodeRow
                                                        key={node.id}
                                                        node={node}
                                                        index={index}
                                                        settings={settings}
                                                    />
                                                ))}
                                            </TableBody>
                                        </Table>
                                    </div>
                                </SectionBody>
                            </>
                        )}
                    </fieldset>
                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    settings: LoggingPerformance;
    readOnly: boolean;
    onSubmit: (values: LoggingPerformanceFormOutput) => void;
}>;

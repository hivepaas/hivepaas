import { type PropsWithChildren, useEffect } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Plus, Trash2 } from "lucide-react";
import { Controller, FormProvider, useFieldArray, useForm, useWatch } from "react-hook-form";
import { ClusterVolumesQueries } from "~/cluster/data/queries";
import type { HivePaaSLoggingSettings } from "~/system-settings/domain";
import { SectionHeader } from "~/system-settings/module-shared";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import {
    Button,
    Checkbox,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui";
import { InputNumber } from "@/components/ui/input-number";

import { EndpointFields, FieldMessage } from "../building-blocks";
import {
    type HivePaaSLoggingSettingsFormInput,
    type HivePaaSLoggingSettingsFormOutput,
    HivePaaSLoggingSettingsFormSchema,
} from "../schemas";

import { emptyLoggingEndpointForm, toLoggingFormInput } from "./hivepaas-logging-settings.form-mappers";

type SchemaInput = HivePaaSLoggingSettingsFormInput;
type SchemaOutput = HivePaaSLoggingSettingsFormOutput;

const LIST_ALL = { pagination: { page: 1, size: 100 } };

/** The rows of one section, indented under its header like every settings page. */
function SectionBody({ children }: PropsWithChildren) {
    return <div className="flex flex-col gap-6 px-3">{children}</div>;
}

export function HivePaaSLoggingSettingsForm({ settings, readOnly, onSubmit, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: toLoggingFormInput(settings),
        resolver: zodResolver(HivePaaSLoggingSettingsFormSchema),
        mode: "onSubmit",
    });
    const { control, register, reset } = methods;

    // After a save the query refetches; show what the server now holds.
    useEffect(() => {
        reset(toLoggingFormInput(settings));
    }, [reset, settings]);

    const forwards = useFieldArray({ control, name: "forwards" });
    const enabled = useWatch({ control, name: "enabled" });
    const backendManaged = useWatch({ control, name: "backendManaged" });
    const volumeId = useWatch({ control, name: "volumeId" });

    const volumes = ClusterVolumesQueries.useFindManyPaginated(LIST_ALL).data?.data ?? [];
    // The backend runs wherever its volume is, so a volume that names no node -
    // by id or by label - leaves that open. Not gated on the cluster having
    // more than one node: a warning about losing logs must not be suppressed by
    // whether some other query has loaded its rows yet.
    const selectedVolume = volumes.find(volume => volume.id === volumeId);
    const isVolumeUnpinned = selectedVolume !== undefined && !selectedVolume.nodeId && !selectedVolume.nodeLabel;

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
                        <div className={cn(dashedBorderBox)}>
                            <span className="font-semibold text-orange-500">Note:</span> Default Docker logs are stored
                            locally on each host, which risks disk exhaustion, loses history when containers are
                            removed, and lacks centralized search across nodes. You can use the configuration below to
                            collect, persist, and centralize logs across your cluster.
                        </div>

                        <SectionHeader>General</SectionHeader>
                        <SectionBody>
                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Enabled"
                                        content="Collect logs on every node and keep them after containers are gone."
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

                        {/* Nothing below applies while logging is off, the way every
                            other settings page hides its configuration. The schema
                            skips its checks too, so turning the feature off never
                            demands a complete form. */}
                        {enabled && (
                            <>
                                <SectionHeader>Sources</SectionHeader>
                                <SectionBody>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="Container logs"
                                                content="Every container on every node: your apps and HivePaaS's own services. They are written into one directory under container ids, so they are collected together - which line belongs to which app is decided when the logs are read."
                                            />
                                        }
                                    >
                                        <Controller
                                            control={control}
                                            name="sources.apps"
                                            render={({ field }) => (
                                                <Checkbox
                                                    checked={field.value}
                                                    onCheckedChange={checked => {
                                                        field.onChange(checked === true);
                                                    }}
                                                />
                                            )}
                                        />
                                        <FieldMessage name="sources.apps" />
                                    </InfoBlock>
                                </SectionBody>

                                <SectionHeader>Backend</SectionHeader>
                                <SectionBody>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="Where logs are stored"
                                                content="HivePaaS can run VictoriaLogs for you, or ship to a backend you run yourself."
                                            />
                                        }
                                    >
                                        <Controller
                                            control={control}
                                            name="backendManaged"
                                            render={({ field }) => (
                                                <Select
                                                    value={field.value ? "managed" : "external"}
                                                    onValueChange={value => {
                                                        field.onChange(value === "managed");
                                                    }}
                                                    disabled={readOnly}
                                                >
                                                    <SelectTrigger className="w-full max-w-[420px]">
                                                        <SelectValue />
                                                    </SelectTrigger>
                                                    <SelectContent>
                                                        <SelectItem value="managed">
                                                            Run VictoriaLogs in the cluster (managed)
                                                        </SelectItem>
                                                        <SelectItem value="external">Use my own endpoint</SelectItem>
                                                    </SelectContent>
                                                </Select>
                                            )}
                                        />
                                    </InfoBlock>

                                    {backendManaged ? (
                                        <>
                                            {isVolumeUnpinned && (
                                                <div className="rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-400">
                                                    <span className="font-medium">
                                                        This volume is not pinned to a node.
                                                    </span>{" "}
                                                    The logging backend runs wherever its volume is, so it is free to
                                                    start on any node in the cluster. Make sure every node can reach
                                                    this volume — a networked or shared-storage driver — or pin the
                                                    volume to a node. If the volume is local to one node, the backend
                                                    can come up elsewhere on an empty directory: the logs collected so
                                                    far stay behind on the old node and stop being shown.
                                                </div>
                                            )}
                                            <InfoBlock
                                                titleWidth={220}
                                                title={
                                                    <LabelWithInfo
                                                        label="Data volume"
                                                        content="Where the logs are stored, and what decides which node the backend runs on. It is kept when logging is turned off."
                                                        isRequired
                                                    />
                                                }
                                            >
                                                <Controller
                                                    control={control}
                                                    name="volumeId"
                                                    render={({ field }) => (
                                                        <Select
                                                            value={field.value || undefined}
                                                            onValueChange={field.onChange}
                                                            disabled={readOnly}
                                                        >
                                                            <SelectTrigger className="w-full max-w-[420px]">
                                                                <SelectValue placeholder="Select a volume" />
                                                            </SelectTrigger>
                                                            <SelectContent>
                                                                {volumes.map(volume => (
                                                                    <SelectItem
                                                                        key={volume.id}
                                                                        value={volume.id}
                                                                    >
                                                                        {volume.name}
                                                                    </SelectItem>
                                                                ))}
                                                            </SelectContent>
                                                        </Select>
                                                    )}
                                                />
                                                <FieldMessage name="volumeId" />
                                            </InfoBlock>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={
                                                    <LabelWithInfo
                                                        label="Data directory"
                                                        content="A directory inside the volume, so one volume can serve more than logging. Empty writes at the volume's root. Changing it on a running backend starts an empty store: the logs collected so far stay where they are, and stop being shown."
                                                    />
                                                }
                                            >
                                                <Input
                                                    {...register("volumeSubpath")}
                                                    placeholder="logs"
                                                    className="max-w-[420px]"
                                                />
                                                <FieldMessage name="volumeSubpath" />
                                            </InfoBlock>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={
                                                    <LabelWithInfo
                                                        label="Retention"
                                                        content="Such as 30d or 12h. VictoriaLogs keeps at least one day."
                                                    />
                                                }
                                            >
                                                <Input
                                                    {...register("retention")}
                                                    placeholder="30d"
                                                    className="max-w-[200px]"
                                                />
                                                <FieldMessage name="retention" />
                                            </InfoBlock>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={
                                                    <LabelWithInfo
                                                        label="Max disk usage %"
                                                        content="Drop the oldest days once the disk is this full. A plain number, not a percent sign. Empty means the retention period is the only limit."
                                                    />
                                                }
                                            >
                                                <Controller
                                                    control={control}
                                                    name="maxDiskUsagePercent"
                                                    render={({ field }) => (
                                                        <InputNumber
                                                            value={field.value ?? undefined}
                                                            placeholder="80"
                                                            min={1}
                                                            max={100}
                                                            showControls={false}
                                                            useGrouping={false}
                                                            className="max-w-[200px]"
                                                            onValueChange={value => {
                                                                field.onChange(
                                                                    typeof value === "number" ? value : null,
                                                                );
                                                            }}
                                                        />
                                                    )}
                                                />
                                                <FieldMessage name="maxDiskUsagePercent" />
                                            </InfoBlock>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={
                                                    <LabelWithInfo
                                                        label="CPU limit"
                                                        content="Cores the backend may use, as a number: 2, or 0.5 for half a core. Empty means no limit, and a heavy query can then take whatever the node has from the apps running beside it."
                                                    />
                                                }
                                            >
                                                <Controller
                                                    control={control}
                                                    name="cpuLimit"
                                                    render={({ field }) => (
                                                        <InputNumber
                                                            value={field.value ?? undefined}
                                                            placeholder="2"
                                                            min={0.25}
                                                            max={256}
                                                            step={0.25}
                                                            showControls={false}
                                                            useGrouping={false}
                                                            className="max-w-[200px]"
                                                            onValueChange={value => {
                                                                field.onChange(
                                                                    typeof value === "number" ? value : null,
                                                                );
                                                            }}
                                                        />
                                                    )}
                                                />
                                                <FieldMessage name="cpuLimit" />
                                            </InfoBlock>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={
                                                    <LabelWithInfo
                                                        label="Memory limit"
                                                        content="Memory the backend may use, written with its unit: 1gb, 512mb. VictoriaLogs sizes its caches from what it is allowed, so this is not only a ceiling - with no limit it sizes itself against the whole node. Set it too low and the container is killed mid-query rather than slowed down."
                                                    />
                                                }
                                            >
                                                <Input
                                                    {...register("memoryLimit")}
                                                    placeholder="1gb"
                                                    className="max-w-[200px]"
                                                />
                                                <FieldMessage name="memoryLimit" />
                                            </InfoBlock>
                                        </>
                                    ) : (
                                        <>
                                            <EndpointFields
                                                prefix="ingest"
                                                urlLabel="Ingest URL"
                                                urlInfo="Where the collector writes."
                                            />
                                            <EndpointFields
                                                prefix="query"
                                                urlLabel="Query URL"
                                                urlInfo="Without it, stored logs cannot be shown in HivePaaS."
                                            />
                                        </>
                                    )}
                                </SectionBody>

                                <SectionHeader>Forwards</SectionHeader>
                                <SectionBody>
                                    <p className="text-sm text-muted-foreground">
                                        A copy of every collected line, sent to a system HivePaaS does not run.
                                    </p>
                                    {forwards.fields.map((item, index) => (
                                        <div
                                            key={item.id}
                                            className="flex flex-col gap-6 rounded-lg border p-4"
                                        >
                                            <InfoBlock
                                                titleWidth={220}
                                                title={
                                                    <LabelWithInfo
                                                        label="Name"
                                                        content="How this destination is reported and removed. Must be unique."
                                                        isRequired
                                                    />
                                                }
                                            >
                                                <Input {...register(`forwards.${index}.name`)} />
                                                <FieldMessage name={`forwards.${index}.name`} />
                                            </InfoBlock>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={
                                                    <LabelWithInfo
                                                        label="Format"
                                                        content="jsonline for anything that accepts NDJSON; native for another VictoriaLogs."
                                                    />
                                                }
                                            >
                                                <Controller
                                                    control={control}
                                                    name={`forwards.${index}.format`}
                                                    render={({ field }) => (
                                                        <Select
                                                            value={field.value}
                                                            onValueChange={field.onChange}
                                                            disabled={readOnly}
                                                        >
                                                            <SelectTrigger className="w-[200px]">
                                                                <SelectValue />
                                                            </SelectTrigger>
                                                            <SelectContent>
                                                                <SelectItem value="jsonline">jsonline</SelectItem>
                                                                <SelectItem value="native">native</SelectItem>
                                                            </SelectContent>
                                                        </Select>
                                                    )}
                                                />
                                            </InfoBlock>
                                            <EndpointFields
                                                prefix={`forwards.${index}`}
                                                urlLabel="URL"
                                                urlInfo="An HTTP endpoint that accepts the format above."
                                            />
                                            <div>
                                                <Button
                                                    type="button"
                                                    variant="outline"
                                                    onClick={() => {
                                                        forwards.remove(index);
                                                    }}
                                                >
                                                    <Trash2 className="size-4" /> Remove
                                                </Button>
                                            </div>
                                        </div>
                                    ))}
                                    <div>
                                        <Button
                                            type="button"
                                            variant="outline"
                                            onClick={() => {
                                                forwards.append({
                                                    ...emptyLoggingEndpointForm,
                                                    name: "",
                                                    format: "jsonline",
                                                });
                                            }}
                                        >
                                            <Plus className="size-4" /> Add forward
                                        </Button>
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
    settings?: HivePaaSLoggingSettings;
    readOnly?: boolean;
    onSubmit: (values: SchemaOutput) => void;
}>;

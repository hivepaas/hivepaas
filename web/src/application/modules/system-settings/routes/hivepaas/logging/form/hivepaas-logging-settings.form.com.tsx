import { type PropsWithChildren, useEffect } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, FormProvider, useFieldArray, useForm, useWatch } from "react-hook-form";
import { ClusterVolumesQueries, NodesQueries } from "~/cluster/data/queries";
import type { HivePaaSLoggingSettings } from "~/system-settings/domain";

import { InfoBlock } from "@application/shared/components";

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

import { EndpointFields, FieldMessage, SectionTitle } from "../building-blocks";
import {
    type HivePaaSLoggingSettingsFormInput,
    type HivePaaSLoggingSettingsFormOutput,
    HivePaaSLoggingSettingsFormSchema,
} from "../schemas";

import { emptyLoggingEndpointForm, toLoggingFormInput } from "./hivepaas-logging-settings.form-mappers";

type SchemaInput = HivePaaSLoggingSettingsFormInput;
type SchemaOutput = HivePaaSLoggingSettingsFormOutput;

const LIST_ALL = { pagination: { page: 1, size: 100 } };

const SOURCES = [
    { name: "sources.apps", title: "Apps", description: "Every app's container output." },
    { name: "sources.hivepaas", title: "HivePaaS", description: "HivePaaS's own services." },
    { name: "sources.traefikAccess", title: "Traefik access log", description: "Requests through the proxy." },
] as const;

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
    const backendManaged = useWatch({ control, name: "backendManaged" });
    const nodeId = useWatch({ control, name: "nodeId" });

    const nodes = NodesQueries.useFindManyPaginated(LIST_ALL).data?.data ?? [];
    const volumes = (ClusterVolumesQueries.useFindManyPaginated(LIST_ALL).data?.data ?? []).filter(
        volume => !volume.nodeId || volume.nodeId === nodeId,
    );

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
                        <InfoBlock
                            titleWidth={220}
                            title="Enabled"
                            description="Collect logs on every node and keep them after containers are gone."
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

                        <section className="flex flex-col gap-4">
                            <SectionTitle>Sources</SectionTitle>
                            {SOURCES.map(source => (
                                <InfoBlock
                                    key={source.name}
                                    titleWidth={220}
                                    title={source.title}
                                    description={source.description}
                                >
                                    <Controller
                                        control={control}
                                        name={source.name}
                                        render={({ field }) => (
                                            <Checkbox
                                                checked={field.value}
                                                onCheckedChange={checked => {
                                                    field.onChange(checked === true);
                                                }}
                                            />
                                        )}
                                    />
                                    {source.name === "sources.apps" && <FieldMessage name="sources.apps" />}
                                </InfoBlock>
                            ))}
                            <InfoBlock
                                titleWidth={220}
                                title="Node logs"
                                description="Not collected yet: reserved for a later release."
                            >
                                <Checkbox
                                    checked={false}
                                    disabled
                                />
                            </InfoBlock>
                        </section>

                        <section className="flex flex-col gap-4">
                            <SectionTitle>Backend</SectionTitle>
                            <InfoBlock
                                titleWidth={220}
                                title="Where logs are stored"
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
                                    <InfoBlock
                                        titleWidth={220}
                                        title="Node"
                                        description="Where VictoriaLogs runs and keeps its data."
                                    >
                                        <Controller
                                            control={control}
                                            name="nodeId"
                                            render={({ field }) => (
                                                <Select
                                                    value={field.value || undefined}
                                                    onValueChange={field.onChange}
                                                    disabled={readOnly}
                                                >
                                                    <SelectTrigger className="w-full max-w-[420px]">
                                                        <SelectValue placeholder="Select a node" />
                                                    </SelectTrigger>
                                                    <SelectContent>
                                                        {nodes.map(node => (
                                                            <SelectItem
                                                                key={node.id}
                                                                value={node.id}
                                                            >
                                                                {node.name || node.hostname || node.refId}
                                                            </SelectItem>
                                                        ))}
                                                    </SelectContent>
                                                </Select>
                                            )}
                                        />
                                        <FieldMessage name="nodeId" />
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={220}
                                        title="Data volume"
                                        description="Only volumes reachable from the chosen node. It is kept when logging is turned off."
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
                                        title="Retention"
                                        description="Such as 30d or 12h. VictoriaLogs keeps at least one day."
                                    >
                                        <Input
                                            {...register("retention")}
                                            className="max-w-[200px]"
                                        />
                                        <FieldMessage name="retention" />
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={220}
                                        title="Max disk usage %"
                                        description="Drop the oldest days once the disk is this full. Optional."
                                    >
                                        <Controller
                                            control={control}
                                            name="maxDiskUsagePercent"
                                            render={({ field }) => (
                                                <InputNumber
                                                    value={field.value ?? undefined}
                                                    min={1}
                                                    max={100}
                                                    showControls={false}
                                                    useGrouping={false}
                                                    className="w-[120px]"
                                                    onValueChange={value => {
                                                        field.onChange(typeof value === "number" ? value : null);
                                                    }}
                                                />
                                            )}
                                        />
                                        <FieldMessage name="maxDiskUsagePercent" />
                                    </InfoBlock>
                                </>
                            ) : (
                                <>
                                    <EndpointFields
                                        prefix="ingest"
                                        urlTitle="Ingest URL"
                                        urlDescription="Where the collector writes."
                                    />
                                    <EndpointFields
                                        prefix="query"
                                        urlTitle="Query URL"
                                        urlDescription="Without it, stored logs cannot be shown in HivePaaS."
                                    />
                                </>
                            )}
                        </section>

                        <section className="flex flex-col gap-4">
                            <SectionTitle>Forwards</SectionTitle>
                            <p className="text-sm text-muted-foreground">
                                A copy of every collected line, sent to a system HivePaaS does not run.
                            </p>
                            {forwards.fields.map((item, index) => (
                                <div
                                    key={item.id}
                                    className="flex flex-col gap-4 rounded-lg border p-4"
                                >
                                    <InfoBlock
                                        titleWidth={220}
                                        title="Name"
                                    >
                                        <Input {...register(`forwards.${index}.name`)} />
                                        <FieldMessage name={`forwards.${index}.name`} />
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={220}
                                        title="Format"
                                        description="jsonline for anything that accepts NDJSON; native for another VictoriaLogs."
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
                                        urlTitle="URL"
                                        showBasicAuth={false}
                                    />
                                    <div>
                                        <Button
                                            type="button"
                                            variant="outline"
                                            onClick={() => {
                                                forwards.remove(index);
                                            }}
                                        >
                                            Remove forward
                                        </Button>
                                    </div>
                                </div>
                            ))}
                            <div>
                                <Button
                                    type="button"
                                    variant="outline"
                                    onClick={() => {
                                        forwards.append({ ...emptyLoggingEndpointForm, name: "", format: "jsonline" });
                                    }}
                                >
                                    Add forward
                                </Button>
                            </div>
                        </section>
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

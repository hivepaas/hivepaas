import { type PropsWithChildren, useEffect } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { Cloud, HardDrive } from "lucide-react";
import { Controller, FormProvider, useForm, useWatch } from "react-hook-form";
import { ClusterVolumesQueries } from "~/cluster/data/queries";
import { type OptionCard, OptionCardGroup } from "~/projects/module-shared/components/option-card-group";
import { CloudStorageQueries } from "~/settings/data/queries";
import type { HivePaaSRegistrySettings, HivePaaSRegistryStorageType } from "~/system-settings/domain";
import { SectionHeader } from "~/system-settings/module-shared";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Checkbox, Input, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui";
import { InputNumber } from "@/components/ui/input-number";

import {
    type HivePaaSRegistrySettingsFormInput,
    type HivePaaSRegistrySettingsFormOutput,
    HivePaaSRegistrySettingsFormSchema,
} from "../schemas";

import { describeCleanup, toRegistryFormInput } from "./hivepaas-registry-settings.form-mappers";

type SchemaInput = HivePaaSRegistrySettingsFormInput;
type SchemaOutput = HivePaaSRegistrySettingsFormOutput;

const LIST_ALL = { pagination: { page: 1, size: 100 } };

const STORAGE_OPTIONS: OptionCard<HivePaaSRegistryStorageType>[] = [
    {
        value: "volume",
        label: "Local volume",
        description:
            "Images are kept on one node's disk, and the registry runs on that node. A layer two apps share is stored once.",
        icon: HardDrive,
    },
    {
        value: "s3",
        label: "S3 storage",
        description:
            "Images are kept in a bucket, so the registry can run on any node. A layer two apps share is stored twice.",
        icon: Cloud,
    },
];

interface Props extends PropsWithChildren {
    settings?: HivePaaSRegistrySettings;
    readOnly?: boolean;
    onSubmit: (values: SchemaOutput) => void;
}

/** The rows of one section, indented under its header like every settings page. */
function SectionBody({ children }: PropsWithChildren) {
    return <div className="flex flex-col gap-6 px-3">{children}</div>;
}

export function HivePaaSRegistrySettingsForm({ settings, readOnly = false, onSubmit, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: toRegistryFormInput(settings),
        resolver: zodResolver(HivePaaSRegistrySettingsFormSchema),
        mode: "onSubmit",
    });
    const { control, register, reset } = methods;

    // After a save the query refetches; show what the server now holds.
    useEffect(() => {
        reset(toRegistryFormInput(settings));
    }, [reset, settings]);

    const enabled = useWatch({ control, name: "enabled" });
    const storageType = useWatch({ control, name: "storageType" });
    const cleanupEnabled = useWatch({ control, name: "cleanupEnabled" });
    const keepLast = useWatch({ control, name: "keepLast" });
    const keepDays = useWatch({ control, name: "keepDays" });

    // Only volumes that are shared with apps reach the registry: its app lives in
    // a project of its own, and a volume that is not inheritable is invisible
    // there. Offering the rest would mean offering a choice the server refuses.
    const volumes = (ClusterVolumesQueries.useFindManyPaginated(LIST_ALL).data?.data ?? []).filter(
        volume => volume.inheritable === true,
    );
    const cloudStorages = CloudStorageQueries.useFindManyPaginated(LIST_ALL).data?.data ?? [];

    // Once the app exists its images are on whatever was chosen, and nothing
    // copies them anywhere: the server refuses a change, so the cards say so
    // rather than accepting a click that will be rejected.
    const isProvisioned = settings?.registryStatus?.provisioned === true;

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
                        <SectionHeader>General</SectionHeader>
                        <SectionBody>
                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Enabled"
                                        content="Run a registry in the cluster, so an image built on one node can be pulled on every other one."
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

                        {/* Nothing below applies while the registry is off, the way
                            every other settings page hides its configuration. The
                            schema skips its checks too, so turning it off never
                            demands a complete form. */}
                        {enabled && (
                            <>
                                <SectionHeader>Address</SectionHeader>
                                <SectionBody>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="Domain"
                                                content="Images are named after it: <domain>/hivepaas/<app>:<commit>. Point it at this cluster before saving."
                                            />
                                        }
                                    >
                                        <div className="flex w-full max-w-[520px] flex-col gap-1">
                                            <Input
                                                {...register("domain")}
                                                placeholder="registry.example.com"
                                                disabled={readOnly || isProvisioned}
                                            />
                                            <p className="text-xs text-muted-foreground">
                                                Set this record to <span className="font-medium">DNS-only</span>.
                                                Through a proxy such as Cloudflare every layer leaves the cluster and
                                                comes back, and meets an upload limit on the way — 100 MB per request on
                                                the free plan, which a single layer passes without trying.
                                            </p>
                                            {isProvisioned && (
                                                <p className="text-xs text-muted-foreground">
                                                    The registry is running. Change its address in the app&apos;s
                                                    routing settings, where the domain is verified and its certificate
                                                    is obtained.
                                                </p>
                                            )}
                                        </div>
                                    </InfoBlock>
                                </SectionBody>

                                <SectionHeader>Storage</SectionHeader>
                                <SectionBody>
                                    <Controller
                                        control={control}
                                        name="storageType"
                                        render={({ field }) => (
                                            <OptionCardGroup
                                                options={STORAGE_OPTIONS.map(option => ({
                                                    ...option,
                                                    disabled: isProvisioned,
                                                }))}
                                                value={field.value}
                                                onChange={field.onChange}
                                                readOnly={readOnly}
                                                className="grid-cols-1 sm:grid-cols-2 max-w-[640px]"
                                            />
                                        )}
                                    />
                                    {isProvisioned && (
                                        <p className="text-xs text-muted-foreground">
                                            The storage cannot be changed once the registry holds images: nothing copies
                                            them from one to the other. Provision a new registry instead.
                                        </p>
                                    )}

                                    {storageType === "volume" ? (
                                        <InfoBlock
                                            titleWidth={220}
                                            title={
                                                <LabelWithInfo
                                                    label="Volume"
                                                    content="The images live here, and the registry runs on whatever node this volume is on."
                                                />
                                            }
                                        >
                                            <div className="flex w-full max-w-[420px] flex-col gap-1">
                                                <Controller
                                                    control={control}
                                                    name="volumeId"
                                                    render={({ field }) => (
                                                        <Select
                                                            value={field.value}
                                                            onValueChange={field.onChange}
                                                            disabled={readOnly || isProvisioned}
                                                        >
                                                            <SelectTrigger className="w-full">
                                                                <SelectValue placeholder="Choose a volume" />
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
                                                {volumes.length === 0 && (
                                                    <p className="text-xs text-muted-foreground">
                                                        No volume here is shared with apps yet. Create one in Cluster
                                                        &rsaquo; Volumes, or edit an existing one and make it available
                                                        to apps.
                                                    </p>
                                                )}
                                            </div>
                                        </InfoBlock>
                                    ) : (
                                        <InfoBlock
                                            titleWidth={220}
                                            title={
                                                <LabelWithInfo
                                                    label="Cloud storage"
                                                    content="The bucket, its region and its credentials come from this setting."
                                                />
                                            }
                                        >
                                            <Controller
                                                control={control}
                                                name="cloudStorageId"
                                                render={({ field }) => (
                                                    <Select
                                                        value={field.value}
                                                        onValueChange={field.onChange}
                                                        disabled={readOnly || isProvisioned}
                                                    >
                                                        <SelectTrigger className="w-full max-w-[420px]">
                                                            <SelectValue placeholder="Choose a cloud storage" />
                                                        </SelectTrigger>
                                                        <SelectContent>
                                                            {cloudStorages.map(storage => (
                                                                <SelectItem
                                                                    key={storage.id}
                                                                    value={storage.id}
                                                                >
                                                                    {storage.name}
                                                                </SelectItem>
                                                            ))}
                                                        </SelectContent>
                                                    </Select>
                                                )}
                                            />
                                        </InfoBlock>
                                    )}

                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="Memory limit"
                                                content="What the registry may use. At least 256mb."
                                            />
                                        }
                                    >
                                        <Input
                                            {...register("memoryLimit")}
                                            placeholder="512mb"
                                            className="w-full max-w-[200px]"
                                        />
                                    </InfoBlock>
                                </SectionBody>

                                <SectionHeader>Cleanup</SectionHeader>
                                <SectionBody>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="Remove old images"
                                                content="One tag is pushed per build, so a registry nobody prunes grows until the disk does."
                                            />
                                        }
                                    >
                                        <Controller
                                            control={control}
                                            name="cleanupEnabled"
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

                                    {cleanupEnabled && (
                                        <>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={
                                                    <LabelWithInfo
                                                        label="Builds to keep"
                                                        content="The newest builds of every app are kept however old they are, which is what protects an app nobody has deployed for a while."
                                                    />
                                                }
                                            >
                                                <Controller
                                                    control={control}
                                                    name="keepLast"
                                                    render={({ field }) => (
                                                        <InputNumber
                                                            value={field.value}
                                                            onChange={field.onChange}
                                                            min={1}
                                                            max={1000}
                                                            className="w-full max-w-[200px]"
                                                        />
                                                    )}
                                                />
                                            </InfoBlock>

                                            <InfoBlock
                                                titleWidth={220}
                                                title={
                                                    <LabelWithInfo
                                                        label="Days to keep"
                                                        content="Everything pushed, or pulled by a node, within this many days is kept as well."
                                                    />
                                                }
                                            >
                                                <Controller
                                                    control={control}
                                                    name="keepDays"
                                                    render={({ field }) => (
                                                        <InputNumber
                                                            value={field.value}
                                                            onChange={field.onChange}
                                                            min={1}
                                                            max={3650}
                                                            className="w-full max-w-[200px]"
                                                        />
                                                    )}
                                                />
                                            </InfoBlock>
                                        </>
                                    )}

                                    {/* Two numbers in a form are not a policy anybody can picture. */}
                                    <p className="px-1 text-sm text-muted-foreground">
                                        {describeCleanup(cleanupEnabled, keepLast, keepDays)}
                                    </p>
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

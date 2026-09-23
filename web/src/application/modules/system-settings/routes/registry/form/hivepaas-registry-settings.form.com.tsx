import { type PropsWithChildren, useEffect, useRef } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Cloud, Globe, HardDrive } from "lucide-react";
import { Controller, FormProvider, useForm, useWatch } from "react-hook-form";
import { ClusterVolumesQueries } from "~/cluster/data/queries";
import { type OptionCard, OptionCardGroup } from "~/projects/module-shared/components/option-card-group";
import { CloudStorageQueries } from "~/settings/data/queries";
import type { HivePaaSRegistrySettings, HivePaaSRegistryStorageType } from "~/system-settings/domain";
import { SectionHeader } from "~/system-settings/module-shared";

import { AppLink, Combobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { getDefaultDomain } from "@application/shared/utils/domain";

import { Button, Checkbox, Input } from "@/components/ui";
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
        // Two lines is all a card shows, so each says the one thing that decides it.
        description: "Kept on one node's disk, and the registry runs there. Shared layers are stored once.",
        icon: HardDrive,
    },
    {
        value: "s3",
        label: "S3 storage",
        description: "Kept in a bucket, so the registry can run on any node. Shared layers are stored twice.",
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
    const { control, register, reset, setValue } = methods;

    const domainInputRef = useRef<HTMLInputElement>(null);
    const domainRegister = register("domain");

    const handleSuggestDomain = () => {
        const suggested = getDefaultDomain("registry");
        if (suggested) {
            setValue("domain", suggested, {
                shouldValidate: true,
                shouldDirty: true,
            });
            const selectFirstSegment = () => {
                const input = domainInputRef.current;
                if (!input) return;
                input.focus();
                const dotIndex = suggested.indexOf(".");
                if (dotIndex > 0) {
                    input.setSelectionRange(0, dotIndex);
                } else {
                    input.select();
                }
            };
            if (domainInputRef.current) {
                domainInputRef.current.value = suggested;
                selectFirstSegment();
            }
            requestAnimationFrame(() => {
                selectFirstSegment();
            });
            setTimeout(() => {
                selectFirstSegment();
            }, 0);
        }
    };

    // After a save the query refetches; show what the server now holds.
    useEffect(() => {
        reset(toRegistryFormInput(settings));
    }, [reset, settings]);

    const enabled = useWatch({ control, name: "enabled" });
    const storageType = useWatch({ control, name: "storageType" });
    const cleanupEnabled = useWatch({ control, name: "cleanupEnabled" });
    const dashboardEnabled = useWatch({ control, name: "dashboardEnabled" });
    const domain = useWatch({ control, name: "domain" });
    const keepLast = useWatch({ control, name: "keepLast" });
    const keepDays = useWatch({ control, name: "keepDays" });

    // Only volumes that are shared with apps reach the registry: its app lives in
    // a project of its own, and a volume that is not inheritable is invisible
    // there. Offering the rest would mean offering a choice the server refuses.
    const volumesQuery = ClusterVolumesQueries.useFindManyPaginated(LIST_ALL);
    const cloudStoragesQuery = CloudStorageQueries.useFindManyPaginated(LIST_ALL);

    const volumes = (volumesQuery.data?.data ?? []).filter(volume => volume.inheritable === true);
    const cloudStorages = cloudStoragesQuery.data?.data ?? [];

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
                        <div className={cn(dashedBorderBox)}>
                            <span className="font-semibold text-orange-500">Note:</span> On a cluster with more than one
                            node, an image built on one node has to be pushed somewhere every other node can pull it
                            from. This registry is that place, running inside the cluster, with old images cleaned up
                            for you. On a single node it is not needed: the image is built where it runs, and nothing
                            has to be pulled.
                        </div>

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
                                            <div className="flex items-center gap-2 w-full">
                                                <div className="flex-1 min-w-0">
                                                    <Input
                                                        {...domainRegister}
                                                        ref={e => {
                                                            domainRegister.ref(e);
                                                            domainInputRef.current = e;
                                                        }}
                                                        placeholder="registry.example.com"
                                                        disabled={readOnly || isProvisioned}
                                                    />
                                                </div>
                                                <Button
                                                    type="button"
                                                    variant="outline"
                                                    size="sm"
                                                    title="Suggest domain based on current hostname"
                                                    disabled={readOnly || isProvisioned}
                                                    onClick={handleSuggestDomain}
                                                    className="h-9 px-3 text-xs font-medium border-border/80 hover:bg-muted/80 text-muted-foreground hover:text-foreground shrink-0"
                                                >
                                                    <Globe className="size-3.5 mr-1 text-emerald-500" />
                                                    Suggest Domain
                                                </Button>
                                            </div>
                                            <p className="text-xs text-muted-foreground">
                                                Set this record to <span className="font-medium">DNS-only</span>.
                                                Through a proxy such as Cloudflare every layer leaves the cluster and
                                                comes back, and meets an upload limit on the way — 100 MB per request on
                                                the free plan, which a single layer passes without trying.
                                            </p>
                                        </div>
                                    </InfoBlock>

                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="Web interface"
                                                content="Zot ships a page for browsing what the registry holds. It is served at the registry's domain, so it needs no address of its own."
                                            />
                                        }
                                    >
                                        <div className="flex w-full max-w-[520px] flex-col gap-1">
                                            <Controller
                                                control={control}
                                                name="dashboardEnabled"
                                                render={({ field }) => (
                                                    <Checkbox
                                                        checked={field.value}
                                                        disabled={readOnly}
                                                        onCheckedChange={checked => {
                                                            field.onChange(checked === true);
                                                        }}
                                                    />
                                                )}
                                            />
                                            <p className="text-xs text-muted-foreground">
                                                Signing in uses the registry account HivePaaS pushes with, which may
                                                delete every image it can read. Anyone who can reach the domain can load
                                                the page, so leave this off unless you want it browsable.
                                            </p>
                                            {dashboardEnabled && domain && (
                                                <p className="text-xs text-muted-foreground">
                                                    It answers at{" "}
                                                    <a
                                                        href={`https://${domain}/`}
                                                        className="text-link underline-offset-4 hover:underline"
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                    >
                                                        https://{domain}/
                                                    </a>{" "}
                                                    once the registry has restarted with this setting.
                                                </p>
                                            )}
                                        </div>
                                    </InfoBlock>
                                </SectionBody>

                                <SectionHeader>Storage</SectionHeader>
                                <SectionBody>
                                    <InfoBlock
                                        titleWidth={220}
                                        title={
                                            <LabelWithInfo
                                                label="Type"
                                                content="Where the images are kept. It is decided when the registry is provisioned, because nothing copies them from one store to the other."
                                            />
                                        }
                                    >
                                        <div className="flex w-full flex-col gap-2">
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
                                                    The storage cannot be changed once the registry holds images:
                                                    nothing copies them from one to the other. Provision a new registry
                                                    instead.
                                                </p>
                                            )}
                                        </div>
                                    </InfoBlock>

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
                                                        <Combobox
                                                            options={volumes.map(volume => ({
                                                                value: { id: volume.id },
                                                                label: volume.name,
                                                            }))}
                                                            value={field.value || null}
                                                            onChange={value => {
                                                                field.onChange(value ?? "");
                                                            }}
                                                            placeholder="Choose a volume"
                                                            emptyText="No volume is shared with apps"
                                                            className="w-full"
                                                            valueKey="id"
                                                            closeOnSelect
                                                            // A volume created in another tab shows up here
                                                            // without leaving the form half filled in.
                                                            loading={volumesQuery.isFetching}
                                                            onRefresh={() => void volumesQuery.refetch()}
                                                            isRefreshing={volumesQuery.isRefetching}
                                                            disabled={readOnly || isProvisioned}
                                                        />
                                                    )}
                                                />
                                                {volumes.length === 0 && !volumesQuery.isFetching && (
                                                    <p className="text-xs text-muted-foreground">
                                                        No volume here is shared with apps yet. Create one in{" "}
                                                        <AppLink.Basic
                                                            to={ROUTE.cluster.volumes.$route}
                                                            className="text-link underline-offset-4 hover:underline"
                                                            target="_blank"
                                                            rel="noopener noreferrer"
                                                        >
                                                            Cluster &rsaquo; Volumes
                                                        </AppLink.Basic>
                                                        , or edit an existing one and make it available to apps.
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
                                                    <Combobox
                                                        options={cloudStorages.map(storage => ({
                                                            value: { id: storage.id },
                                                            label: storage.name,
                                                        }))}
                                                        value={field.value || null}
                                                        onChange={value => {
                                                            field.onChange(value ?? "");
                                                        }}
                                                        placeholder="Choose a cloud storage"
                                                        emptyText="No cloud storage is configured"
                                                        className="w-full max-w-[420px]"
                                                        valueKey="id"
                                                        closeOnSelect
                                                        loading={cloudStoragesQuery.isFetching}
                                                        onRefresh={() => void cloudStoragesQuery.refetch()}
                                                        isRefreshing={cloudStoragesQuery.isRefetching}
                                                        disabled={readOnly || isProvisioned}
                                                    />
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
                                                        content="This many newest builds are kept for each environment of each app, however old they are, which is what protects an app nobody has deployed for a while. It is a floor, not a limit: nothing is removed while the window below still keeps it."
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
                                                        content="Nothing pushed within this many days is removed, however many builds there are, so a rollback can reach any of them. Raising it costs disk; the count above is what is left once it expires."
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

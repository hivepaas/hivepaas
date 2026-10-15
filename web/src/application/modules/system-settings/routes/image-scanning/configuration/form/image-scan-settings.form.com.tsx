import type { PropsWithChildren } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Controller, FormProvider, useForm, useWatch } from "react-hook-form";
import { ClusterVolumesQueries } from "~/cluster/data/queries";
import type { ClusterVolume } from "~/cluster/domain";
import { useNotificationSettingsSources } from "~/settings/module-shared/hooks";
import type { SystemImageScanSettings } from "~/system-settings/domain";
import { SectionHeader } from "~/system-settings/module-shared";

import {
    AppLink,
    BrowserTimezoneNote,
    Combobox,
    InfoBlock,
    LabelWithInfo,
    NextRunsField,
} from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { NotificationSettings } from "@application/shared/form";

import {
    Checkbox,
    Field,
    FieldError,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { InputNumber } from "@/components/ui/input-number";

import {
    IMAGE_SCAN_INTERVALS,
    type ImageScanSettingsFormInput,
    type ImageScanSettingsFormOutput,
    ImageScanSettingsFormSchema,
} from "../schemas";

import { toImageScanFormInput } from "./image-scan-settings.form-mappers";

type SchemaInput = ImageScanSettingsFormInput;
type SchemaOutput = ImageScanSettingsFormOutput;

const TITLE_WIDTH = 220;

const LIST_ALL = { pagination: { page: 1, size: 1000 } };

/** Where scans run, by the volume's pin. */
function whereScansRun(volume: ClusterVolume): string {
    if (volume.nodeId) {
        return "Every scan runs on the node this volume is on. An image that node does not have is read from its registry.";
    }
    if (volume.nodeLabel) {
        return `Scans run on a node carrying the label ${volume.nodeLabel}: the one running the image when there is one. Each of those nodes keeps its own copy of the DBs.`;
    }
    return "Each scan runs on the node running its image, reading it from that node's Docker. Every node must reach this volume, as a networked or shared one does.";
}

function SectionBody({ children }: PropsWithChildren) {
    return <div className="flex flex-col gap-6 px-3">{children}</div>;
}

const INTERVAL_LABELS: Record<string, string> = {
    "6h": "Every 6 hours",
    "12h": "Every 12 hours",
    "1d": "Every day",
    "2d": "Every 2 days",
    "7d": "Every week",
};

/**
 * Image scanning: the volume Trivy keeps its data on, which decides where scans
 * run; what each scan's container runs under; where the vulnerability DBs come
 * from; and the daily scan.
 */
export function ImageScanSettingsForm({ settings, readOnly = false, onSubmit, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: toImageScanFormInput(settings),
        resolver: zodResolver(ImageScanSettingsFormSchema),
        mode: "onSubmit",
    });
    const { control, register, formState } = methods;
    const enabled = useWatch({ control, name: "enabled" });
    const scheduleInterval = useWatch({ control, name: "scheduleInterval" });
    const volumeId = useWatch({ control, name: "volumeId" });

    // A Tmpfs volume is memory, emptied when each scan's container stops: the
    // server refuses it.
    const volumesQuery = ClusterVolumesQueries.useFindManyPaginated(LIST_ALL);
    const volumes = (volumesQuery.data?.data ?? []).filter(volume => !volume.tmpfsOptions);
    const selectedVolume = volumes.find(volume => volume.id === volumeId);
    const { sources, manageLink } = useNotificationSettingsSources({ type: "settings" });

    // An interval set through the API that is none of these is still shown.
    const intervals: readonly string[] = IMAGE_SCAN_INTERVALS.some(value => value === scheduleInterval)
        ? IMAGE_SCAN_INTERVALS
        : [scheduleInterval, ...IMAGE_SCAN_INTERVALS];

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
                            titleWidth={TITLE_WIDTH}
                            title="Enabled"
                        >
                            <Controller
                                control={control}
                                name="enabled"
                                render={({ field }) => (
                                    <Checkbox
                                        aria-label="Enabled"
                                        checked={field.value}
                                        onCheckedChange={checked => {
                                            field.onChange(checked === true);
                                        }}
                                    />
                                )}
                            />
                        </InfoBlock>

                        <div className={cn(dashedBorderBox)}>
                            <span className="font-semibold text-orange-500">Note:</span>{" "}
                            {enabled ? (
                                <span>
                                    The images apps run are scanned after each deployment, every day, and when asked.
                                    Each scan is a container of Trivy&apos;s image, started for it and removed after:
                                    nothing runs between scans. Trivy keeps its vulnerability DB (about 1.5 GB), its
                                    Java DB (about 1.4 GB, for images with jars) and what it analyzed on the volume.
                                </span>
                            ) : (
                                <span>
                                    Switched off, nothing is scanned. What the scans found stays on the apps&apos;
                                    Security tabs, and the volume keeps Trivy&apos;s data.
                                </span>
                            )}
                        </div>

                        {enabled && (
                            <>
                                <SectionHeader>Storage</SectionHeader>
                                <SectionBody>
                                    <InfoBlock
                                        titleWidth={TITLE_WIDTH}
                                        title={
                                            <LabelWithInfo
                                                label="Data volume"
                                                content="Where Trivy keeps its DBs and the layers it analyzed, and what decides where scans run. It can be changed at any time: the next scan downloads the DBs to the new one."
                                                isRequired
                                            />
                                        }
                                    >
                                        <Field>
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
                                                        placeholder="Select a volume"
                                                        emptyText="No volumes available"
                                                        className="w-full max-w-[420px]"
                                                        valueKey="id"
                                                        closeOnSelect
                                                        loading={volumesQuery.isFetching}
                                                        onRefresh={() => void volumesQuery.refetch()}
                                                        isRefreshing={volumesQuery.isRefetching}
                                                        disabled={readOnly}
                                                    />
                                                )}
                                            />
                                            <FieldError errors={[formState.errors.volumeId]} />
                                            {selectedVolume && (
                                                <p
                                                    className="text-xs text-muted-foreground"
                                                    data-testid="image-scan-where"
                                                >
                                                    {whereScansRun(selectedVolume)}
                                                </p>
                                            )}
                                            <p className="text-xs text-muted-foreground">
                                                {volumes.length === 0 && !volumesQuery.isFetching
                                                    ? "No volume yet. Create one in "
                                                    : "Volumes are created and managed in "}
                                                <AppLink.Basic
                                                    to={ROUTE.cluster.volumes.$route}
                                                    className="text-link underline-offset-4 hover:underline"
                                                    target="_blank"
                                                    rel="noopener noreferrer"
                                                >
                                                    Cluster &rsaquo; Volumes
                                                </AppLink.Basic>
                                                . A Tmpfs one is memory and is not offered.
                                            </p>
                                        </Field>
                                    </InfoBlock>
                                </SectionBody>

                                <SectionHeader>Scans</SectionHeader>
                                <SectionBody>
                                    <InfoBlock
                                        titleWidth={TITLE_WIDTH}
                                        title={
                                            <LabelWithInfo
                                                label="CPU limit"
                                                content="CPU cores one scan's container may use. Scans on one volume run one at a time."
                                            />
                                        }
                                    >
                                        <Controller
                                            control={control}
                                            name="cpuLimit"
                                            render={({ field }) => (
                                                <InputNumber
                                                    value={field.value ?? undefined}
                                                    placeholder="1"
                                                    min={0.25}
                                                    max={256}
                                                    step={0.25}
                                                    showControls={false}
                                                    useGrouping={false}
                                                    className="w-full max-w-[200px]"
                                                    onValueChange={value => {
                                                        field.onChange(typeof value === "number" ? value : null);
                                                    }}
                                                />
                                            )}
                                        />
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={TITLE_WIDTH}
                                        title={
                                            <LabelWithInfo
                                                label="Memory limit"
                                                content="What one scan's container may use, the DB's download included. At least 256mb: a cold scan of a large image was killed at 128mb."
                                            />
                                        }
                                    >
                                        <Field>
                                            <Input
                                                {...register("memoryLimit")}
                                                placeholder="512mb"
                                                className="w-full max-w-[200px]"
                                                aria-label="Memory limit"
                                            />
                                            <FieldError errors={[formState.errors.memoryLimit]} />
                                        </Field>
                                    </InfoBlock>
                                </SectionBody>

                                <SectionHeader>Vulnerability DB</SectionHeader>
                                <SectionBody>
                                    <InfoBlock
                                        titleWidth={TITLE_WIDTH}
                                        title={
                                            <LabelWithInfo
                                                label="DB repository"
                                                content="Where Trivy fetches its vulnerability DB from: empty for Trivy's own. An installation without the internet pushes the DB to a registry it reaches and names it here."
                                            />
                                        }
                                    >
                                        <Field>
                                            <Input
                                                {...register("dbRepository")}
                                                placeholder="mirror.gcr.io/aquasec/trivy-db:2"
                                                className="w-full max-w-[480px] font-mono text-xs"
                                                aria-label="DB repository"
                                            />
                                            <FieldError errors={[formState.errors.dbRepository]} />
                                        </Field>
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={TITLE_WIDTH}
                                        title={
                                            <LabelWithInfo
                                                label="Java DB repository"
                                                content="Where scans fetch Trivy's Java DB from, for images with jars: empty for Trivy's own."
                                            />
                                        }
                                    >
                                        <Field>
                                            <Input
                                                {...register("javaDbRepository")}
                                                placeholder="mirror.gcr.io/aquasec/trivy-java-db:1"
                                                className="w-full max-w-[480px] font-mono text-xs"
                                                aria-label="Java DB repository"
                                            />
                                            <FieldError errors={[formState.errors.javaDbRepository]} />
                                        </Field>
                                    </InfoBlock>
                                </SectionBody>

                                <SectionHeader>Daily scan</SectionHeader>
                                <SectionBody>
                                    <div className={cn(dashedBorderBox)}>
                                        Scans again what every app runs: the DB knows today what it did not yesterday.
                                        An image scanned within the hour is left as it is, and the results of an image
                                        no app has run for 30 days are removed.
                                    </div>
                                    <InfoBlock
                                        titleWidth={TITLE_WIDTH}
                                        title="Interval"
                                    >
                                        <Controller
                                            control={control}
                                            name="scheduleInterval"
                                            render={({ field }) => (
                                                <Select
                                                    value={field.value}
                                                    onValueChange={field.onChange}
                                                >
                                                    <SelectTrigger
                                                        className="w-full max-w-[400px]"
                                                        aria-label="Interval"
                                                    >
                                                        <SelectValue />
                                                    </SelectTrigger>
                                                    <SelectContent>
                                                        {intervals.map(value => (
                                                            <SelectItem
                                                                key={value}
                                                                value={value}
                                                            >
                                                                {INTERVAL_LABELS[value] ?? `Every ${value}`}
                                                            </SelectItem>
                                                        ))}
                                                    </SelectContent>
                                                </Select>
                                            )}
                                        />
                                    </InfoBlock>
                                    <InfoBlock
                                        titleWidth={TITLE_WIDTH}
                                        title="Schedule From"
                                    >
                                        <Controller
                                            control={control}
                                            name="scheduleFrom"
                                            render={({ field }) => (
                                                <Field>
                                                    <DateTimePicker
                                                        value={field.value ?? undefined}
                                                        onChange={date => {
                                                            field.onChange(date ?? null);
                                                        }}
                                                        placeholder="select date time"
                                                        granularity="minute"
                                                        showClearButton
                                                        containerClassName="max-w-[400px]"
                                                    />
                                                    <BrowserTimezoneNote />
                                                </Field>
                                            )}
                                        />
                                    </InfoBlock>
                                    <NextRunsField
                                        nextRuns={settings?.nextRuns ?? []}
                                        titleWidth={TITLE_WIDTH}
                                    />
                                </SectionBody>

                                <SectionHeader>Notification Configuration</SectionHeader>
                                <SectionBody>
                                    <div className={cn(dashedBorderBox)}>
                                        A daily scan that succeeded notifies only when it found something new: an app
                                        that runs a fixable CRITICAL or HIGH vulnerability its last report did not have.
                                    </div>
                                    <NotificationSettings<SchemaInput>
                                        names={{
                                            successUseDefault: "notification.successUseDefault",
                                            success: "notification.success",
                                            failureUseDefault: "notification.failureUseDefault",
                                            failure: "notification.failure",
                                        }}
                                        sources={sources}
                                        manageLink={manageLink}
                                        readOnly={readOnly}
                                        titleWidth={TITLE_WIDTH}
                                    />
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
    settings?: SystemImageScanSettings;
    readOnly?: boolean;
    onSubmit: (values: SchemaOutput) => void;
}>;

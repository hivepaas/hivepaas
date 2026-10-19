import { Tabs, TabsList, TabsTrigger } from "@components/ui/tabs";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useWatch } from "react-hook-form";
import { type AppManagedStatus, MANAGED_RESTORE_TARGET, type ManagedRestoreTargetKind } from "~/projects/domain";
import { formatManagedTime, managedEngineOf } from "~/projects/module-shared/utils";

import { BrowserTimezoneNote, InfoBlock } from "@application/shared/components";

import { Button } from "@/components/ui/button";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { DialogActionFooter, DialogBody } from "@/components/ui/dialog";
import { Field, FieldDescription, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import {
    type ManagedRestoreFormInput,
    type ManagedRestoreFormOutput,
    createManagedRestoreFormSchema,
} from "../schemas";
import type { ManagedRestorePreset } from "../types";

const TITLE_WIDTH = 150;

const TARGET_LABEL: Record<ManagedRestoreTargetKind, string> = {
    [MANAGED_RESTORE_TARGET.latest]: "Latest",
    [MANAGED_RESTORE_TARGET.time]: "A time",
    [MANAGED_RESTORE_TARGET.point]: "A restore point",
    [MANAGED_RESTORE_TARGET.backup]: "A base backup",
};

/**
 * A cluster volume the new app could keep its data on: where its data is, whether that is every node - shared
 * storage, which needs the archive in object storage - and whether it is kept in memory, which it never can be.
 */
export interface ManagedRestoreVolumeOption {
    id: string;
    name: string;
    place: string;
    shared: boolean;
    inMemory: boolean;
}

export function ManagedRestoreForm({
    sourceName,
    status,
    volumes,
    preset,
    isPending,
    onSubmit,
    readOnly = false,
}: Props) {
    const { restoreWindow } = status;
    const offSite = status.placement === "off-site";
    const { archiveNoun } = managedEngineOf(status.engine);
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<ManagedRestoreFormInput, unknown, ManagedRestoreFormOutput>({
        defaultValues: {
            kind: preset?.kind ?? MANAGED_RESTORE_TARGET.latest,
            time: null,
            point: preset?.point ?? "",
            backup: preset?.backup ?? "",
            dataVolume: "",
            name: "",
        },
        resolver: zodResolver(createManagedRestoreFormSchema(restoreWindow)),
        mode: "onSubmit",
    });

    const { field: kindField } = useController({ name: "kind", control });
    const {
        field: timeField,
        fieldState: { invalid: isTimeInvalid },
    } = useController({ name: "time", control });
    const { field: pointField } = useController({ name: "point", control });
    const { field: backupField } = useController({ name: "backup", control });
    const { field: volumeField } = useController({ name: "dataVolume", control });
    const { field: nameField } = useController({ name: "name", control });
    const kind = useWatch({ control, name: "kind" });

    // The newest first: the one a restore most often wants.
    const backups = [...status.backups].reverse();

    function onValid(values: ManagedRestoreFormOutput) {
        if (readOnly) {
            return;
        }

        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<ManagedRestoreFormOutput>) {
        console.error(_errors);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                if (readOnly) {
                    return;
                }

                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <fieldset
                disabled={readOnly || isPending}
                className="contents"
            >
                <DialogBody>
                    <FieldGroup>
                        <p className="text-xs text-muted-foreground">
                            The database is restored into a new app, from the same template, version and parameters.{" "}
                            {sourceName} is only read: it goes on as it is.
                        </p>

                        <Field>
                            <InfoBlock
                                title="Restore to"
                                titleWidth={TITLE_WIDTH}
                            >
                                <Tabs
                                    value={kindField.value}
                                    onValueChange={value => {
                                        kindField.onChange(value);
                                    }}
                                >
                                    <TabsList className="flex-wrap h-auto">
                                        {Object.entries(TARGET_LABEL).map(([value, label]) => (
                                            <TabsTrigger
                                                key={value}
                                                value={value}
                                                className="flex-1"
                                            >
                                                {label}
                                            </TabsTrigger>
                                        ))}
                                    </TabsList>
                                </Tabs>
                                {restoreWindow && (
                                    <FieldDescription>
                                        A restore can go from {formatManagedTime(restoreWindow.from)} to{" "}
                                        {formatManagedTime(restoreWindow.to)}.
                                    </FieldDescription>
                                )}
                            </InfoBlock>
                        </Field>

                        {kind === MANAGED_RESTORE_TARGET.latest && (
                            <p className="text-xs text-muted-foreground">
                                Everything archived when the restore starts: about the archive timeout&apos;s worth of
                                writes behind the source.
                            </p>
                        )}

                        {kind === MANAGED_RESTORE_TARGET.time && (
                            <Field>
                                <InfoBlock
                                    title="Time"
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <DateTimePicker
                                        value={timeField.value ?? undefined}
                                        onChange={date => {
                                            timeField.onChange(date ?? null);
                                        }}
                                        displayFormat={{ hour24: "yyyy-MM-dd HH:mm:ss" }}
                                        granularity="second"
                                        aria-invalid={isTimeInvalid}
                                        disabled={readOnly}
                                    />
                                    <BrowserTimezoneNote />
                                    <FieldDescription>
                                        The database as it was then: a transaction committed after it is not there.
                                    </FieldDescription>
                                    <FieldError errors={[errors.time]} />
                                </InfoBlock>
                            </Field>
                        )}

                        {kind === MANAGED_RESTORE_TARGET.point && (
                            <Field>
                                <InfoBlock
                                    title="Restore point"
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <Select
                                        value={pointField.value}
                                        onValueChange={pointField.onChange}
                                        disabled={readOnly}
                                    >
                                        <SelectTrigger
                                            className="w-full"
                                            aria-label="Restore point"
                                        >
                                            <SelectValue placeholder="Choose a restore point" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {status.restorePoints.map(point => (
                                                <SelectItem
                                                    key={point.name}
                                                    value={point.name}
                                                    textValue={point.name}
                                                    // One no base backup kept reaches cannot be restored.
                                                    disabled={!point.restorableUntil}
                                                >
                                                    {point.name}
                                                    <span className="ml-2 text-xs text-muted-foreground">
                                                        {formatManagedTime(point.at)}
                                                        {!point.restorableUntil && ", no longer restorable"}
                                                    </span>
                                                </SelectItem>
                                            ))}
                                            {status.restorePoints.length === 0 && (
                                                <div className="px-2 py-1.5 text-xs text-muted-foreground">
                                                    No restore point yet.
                                                </div>
                                            )}
                                        </SelectContent>
                                    </Select>
                                    <FieldDescription>
                                        Exactly the database when the point was made, by its position in the{" "}
                                        {archiveNoun}.
                                    </FieldDescription>
                                    <FieldError errors={[errors.point]} />
                                </InfoBlock>
                            </Field>
                        )}

                        {kind === MANAGED_RESTORE_TARGET.backup && (
                            <Field>
                                <InfoBlock
                                    title="Base backup"
                                    titleWidth={TITLE_WIDTH}
                                >
                                    <Select
                                        value={backupField.value}
                                        onValueChange={backupField.onChange}
                                        disabled={readOnly}
                                    >
                                        <SelectTrigger
                                            className="w-full"
                                            aria-label="Base backup"
                                        >
                                            <SelectValue placeholder="Choose a base backup" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {backups.map(backup => (
                                                <SelectItem
                                                    key={backup.name}
                                                    value={backup.name}
                                                    textValue={backup.name}
                                                >
                                                    {formatManagedTime(backup.at)}
                                                    <span className="ml-2 font-mono text-[11px] text-muted-foreground">
                                                        {backup.name}
                                                    </span>
                                                </SelectItem>
                                            ))}
                                            {backups.length === 0 && (
                                                <div className="px-2 py-1.5 text-xs text-muted-foreground">
                                                    No base backup yet.
                                                </div>
                                            )}
                                        </SelectContent>
                                    </Select>
                                    <FieldDescription>
                                        The database as the backup was taken, with no {archiveNoun} replayed after it.
                                    </FieldDescription>
                                    <FieldError errors={[errors.backup]} />
                                </InfoBlock>
                            </Field>
                        )}

                        <Field>
                            <InfoBlock
                                title="Data volume"
                                titleWidth={TITLE_WIDTH}
                            >
                                <Select
                                    value={volumeField.value}
                                    onValueChange={volumeField.onChange}
                                    disabled={readOnly}
                                >
                                    <SelectTrigger
                                        className="w-full"
                                        aria-label="Data volume"
                                    >
                                        <SelectValue placeholder="Choose a cluster volume" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {volumes.map(volume => (
                                            <SelectItem
                                                key={volume.id}
                                                value={volume.id}
                                                textValue={volume.name}
                                                disabled={volumeRefusal(volume, offSite) !== null}
                                            >
                                                {volume.name}
                                                <span className="ml-2 text-xs text-muted-foreground">
                                                    {volumeRefusal(volume, offSite) ?? volume.place}
                                                </span>
                                            </SelectItem>
                                        ))}
                                        {volumes.length === 0 && (
                                            <div className="px-2 py-1.5 text-xs text-muted-foreground">
                                                No cluster volume in this project yet.
                                            </div>
                                        )}
                                    </SelectContent>
                                </Select>
                                <FieldDescription>
                                    A volume of its own: the new app runs where it is, and recovers its data into it.
                                    One on every node needs the archive in object storage.
                                </FieldDescription>
                                <FieldError errors={[errors.dataVolume]} />
                            </InfoBlock>
                        </Field>

                        <Field>
                            <InfoBlock
                                title="App name"
                                titleWidth={TITLE_WIDTH}
                            >
                                <Input
                                    {...nameField}
                                    aria-label="App name"
                                    placeholder={`${sourceName}-restored-yyyymmdd-hhmmss`}
                                    aria-invalid={Boolean(errors.name)}
                                />
                                <FieldDescription>Optional.</FieldDescription>
                                <FieldError errors={[errors.name]} />
                            </InfoBlock>
                        </Field>
                    </FieldGroup>
                </DialogBody>
                <DialogActionFooter>
                    <Button
                        type="submit"
                        isLoading={isPending}
                        disabled={readOnly || !restoreWindow}
                        className="min-w-[100px]"
                    >
                        Restore
                    </Button>
                </DialogActionFooter>
            </fieldset>
        </form>
    );
}

interface Props {
    sourceName: string;
    status: AppManagedStatus;
    volumes: ManagedRestoreVolumeOption[];
    preset: ManagedRestorePreset | null;
    isPending: boolean;
    onSubmit: (values: ManagedRestoreFormOutput) => Promise<void> | void;
    readOnly?: boolean;
}

/** Why the new app cannot keep its data on a volume, null when it can. */
function volumeRefusal(volume: ManagedRestoreVolumeOption, offSite: boolean): string | null {
    if (volume.inMemory) {
        return "kept in memory";
    }
    return volume.shared && !offSite ? "every node: needs an archive in object storage" : null;
}

import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useFormState, useWatch } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { estimateKeptBytes, managedEngineOf } from "~/projects/module-shared/utils";

import { BrowserTimezoneNote, InfoBlock } from "@application/shared/components";
import { formatDataSizeCompact } from "@application/shared/utils/data-size";

import { Button } from "@/components/ui/button";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { DialogActionFooter, DialogBody } from "@/components/ui/dialog";
import { Field, FieldDescription, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

import {
    type ManagedRestorePointFormInput,
    type ManagedRestorePointFormOutput,
    createManagedRestorePointFormSchema,
} from "../schemas";

const TITLE_WIDTH = 150;

export function ManagedRestorePointForm({
    creating,
    initialValues,
    engine,
    retainFull,
    lastBackupBytes,
    isPending,
    onSubmit,
    onHasChanges,
    readOnly = false,
}: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<ManagedRestorePointFormInput, unknown, ManagedRestorePointFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            note: initialValues?.note ?? "",
            keepUntil: initialValues?.keepUntil ?? null,
        },
        resolver: zodResolver(createManagedRestorePointFormSchema(creating)),
        mode: "onSubmit",
    });

    const { isDirty } = useFormState({ control });

    useUpdateEffect(() => {
        onHasChanges?.(readOnly ? false : isDirty);
    }, [isDirty, readOnly]);

    const { field: name } = useController({ name: "name", control });
    const { field: note } = useController({ name: "note", control });
    const {
        field: keepUntil,
        fieldState: { invalid: isKeepUntilInvalid },
    } = useController({ name: "keepUntil", control });
    const keptUntil = useWatch({ control, name: "keepUntil" });
    const estimate = keptUntil ? estimateKeptBytes(keptUntil, new Date(), retainFull, lastBackupBytes) : null;
    const { archiveNoun } = managedEngineOf(engine);

    function onValid(values: ManagedRestorePointFormOutput) {
        if (readOnly) {
            return;
        }

        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<ManagedRestorePointFormOutput>) {
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
                        <Field>
                            <InfoBlock
                                title="Name"
                                titleWidth={TITLE_WIDTH}
                            >
                                <Input
                                    {...name}
                                    aria-label="Name"
                                    placeholder="before-migration-42"
                                    disabled={!creating || readOnly}
                                    aria-invalid={Boolean(errors.name)}
                                />
                                <FieldDescription>
                                    {creating
                                        ? "What the point is known by. A restore targets its position in the " +
                                          `${archiveNoun}, so a name used again in the past does not matter.`
                                        : "A point's name stays as it was made."}
                                </FieldDescription>
                                <FieldError errors={[errors.name]} />
                            </InfoBlock>
                        </Field>

                        <Field>
                            <InfoBlock
                                title="Note"
                                titleWidth={TITLE_WIDTH}
                            >
                                <Textarea
                                    {...note}
                                    aria-label="Note"
                                    rows={3}
                                    placeholder="What was about to happen"
                                    aria-invalid={Boolean(errors.note)}
                                />
                                <FieldError errors={[errors.note]} />
                            </InfoBlock>
                        </Field>

                        <Field>
                            <InfoBlock
                                title="Keep until"
                                titleWidth={TITLE_WIDTH}
                            >
                                <DateTimePicker
                                    value={keepUntil.value ?? undefined}
                                    onChange={date => {
                                        keepUntil.onChange(date ?? null);
                                    }}
                                    displayFormat={{ hour24: "yyyy-MM-dd HH:mm" }}
                                    granularity="minute"
                                    showClearButton
                                    aria-invalid={isKeepUntilInvalid}
                                    disabled={readOnly}
                                />
                                <BrowserTimezoneNote />
                                <FieldDescription>
                                    {keptUntil
                                        ? keptCopy(estimate, archiveNoun)
                                        : "Without a date, it can be restored while retention keeps the base backup " +
                                          `before it: about ${retainFull} ${retainFull === 1 ? "day" : "days"}.`}
                                </FieldDescription>
                                <FieldError errors={[errors.keepUntil]} />
                            </InfoBlock>
                        </Field>
                    </FieldGroup>
                </DialogBody>
                <DialogActionFooter>
                    {isPending && creating && (
                        <span className="mr-auto text-xs text-muted-foreground">
                            Waiting for the point&apos;s {archiveNoun} to be archived, so it can be restored once
                            listed.
                        </span>
                    )}
                    <Button
                        type="submit"
                        isLoading={isPending}
                        disabled={readOnly}
                        className="min-w-[100px]"
                    >
                        {creating ? "Make point" : "Save"}
                    </Button>
                </DialogActionFooter>
            </fieldset>
        </form>
    );
}

function keptCopy(estimate: ReturnType<typeof estimateKeptBytes>, archiveNoun: string): string {
    const holds =
        "Retention keeps every base backup from the one before it, and all " + `${archiveNoun} after, until then.`;
    if (estimate === null) {
        return `${holds} There is no base backup yet to estimate the storage from.`;
    }
    if (estimate.backups === 0) {
        return `${holds} Retention keeps that long anyway: nothing more is held back.`;
    }
    return (
        `${holds} About ${formatDataSizeCompact(estimate.bytes)} more, at ${estimate.backups} base ` +
        `${estimate.backups === 1 ? "backup" : "backups"} the size of the last, with the ${archiveNoun} written ` +
        "meanwhile on top."
    );
}

interface Props {
    creating: boolean;
    initialValues?: ManagedRestorePointFormInput;
    /** The database's engine, which the copy names what it archives after. */
    engine: string;
    retainFull: number;
    lastBackupBytes?: number;
    isPending: boolean;
    onSubmit: (values: ManagedRestorePointFormOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
    readOnly?: boolean;
}

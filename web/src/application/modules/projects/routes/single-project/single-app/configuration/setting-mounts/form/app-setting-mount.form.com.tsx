import { useEffect } from "react";

import { Badge } from "@components/ui/badge";
import { zodResolver } from "@hookform/resolvers/zod";
import { LockIcon } from "lucide-react";
import { type FieldErrors, useController, useFieldArray, useForm, useWatch } from "react-hook-form";
import type { AppSettingMountSource } from "~/projects/domain";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { AppSettingMountsTableDefs } from "~/projects/module-shared/definitions/tables/app-setting-mounts";

import { FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";

import { type AppSettingMountFormInput, type AppSettingMountFormOutput, AppSettingMountFormSchema } from "../schemas";
import { GATED_PART_REASON, type Grant, rowsFor, suggestPath, widensGrants } from "../utils";

import { SourcePicker } from "./source-picker.com";

function grantsOf(values: AppSettingMountFormOutput): Grant[] {
    const source = values.source?.id ?? "";
    return values.files.filter(file => file.enabled && file.gated).map(file => ({ source, part: file.part }));
}

export function AppSettingMountForm({
    sources,
    mayMountSensitive,
    isPending,
    initialValues,
    initialGrants,
    isDisabledEntry = false,
    onSubmit,
    onHasChanges,
    onClose,
    readOnly = false,
    stickyActions = false,
}: Props) {
    const firstType = sources[0]?.type ?? "";
    const {
        handleSubmit,
        control,
        setError,
        setValue,
        getValues,
        formState: { errors, isDirty },
    } = useForm<AppSettingMountFormInput, unknown, AppSettingMountFormOutput>({
        defaultValues: initialValues ?? {
            name: "",
            inheritable: false,
            sourceType: firstType,
            source: null,
            files: rowsFor(sources, firstType),
        },
        resolver: zodResolver(AppSettingMountFormSchema),
        mode: "onSubmit",
    });
    const { fields, replace } = useFieldArray({ control, name: "files" });

    const sourceType = useWatch({ control, name: "sourceType" });
    const source = useWatch({ control, name: "source" });
    const rows = useWatch({ control, name: "files" });

    useEffect(() => {
        onHasChanges?.(readOnly ? false : isDirty);
    }, [isDirty, onHasChanges, readOnly]);

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const { field: inheritable } = useController({ name: "inheritable", control });
    const {
        field: sourceField,
        fieldState: { invalid: isSourceInvalid },
    } = useController({ name: "source", control });

    // A disabled entry hands out nothing, before or after the save: enabling it
    // is what asks.
    const gatesSave = !mayMountSensitive && !isDisabledEntry;

    function isLocked(part: string, gated: boolean): boolean {
        return gatesSave && gated && !initialGrants.some(grant => grant.source === source?.id && grant.part === part);
    }

    function onValid(values: AppSettingMountFormOutput) {
        if (readOnly) {
            return;
        }
        // Only what the save adds is gated: an entry that already mounts a
        // private key still saves a change of path.
        if (gatesSave && widensGrants(initialGrants, grantsOf(values)).length > 0) {
            setError("root", { message: GATED_PART_REASON });
            return;
        }
        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<AppSettingMountFormOutput>) {
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
                disabled={readOnly}
                className="contents"
            >
                <div className="flex flex-col gap-6">
                    <InfoBlock
                        titleWidth={240}
                        title={
                            <LabelWithInfo
                                label="Name"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    id="app-setting-mount-name"
                                    {...name}
                                    placeholder="tls-files"
                                    aria-invalid={isNameInvalid}
                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                />
                                <p className="text-sm text-muted-foreground">
                                    Lowercase letters, digits and hyphens. Names the Docker objects that hold the files.
                                </p>
                                <FieldError errors={[errors.name]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={240}
                        title={
                            <LabelWithInfo
                                label="Mount From"
                                isRequired
                            />
                        }
                    >
                        <Tabs
                            value={sourceType}
                            onValueChange={nextType => {
                                // The parts of the old type go with it: a certificate's
                                // private key must not stay ticked under basic auth.
                                setValue("sourceType", nextType, { shouldDirty: true });
                                setValue("source", null, { shouldDirty: true });
                                replace(rowsFor(sources, nextType));
                            }}
                        >
                            <TabsList className="bg-muted/80 p-1 rounded-lg flex-wrap h-auto">
                                {sources.map(item => (
                                    <TabsTrigger
                                        key={item.type}
                                        value={item.type}
                                    >
                                        {AppSettingMountsTableDefs.sourceTypeLabels[item.type] ?? item.type}
                                    </TabsTrigger>
                                ))}
                            </TabsList>
                        </Tabs>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={240}
                        title={
                            <LabelWithInfo
                                label="Setting"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <SourcePicker
                                    sourceType={sourceType}
                                    value={sourceField.value}
                                    invalid={isSourceInvalid}
                                    disabled={readOnly}
                                    onChange={next => {
                                        // A path still at its suggestion follows the setting's name.
                                        const oldName = getValues("source")?.name ?? "";
                                        getValues("files").forEach((row, index) => {
                                            if (row.path === suggestPath(sourceType, row.part, oldName)) {
                                                setValue(
                                                    `files.${index}.path`,
                                                    suggestPath(sourceType, row.part, next?.name ?? ""),
                                                );
                                            }
                                        });
                                        sourceField.onChange(next);
                                    }}
                                />
                                <FieldError errors={[errors.source]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={240}
                        title={
                            <LabelWithInfo
                                label="Files"
                                isRequired
                            />
                        }
                    >
                        <div className="flex flex-col gap-4">
                            {fields.map((field, index) => {
                                const row = rows[index] ?? field;
                                const locked = isLocked(row.part, row.gated);
                                const rowErrors = errors.files?.[index];
                                return (
                                    <div
                                        key={field.id}
                                        className="flex flex-col gap-2 rounded-md border p-3"
                                    >
                                        <div className="flex flex-wrap items-center gap-2">
                                            <Checkbox
                                                id={`app-setting-mount-part-${row.part}`}
                                                checked={row.enabled}
                                                disabled={readOnly || (locked && !row.enabled)}
                                                onCheckedChange={checked => {
                                                    setValue(`files.${index}.enabled`, checked === true, {
                                                        shouldDirty: true,
                                                    });
                                                }}
                                            />
                                            <label
                                                htmlFor={`app-setting-mount-part-${row.part}`}
                                                className="font-medium"
                                            >
                                                {row.part}
                                            </label>
                                            {row.gated && (
                                                <Badge className="bg-amber-500 text-white">
                                                    <LockIcon className="size-3" /> Sensitive
                                                </Badge>
                                            )}
                                            {row.secret && !row.gated && (
                                                <Badge variant="outline">Stored as secret</Badge>
                                            )}
                                        </div>
                                        {locked && <p className="text-xs text-muted-foreground">{GATED_PART_REASON}</p>}
                                        {row.gated && isDisabledEntry && !mayMountSensitive && (
                                            <p className="text-xs text-muted-foreground">
                                                The entry is disabled: enabling it with this part takes the Can Reveal
                                                Secrets permission.
                                            </p>
                                        )}
                                        <div className="flex flex-wrap gap-2">
                                            <Input
                                                aria-label={`${row.part} path`}
                                                value={row.path}
                                                disabled={!row.enabled}
                                                onChange={event => {
                                                    setValue(`files.${index}.path`, event.target.value, {
                                                        shouldDirty: true,
                                                    });
                                                }}
                                                aria-invalid={Boolean(rowErrors?.path)}
                                                className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                            />
                                            {(["mode", "uid", "gid"] as const).map(key => (
                                                <Input
                                                    key={key}
                                                    aria-label={`${row.part} ${key}`}
                                                    placeholder={key}
                                                    value={row[key]}
                                                    disabled={!row.enabled}
                                                    onChange={event => {
                                                        setValue(`files.${index}.${key}`, event.target.value, {
                                                            shouldDirty: true,
                                                        });
                                                    }}
                                                    aria-invalid={Boolean(rowErrors?.[key])}
                                                    className="max-w-[120px]"
                                                />
                                            ))}
                                        </div>
                                        <FieldError errors={[rowErrors?.path, rowErrors?.mode]} />
                                    </div>
                                );
                            })}
                            <FieldError errors={[errors.files?.root ?? errors.files]} />
                        </div>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={240}
                        title={<LabelWithInfo label="Inheritable" />}
                    >
                        <div className="flex flex-col gap-1">
                            <div className="flex items-center gap-2">
                                <Checkbox
                                    id="app-setting-mount-inheritable"
                                    checked={inheritable.value}
                                    onCheckedChange={checked => {
                                        inheritable.onChange(checked === true);
                                    }}
                                />
                                <label htmlFor="app-setting-mount-inheritable">
                                    Previews and clones get this entry
                                </label>
                            </div>
                            <p className="text-sm text-muted-foreground">
                                A preview runs a pull request&apos;s code: whatever this entry mounts reaches it.
                            </p>
                        </div>
                    </InfoBlock>

                    {errors.root && <FieldError errors={[errors.root]} />}
                </div>
                {!readOnly && (
                    <FormActionBar sticky={stickyActions}>
                        <Button
                            type="button"
                            variant="outline"
                            className="min-w-[100px]"
                            disabled={isPending}
                            onClick={onClose}
                        >
                            Cancel
                        </Button>
                        <Button
                            type="submit"
                            isLoading={isPending}
                            className="min-w-[100px]"
                        >
                            Save
                        </Button>
                    </FormActionBar>
                )}
                {readOnly && (
                    <FormActionBar sticky={stickyActions}>
                        <Button
                            type="button"
                            onClick={onClose}
                            className="min-w-[100px]"
                        >
                            Close
                        </Button>
                    </FormActionBar>
                )}
            </fieldset>
        </form>
    );
}

interface Props {
    sources: AppSettingMountSource[];
    mayMountSensitive: boolean;
    isPending: boolean;
    readOnly?: boolean;
    initialValues?: AppSettingMountFormInput;
    /** The grants the entry holds as loaded: none for a new or a disabled entry. */
    initialGrants: Grant[];
    /** The entry being edited is disabled: a save hands out nothing, so nothing is locked. */
    isDisabledEntry?: boolean;
    onSubmit: (values: AppSettingMountFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    onClose?: () => void;
    stickyActions?: boolean;
}

import { useEffect, useId, useMemo } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useController, useForm, useFormState } from "react-hook-form";
import type { ProjectCommandTemplate } from "~/projects/domain";
import { CommandArgGroupsSection } from "~/projects/module-shared/components";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import {
    InheritedSettingReadonlyNotice,
    PermissionReadonlyNotice,
    SettingsFormCancelAction,
} from "~/settings/module-shared/components";

import { ContentBlock, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ECommandTemplateKind } from "@application/shared/enums";
import { KeyValueList } from "@application/shared/form";

import {
    Button,
    Checkbox,
    Field,
    FieldError,
    FieldGroup,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
    Tabs,
    TabsList,
    TabsTrigger,
} from "@/components/ui";
import { InputNumber } from "@/components/ui/input-number";

import {
    PROJECT_COMMAND_TEMPLATE_COMMAND_MODE,
    type ProjectCommandTemplateFormInput,
    type ProjectCommandTemplateFormOutput,
    ProjectCommandTemplateFormSchema,
} from "../schemas";

import {
    createEmptyProjectCommandTemplateFormDefaults,
    mapProjectCommandTemplateToFormInput,
} from "./project-command-template.form-mappers";
import { ProjectCommandTemplateScriptEditorField } from "./script-editor-field.com";

const INFO_BLOCK_TITLE_WIDTH = 220;

const COMMAND_TEMPLATE_KIND_OPTIONS = [
    { value: ECommandTemplateKind.Backup, label: "backup" },
    { value: ECommandTemplateKind.DataOps, label: "data-ops" },
    { value: ECommandTemplateKind.Database, label: "database" },
    { value: ECommandTemplateKind.Deployment, label: "deployment" },
    { value: ECommandTemplateKind.Diagnostics, label: "diagnostics" },
    { value: ECommandTemplateKind.Maintenance, label: "maintenance" },
    { value: ECommandTemplateKind.Testing, label: "testing" },
] as const;

type SchemaInput = ProjectCommandTemplateFormInput;
type SchemaOutput = ProjectCommandTemplateFormOutput;

export function ProjectCommandTemplateForm({
    isPending,
    onSubmit,
    initialValues,
    onHasChanges,
    savedVersion = 0,
    readOnlyInherited = false,
    readOnly = false,
    onClose,
}: Props) {
    const isReadOnly = readOnlyInherited || readOnly;
    const defaultValues = useMemo(() => {
        if (initialValues) {
            return mapProjectCommandTemplateToFormInput(initialValues);
        }

        return createEmptyProjectCommandTemplateFormDefaults();
    }, [initialValues]);

    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues,
        resolver: zodResolver(ProjectCommandTemplateFormSchema),
        mode: "onSubmit",
    });

    const {
        control,
        handleSubmit,
        getValues,
        reset,
        formState: { errors },
    } = methods;
    const { isDirty } = useFormState({ control });

    useEffect(() => {
        if (savedVersion === 0) {
            return;
        }

        reset(getValues());
        onHasChanges?.(false);
    }, [getValues, onHasChanges, reset, savedVersion]);

    useEffect(() => {
        onHasChanges?.(isReadOnly ? false : isDirty);
    }, [isDirty, isReadOnly, onHasChanges]);

    const consoleWidthInputId = useId();
    const consoleHeightInputId = useId();

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ control, name: "name" });
    const {
        field: kind,
        fieldState: { invalid: isKindInvalid },
    } = useController({ control, name: "kind" });
    const { field: commandMode } = useController({ control, name: "commandMode" });
    const {
        field: command,
        fieldState: { invalid: isCommandInvalid },
    } = useController({ control, name: "command" });
    const {
        field: script,
        fieldState: { invalid: isScriptInvalid },
    } = useController({ control, name: "script" });
    const {
        field: workingDir,
        fieldState: { invalid: isWorkingDirInvalid },
    } = useController({ control, name: "workingDir" });
    const { field: tty } = useController({ control, name: "tty" });
    const {
        field: consoleWidth,
        fieldState: { invalid: isConsoleWidthInvalid },
    } = useController({ control, name: "consoleSize.width" });
    const {
        field: consoleHeight,
        fieldState: { invalid: isConsoleHeightInvalid },
    } = useController({ control, name: "consoleSize.height" });
    const { field: defaultField } = useController({ control, name: "default" });

    function onValid(values: SchemaOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<SchemaOutput>) {
        console.error(_errors);
    }

    return (
        <FormProvider {...methods}>
            <form
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onValid, onInvalid)(event);
                }}
                className="min-h-0 flex flex-1 flex-col"
            >
                <div>
                    {readOnlyInherited && <InheritedSettingReadonlyNotice />}
                    {readOnly && !readOnlyInherited && <PermissionReadonlyNotice />}
                    <fieldset
                        disabled={isReadOnly}
                        className="contents"
                    >
                        <FieldGroup className="gap-6">
                            <InfoBlock
                                titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                title={
                                    <LabelWithInfo
                                        label="Name"
                                        isRequired
                                    />
                                }
                            >
                                <Field>
                                    <Input
                                        {...name}
                                        placeholder="template name"
                                        aria-invalid={isNameInvalid}
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                    />
                                    <FieldError errors={[errors.name]} />
                                </Field>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                title={
                                    <LabelWithInfo
                                        label="Type"
                                        isRequired
                                    />
                                }
                            >
                                <Field>
                                    <Select
                                        value={kind.value}
                                        onValueChange={kind.onChange}
                                    >
                                        <SelectTrigger
                                            aria-invalid={isKindInvalid}
                                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                        >
                                            <SelectValue placeholder="select type" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {COMMAND_TEMPLATE_KIND_OPTIONS.map(option => (
                                                <SelectItem
                                                    key={option.value}
                                                    value={option.value}
                                                >
                                                    {option.label}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                    <FieldError errors={[errors.kind]} />
                                </Field>
                            </InfoBlock>

                            <ContentBlock label="Command">
                                <div className="flex flex-col gap-6">
                                    <InfoBlock
                                        title="Type"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <Tabs
                                            value={commandMode.value}
                                            onValueChange={commandMode.onChange}
                                        >
                                            <TabsList>
                                                <TabsTrigger value={PROJECT_COMMAND_TEMPLATE_COMMAND_MODE.Command}>
                                                    Command
                                                </TabsTrigger>
                                                <TabsTrigger value={PROJECT_COMMAND_TEMPLATE_COMMAND_MODE.Script}>
                                                    Script
                                                </TabsTrigger>
                                            </TabsList>
                                        </Tabs>
                                    </InfoBlock>

                                    <InfoBlock
                                        title={
                                            <LabelWithInfo
                                                label={
                                                    commandMode.value === PROJECT_COMMAND_TEMPLATE_COMMAND_MODE.Script
                                                        ? "Script"
                                                        : "Command"
                                                }
                                                isRequired
                                            />
                                        }
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        {commandMode.value === PROJECT_COMMAND_TEMPLATE_COMMAND_MODE.Script ? (
                                            <Field>
                                                <ProjectCommandTemplateScriptEditorField
                                                    value={script.value}
                                                    onChange={script.onChange}
                                                    invalid={isScriptInvalid}
                                                    error={errors.script}
                                                    readOnly={isReadOnly}
                                                />
                                            </Field>
                                        ) : (
                                            <Field>
                                                <Input
                                                    {...command}
                                                    placeholder='echo "$CMD_ARG_GROUP_1"'
                                                    aria-invalid={isCommandInvalid}
                                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                />
                                                <FieldError errors={[errors.command]} />
                                            </Field>
                                        )}
                                    </InfoBlock>

                                    <InfoBlock
                                        title="Working Dir"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <Field>
                                            <Input
                                                {...workingDir}
                                                placeholder="/path/to/working/dir"
                                                aria-invalid={isWorkingDirInvalid}
                                                className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                            />
                                            <FieldError errors={[errors.workingDir]} />
                                        </Field>
                                    </InfoBlock>

                                    <InfoBlock
                                        title="Terminal"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <div className="flex w-full max-w-[430px] flex-wrap items-start gap-x-4 gap-y-3">
                                            <div className="flex h-9 items-center gap-3 text-sm font-medium">
                                                <span>TTY</span>
                                                <Checkbox
                                                    checked={tty.value}
                                                    onCheckedChange={checked => {
                                                        tty.onChange(checked === true);
                                                    }}
                                                    aria-label="TTY"
                                                />
                                            </div>

                                            {tty.value && (
                                                <>
                                                    <div className="flex min-w-[140px] flex-1 flex-col gap-1.5">
                                                        <div className="flex min-w-0 items-center gap-2">
                                                            <label
                                                                htmlFor={consoleWidthInputId}
                                                                className="shrink-0 text-sm font-medium"
                                                            >
                                                                Width
                                                            </label>
                                                            <InputNumber
                                                                id={consoleWidthInputId}
                                                                ref={consoleWidth.ref}
                                                                name={consoleWidth.name}
                                                                value={consoleWidth.value}
                                                                onBlur={consoleWidth.onBlur}
                                                                onValueChange={value => {
                                                                    const nextValue =
                                                                        value !== undefined && Number.isFinite(value)
                                                                            ? value
                                                                            : undefined;

                                                                    consoleWidth.onChange(nextValue);
                                                                }}
                                                                useGrouping={false}
                                                                showControls={false}
                                                                placeholder="120"
                                                                aria-invalid={isConsoleWidthInvalid}
                                                                className="min-w-0 flex-1"
                                                            />
                                                        </div>
                                                        <FieldError errors={[errors.consoleSize?.width]} />
                                                    </div>

                                                    <div className="flex min-w-[140px] flex-1 flex-col gap-1.5">
                                                        <div className="flex min-w-0 items-center gap-2">
                                                            <label
                                                                htmlFor={consoleHeightInputId}
                                                                className="shrink-0 text-sm font-medium"
                                                            >
                                                                Height
                                                            </label>
                                                            <InputNumber
                                                                id={consoleHeightInputId}
                                                                ref={consoleHeight.ref}
                                                                name={consoleHeight.name}
                                                                value={consoleHeight.value}
                                                                onBlur={consoleHeight.onBlur}
                                                                onValueChange={value => {
                                                                    const nextValue =
                                                                        value !== undefined && Number.isFinite(value)
                                                                            ? value
                                                                            : undefined;

                                                                    consoleHeight.onChange(nextValue);
                                                                }}
                                                                useGrouping={false}
                                                                showControls={false}
                                                                placeholder="40"
                                                                aria-invalid={isConsoleHeightInvalid}
                                                                className="min-w-0 flex-1"
                                                            />
                                                        </div>
                                                        <FieldError errors={[errors.consoleSize?.height]} />
                                                    </div>
                                                </>
                                            )}
                                        </div>
                                    </InfoBlock>

                                    <InfoBlock
                                        title="Env"
                                        titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                    >
                                        <KeyValueList<SchemaInput>
                                            name="envVars"
                                            keyLabel="Name"
                                            valueLabel="Value"
                                            keyPlaceholder="name"
                                            valuePlaceholder="value"
                                            className="max-w-[600px]"
                                            checkDuplicates
                                            disabled={isReadOnly}
                                        />
                                    </InfoBlock>
                                </div>
                            </ContentBlock>

                            <ContentBlock label="Arg Groups">
                                <CommandArgGroupsSection readOnly={isReadOnly} />
                            </ContentBlock>

                            <InfoBlock
                                titleWidth={INFO_BLOCK_TITLE_WIDTH}
                                title={<LabelWithInfo label="Default" />}
                            >
                                <Checkbox
                                    checked={defaultField.value}
                                    onCheckedChange={checked => {
                                        defaultField.onChange(Boolean(checked));
                                    }}
                                />
                            </InfoBlock>
                        </FieldGroup>
                    </fieldset>
                </div>
                {!isReadOnly && (
                    <div className="pb-6 flex justify-end mt-6">
                        <div className="flex items-center gap-3">
                            <SettingsFormCancelAction
                                onCancel={onClose}
                                disabled={isPending}
                            />
                            <Button
                                type="submit"
                                isLoading={isPending}
                                className="min-w-[100px]"
                            >
                                Save
                            </Button>
                        </div>
                    </div>
                )}
                {isReadOnly && (
                    <div className="shrink-0 px-0 mt-6 pb-6 flex justify-end">
                        <Button
                            type="button"
                            onClick={onClose}
                        >
                            Close
                        </Button>
                    </div>
                )}
            </form>
        </FormProvider>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: SchemaOutput) => void;
    initialValues?: ProjectCommandTemplate;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}

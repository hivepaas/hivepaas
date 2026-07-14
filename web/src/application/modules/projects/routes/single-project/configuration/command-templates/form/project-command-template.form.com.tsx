import { useEffect, useMemo } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useController, useForm, useFormState } from "react-hook-form";
import type { ProjectCommandTemplate } from "~/projects/domain";
import { CommandArgGroupsSection, CommandConfigSection } from "~/projects/module-shared/components";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import {
    InheritedSettingReadonlyNotice,
    PermissionReadonlyNotice,
    SettingsFormCancelAction,
} from "~/settings/module-shared/components";

import {
    ContentBlock,
    EditableCombobox,
    FormActionBar,
    InfoBlock,
    LabelWithInfo,
} from "@application/shared/components";
import { ECommandTemplateKind } from "@application/shared/enums";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";

import {
    type ProjectCommandTemplateFormInput,
    type ProjectCommandTemplateFormOutput,
    ProjectCommandTemplateFormSchema,
} from "../schemas";

import {
    createEmptyProjectCommandTemplateFormDefaults,
    mapProjectCommandTemplateToFormInput,
} from "./project-command-template.form-mappers";

const INFO_BLOCK_TITLE_WIDTH = 150;

const COMMAND_TEMPLATE_KIND_OPTIONS = Object.values(ECommandTemplateKind);

type SchemaInput = ProjectCommandTemplateFormInput;
type SchemaOutput = ProjectCommandTemplateFormOutput;

export function ProjectCommandTemplateForm({
    projectId: _projectId,
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

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ control, name: "name" });
    const {
        field: kind,
        fieldState: { invalid: isKindInvalid },
    } = useController({ control, name: "kind" });
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
                                    <EditableCombobox
                                        options={COMMAND_TEMPLATE_KIND_OPTIONS}
                                        value={kind.value}
                                        onChange={kind.onChange}
                                        placeholder="Select or enter type"
                                        aria-invalid={isKindInvalid}
                                        disabled={isReadOnly}
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                    />
                                    <FieldError errors={[errors.kind]} />
                                </Field>
                            </InfoBlock>

                            <CommandConfigSection
                                label="Command"
                                readOnly={isReadOnly}
                                envLabel="Env"
                            />

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
                    <FormActionBar>
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
                    </FormActionBar>
                )}
                {isReadOnly && (
                    <FormActionBar>
                        <Button
                            type="button"
                            onClick={onClose}
                            className="min-w-[100px]"
                        >
                            Close
                        </Button>
                    </FormActionBar>
                )}
            </form>
        </FormProvider>
    );
}

interface Props {
    projectId?: string;
    isPending: boolean;
    onSubmit: (values: SchemaOutput) => void;
    initialValues?: ProjectCommandTemplate;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}

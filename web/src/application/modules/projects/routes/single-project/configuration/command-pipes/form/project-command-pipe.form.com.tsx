import { useEffect, useMemo } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useController, useForm, useFormState } from "react-hook-form";
import type { ProjectCommandPipe } from "~/projects/domain";
import {
    InheritedSettingReadonlyNotice,
    PermissionReadonlyNotice,
    SettingsFormCancelAction,
} from "~/settings/module-shared/components";

import { FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";

import {
    type ProjectCommandPipeFormInput,
    type ProjectCommandPipeFormOutput,
    ProjectCommandPipeFormSchema,
} from "../schemas";

import { CommandTemplateSelectField } from "./command-template-select-field.com";
import {
    createEmptyProjectCommandPipeFormDefaults,
    mapProjectCommandPipeToFormInput,
} from "./project-command-pipe.form-mappers";

type SchemaInput = ProjectCommandPipeFormInput;
type SchemaOutput = ProjectCommandPipeFormOutput;

export function ProjectCommandPipeForm({
    projectId,
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
            return mapProjectCommandPipeToFormInput(initialValues);
        }

        return createEmptyProjectCommandPipeFormDefaults();
    }, [initialValues]);

    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues,
        resolver: zodResolver(ProjectCommandPipeFormSchema),
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
                                titleWidth={150}
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
                                        placeholder="pipe name"
                                        aria-invalid={isNameInvalid}
                                        className="w-full max-w-[600px]"
                                    />
                                    <FieldError errors={[errors.name]} />
                                </Field>
                            </InfoBlock>

                            <CommandTemplateSelectField
                                projectId={projectId}
                                name="sourceCommandId"
                                label="Source Command"
                                placeholder="select command"
                                readOnly={isReadOnly}
                            />

                            <CommandTemplateSelectField
                                projectId={projectId}
                                name="targetCommandId"
                                label="Target Command"
                                placeholder="select command"
                                readOnly={isReadOnly}
                            />

                            <InfoBlock
                                titleWidth={150}
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
    projectId: string;
    isPending: boolean;
    onSubmit: (values: SchemaOutput) => void;
    initialValues?: ProjectCommandPipe;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}

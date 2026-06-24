import * as React from "react";
import { type PropsWithChildren, useImperativeHandle } from "react";

import { FieldError, Input, TagInput } from "@components/ui";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { FormProvider, useController, useForm } from "react-hook-form";
import { type ProjectAppDetails, type ProjectEnvEntity } from "~/projects/domain";
import { ProjectAppStatusBadge, ProjectEnvBadge } from "~/projects/module-shared/components";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock } from "@application/shared/components";

import {
    AppConfigGeneralFormSchema,
    type AppConfigGeneralFormSchemaInput,
    type AppConfigGeneralFormSchemaOutput,
} from "../schemas";
import { type AppConfigGeneralFormRef } from "../types";

export function AppConfigGeneralForm({ ref, defaultValues, envs, onSubmit, readOnly = false, children }: Props) {
    const methods = useForm<AppConfigGeneralFormSchemaInput, unknown, AppConfigGeneralFormSchemaOutput>({
        defaultValues: {
            name: defaultValues.name,
            tags: defaultValues.tags,
            note: defaultValues.note,
        },
        resolver: zodResolver(AppConfigGeneralFormSchema),
        mode: "onSubmit",
    });

    const {
        control,
        formState: { errors },
        watch,
        setValue,
    } = methods;

    const tags = watch("tags");
    const selectedEnv = envs.find(env => env.name === defaultValues.env);

    useImperativeHandle(
        ref,
        () => ({
            setValues: (values: Partial<AppConfigGeneralFormSchemaInput>) => {
                methods.reset({
                    ...methods.getValues(),
                    ...values,
                });
            },
            onError: () => {
                // Implementation for error handling if needed
            },
        }),
        [methods],
    );

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({
        control,
        name: "name",
    });

    const {
        field: note,
        fieldState: { invalid: isNoteInvalid },
    } = useController({
        control,
        name: "note",
    });

    function handleCreateTag(tag: string) {
        if (readOnly) {
            return;
        }

        if (!tags.includes(tag)) {
            setValue("tags", [...tags, tag]);
        }
    }

    function handleDeleteTag(tagToRemove: string) {
        if (readOnly) {
            return;
        }

        setValue(
            "tags",
            tags.filter(tag => tag !== tagToRemove),
        );
    }

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
                        className="contents"
                    >
                        {/* ID - Read Only */}
                        <InfoBlock title="ID">
                            <Input
                                value={defaultValues.id}
                                type="text"
                                className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                disabled
                                readOnly
                            />
                        </InfoBlock>

                        {/* Name */}
                        <InfoBlock title="Name">
                            <Input
                                {...name}
                                value={name.value}
                                onChange={name.onChange}
                                type="text"
                                className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                placeholder="Enter app name"
                                aria-invalid={isNameInvalid}
                            />
                            <FieldError errors={[errors.name]} />
                        </InfoBlock>

                        {/* Key - Read Only */}
                        <InfoBlock title="Key">
                            <Input
                                value={defaultValues.key}
                                type="text"
                                className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                disabled
                                readOnly
                            />
                        </InfoBlock>

                        {/* Status - Show Label */}
                        <InfoBlock title="Status">
                            <ProjectAppStatusBadge status={defaultValues.status} />
                        </InfoBlock>

                        {/* Environment - Read Only */}
                        <InfoBlock title="Environment">
                            <ProjectEnvBadge
                                name={defaultValues.env}
                                color={selectedEnv?.color}
                            />
                        </InfoBlock>

                        {/* Tags */}
                        <InfoBlock title="Tags">
                            <div className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}>
                                <TagInput
                                    tags={tags}
                                    onCreate={handleCreateTag}
                                    onDelete={handleDeleteTag}
                                    placeholder="Enter tag"
                                    disabled={readOnly}
                                />
                                <FieldError errors={[errors.tags]} />
                            </div>
                        </InfoBlock>

                        {/* Notes */}
                        <InfoBlock title="Notes">
                            <Textarea
                                {...note}
                                value={note.value}
                                onChange={note.onChange}
                                className={`${PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS} min-h-[120px]`}
                                placeholder="Enter app notes"
                                rows={4}
                                aria-invalid={isNoteInvalid}
                            />
                            <FieldError errors={[errors.note]} />
                        </InfoBlock>

                        {children}
                    </fieldset>
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigGeneralFormRef>;
    defaultValues: ProjectAppDetails;
    envs: ProjectEnvEntity[];
    onSubmit: (values: AppConfigGeneralFormSchemaOutput) => void;
    readOnly?: boolean;
}>;

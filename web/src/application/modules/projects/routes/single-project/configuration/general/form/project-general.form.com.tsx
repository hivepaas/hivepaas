import * as React from "react";
import { type PropsWithChildren, useImperativeHandle } from "react";

import { FieldError, Input, TagInput } from "@components/ui";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { FormProvider, useController, useForm } from "react-hook-form";
import { type ProjectDetailsEntity } from "~/projects/domain";

import { InfoBlock } from "@application/shared/components";

import { ProjectStatusBadge } from "@application/modules/projects/module-shared/components";

import {
    ProjectGeneralFormSchema,
    type ProjectGeneralFormSchemaInput,
    type ProjectGeneralFormSchemaOutput,
} from "../schemas";
import { type ProjectGeneralFormRef } from "../types";

export function ProjectGeneralForm({ ref, defaultValues, onSubmit, children }: Props) {
    const methods = useForm<ProjectGeneralFormSchemaInput, unknown, ProjectGeneralFormSchemaOutput>({
        defaultValues: {
            name: defaultValues.name,
            tags: defaultValues.tags,
            note: defaultValues.note,
        },
        resolver: zodResolver(ProjectGeneralFormSchema),
        mode: "onSubmit",
    });

    const {
        control,
        formState: { errors },
        watch,
        setValue,
    } = methods;

    const tags = watch("tags");

    useImperativeHandle(
        ref,
        () => ({
            setValues: (values: Partial<ProjectGeneralFormSchemaInput>) => {
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
        if (!tags.includes(tag)) {
            setValue("tags", [...tags, tag]);
        }
    }

    function handleDeleteTag(tagToRemove: string) {
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

                        void methods.handleSubmit(onSubmit)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    {/* Name */}
                    <InfoBlock title="Name">
                        <Input
                            {...name}
                            value={name.value}
                            onChange={name.onChange}
                            type="text"
                            className="max-w-[400px]"
                            placeholder="Enter project name"
                            aria-invalid={isNameInvalid}
                        />
                        <FieldError errors={[errors.name]} />
                    </InfoBlock>

                    {/* Key - Read Only */}
                    <InfoBlock title="Key">
                        <Input
                            value={defaultValues.key}
                            type="text"
                            className="max-w-[400px]"
                            disabled
                            readOnly
                        />
                    </InfoBlock>

                    {/* Status - Show Label */}
                    <InfoBlock title="Status">
                        <ProjectStatusBadge status={defaultValues.status} />
                    </InfoBlock>

                    {/* Tags */}
                    <InfoBlock title="Tags">
                        <div className="max-w-[400px]">
                            <TagInput
                                tags={tags}
                                onCreate={handleCreateTag}
                                onDelete={handleDeleteTag}
                                placeholder="Enter tag"
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
                            className="w-[100%] min-h-[120px]"
                            placeholder="Enter project notes"
                            rows={4}
                            aria-invalid={isNoteInvalid}
                        />
                        <FieldError errors={[errors.note]} />
                    </InfoBlock>

                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<ProjectGeneralFormRef>;
    defaultValues: ProjectDetailsEntity;
    onSubmit: (values: ProjectGeneralFormSchemaOutput) => void;
}>;

import * as React from "react";
import { type PropsWithChildren, useImperativeHandle } from "react";

import { FieldError, Input, TagInput } from "@components/ui";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { FormProvider, useController, useForm } from "react-hook-form";
import { type ProjectAppDetails } from "~/projects/domain";
import { ProjectAppStatusBadge } from "~/projects/module-shared/components";

import { InfoBlock } from "@application/shared/components";

import {
    AppConfigGeneralFormSchema,
    type AppConfigGeneralFormSchemaInput,
    type AppConfigGeneralFormSchemaOutput,
} from "../schemas";
import { type AppConfigGeneralFormRef } from "../types";

export function AppConfigGeneralForm({ ref, defaultValues, onSubmit, children }: Props) {
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
                            className="max-w-[400px]"
                            disabled
                            readOnly
                        />
                    </InfoBlock>

                    {/* Status - Show Label */}
                    <InfoBlock title="Status">
                        <ProjectAppStatusBadge status={defaultValues.status} />
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
                            className="w-full min-h-[120px]"
                            placeholder="Enter app notes"
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
    ref?: React.Ref<AppConfigGeneralFormRef>;
    defaultValues: ProjectAppDetails;
    onSubmit: (values: AppConfigGeneralFormSchemaOutput) => void;
}>;

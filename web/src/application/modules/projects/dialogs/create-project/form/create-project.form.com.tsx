import React, { useEffect } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { type ProjectEnvEntity } from "~/projects/domain";
import { ProjectEnvInput } from "~/projects/module-shared/components";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { TagInput } from "@/components/ui/tag-input";
import { Textarea } from "@/components/ui/textarea";

import {
    type CreateProjectFormInput,
    type CreateProjectFormOutput,
    CreateProjectFormSchema,
    DEFAULT_PROJECT_ENVS,
} from "../schemas";

export function CreateProjectForm({ isPending, readOnly = false, onSubmit, onHasChanges }: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
        watch,
        setValue,
    } = useForm<CreateProjectFormInput, unknown, CreateProjectFormOutput>({
        defaultValues: {
            name: "",
            note: "",
            envs: DEFAULT_PROJECT_ENVS.map(env => ({ ...env })),
            tags: [],
        },
        resolver: zodResolver(CreateProjectFormSchema),
        mode: "onSubmit",
    });

    useEffect(() => {
        onHasChanges?.(readOnly ? false : isDirty);
    }, [isDirty, onHasChanges, readOnly]);

    const tags = watch("tags");
    const envs = watch("envs");

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({
        name: "name",
        control,
    });

    const {
        field: note,
        fieldState: { invalid: isNoteInvalid },
    } = useController({
        name: "note",
        control,
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

    function handleCreateEnv(env: ProjectEnvEntity) {
        if (readOnly) {
            return;
        }

        if (!envs.some(item => item.name === env.name)) {
            setValue("envs", [...envs, env], { shouldDirty: true });
        }
    }

    function handleDeleteEnv(envName: string) {
        if (readOnly) {
            return;
        }

        setValue(
            "envs",
            envs.filter(env => env.name !== envName),
            { shouldDirty: true },
        );
    }

    function handleUpdateEnvColor(envName: string, color: string) {
        if (readOnly) {
            return;
        }

        setValue(
            "envs",
            envs.map(env => (env.name === envName ? { ...env, color } : env)),
            { shouldDirty: true },
        );
    }

    function onValid(values: CreateProjectFormOutput) {
        if (readOnly) {
            return;
        }

        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateProjectFormOutput>) {
        // Optional: log errors or show notification
    }

    return (
        <div className="flex flex-col gap-6">
            <form
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onValid, onInvalid)(event);
                }}
                className="flex flex-col gap-6"
            >
                <fieldset
                    disabled={readOnly}
                    className="m-0 flex min-w-0 flex-col gap-6 border-0 p-0"
                >
                    <InfoBlock
                        titleWidth={150}
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
                                    id="name"
                                    {...name}
                                    placeholder="E.g. My Project"
                                    aria-invalid={isNameInvalid}
                                />
                                <FieldError errors={[errors.name]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={150}
                        title={<LabelWithInfo label="Environments" />}
                    >
                        <FieldGroup>
                            <Field>
                                <ProjectEnvInput
                                    envs={envs}
                                    onCreate={handleCreateEnv}
                                    onDelete={handleDeleteEnv}
                                    onUpdateColor={handleUpdateEnvColor}
                                    placeholder="Enter env"
                                    disabled={readOnly}
                                />
                                <FieldError errors={[errors.envs]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={150}
                        title={<LabelWithInfo label="Note" />}
                    >
                        <FieldGroup>
                            <Field>
                                <Textarea
                                    className="min-h-[120px]"
                                    id="notes"
                                    {...note}
                                    placeholder=""
                                    rows={6}
                                    aria-invalid={isNoteInvalid}
                                />
                                <FieldError errors={[errors.note]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={150}
                        title={<LabelWithInfo label="Tags" />}
                    >
                        <FieldGroup>
                            <Field>
                                <TagInput
                                    tags={tags}
                                    onCreate={handleCreateTag}
                                    onDelete={handleDeleteTag}
                                    placeholder="Enter tag"
                                    disabled={readOnly}
                                />
                                <FieldError errors={[errors.tags]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>
                </fieldset>

                <Field>
                    <div className="flex justify-end">
                        <Button
                            type="submit"
                            isLoading={isPending}
                            disabled={readOnly}
                        >
                            Create Project
                        </Button>
                    </div>
                </Field>
            </form>
        </div>
    );
}

interface Props {
    isPending: boolean;
    readOnly?: boolean;
    onSubmit: (values: CreateProjectFormOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
}

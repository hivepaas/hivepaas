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

export function CreateProjectForm({ isPending, onSubmit, onHasChanges }: Props) {
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
        onHasChanges?.(isDirty);
    }, [isDirty, onHasChanges]);

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

    function handleCreateEnv(env: ProjectEnvEntity) {
        if (!envs.some(item => item.name === env.name)) {
            setValue("envs", [...envs, env], { shouldDirty: true });
        }
    }

    function handleDeleteEnv(envName: string) {
        setValue(
            "envs",
            envs.filter(env => env.name !== envName),
            { shouldDirty: true },
        );
    }

    function handleUpdateEnvColor(envName: string, color: string) {
        setValue(
            "envs",
            envs.map(env => (env.name === envName ? { ...env, color } : env)),
            { shouldDirty: true },
        );
    }

    function onValid(values: CreateProjectFormOutput) {
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
                            />
                            <FieldError errors={[errors.tags]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                <Field>
                    <div className="flex justify-end">
                        <Button
                            type="submit"
                            isLoading={isPending}
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
    onSubmit: (values: CreateProjectFormOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
}

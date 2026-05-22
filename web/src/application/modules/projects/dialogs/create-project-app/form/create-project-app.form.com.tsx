import React, { useEffect, useMemo } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { type ProjectEnvEntity } from "~/projects/domain";
import { ProjectEnvBadge } from "~/projects/module-shared/components";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger } from "@/components/ui/select";
import { TagInput } from "@/components/ui/tag-input";
import { Textarea } from "@/components/ui/textarea";

import {
    type CreateProjectAppFormInput,
    type CreateProjectAppFormOutput,
    createCreateProjectAppFormSchema,
} from "../schemas";

export function CreateProjectAppForm({ envs, isPending, onSubmit, onHasChanges }: Props) {
    const defaultEnv = envs[0]?.name ?? "";
    const envNames = useMemo(() => envs.map(env => env.name), [envs]);
    const schema = useMemo(() => createCreateProjectAppFormSchema(envNames), [envNames]);

    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
        watch,
        setValue,
    } = useForm<CreateProjectAppFormInput, unknown, CreateProjectAppFormOutput>({
        defaultValues: {
            name: "",
            env: defaultEnv,
            note: "",
            tags: [],
        },
        resolver: zodResolver(schema),
        mode: "onSubmit",
    });

    useEffect(() => {
        onHasChanges?.(isDirty);
    }, [isDirty, onHasChanges]);

    const selectedEnvName = watch("env");
    const tags = watch("tags");
    const selectedEnv = envs.find(env => env.name === selectedEnvName);

    useEffect(() => {
        const nextEnv = envs[0]?.name ?? "";
        const currentEnvExists = selectedEnvName === "" || envs.some(env => env.name === selectedEnvName);

        if ((!selectedEnvName || !currentEnvExists) && selectedEnvName !== nextEnv) {
            setValue("env", nextEnv, { shouldDirty: false });
        }
    }, [envs, selectedEnvName, setValue]);

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({
        name: "name",
        control,
    });

    const {
        field: env,
        fieldState: { invalid: isEnvInvalid },
    } = useController({
        name: "env",
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

    function onValid(values: CreateProjectAppFormOutput) {
        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateProjectAppFormOutput>) {
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
                                placeholder="E.g. My App"
                                aria-invalid={isNameInvalid}
                            />
                            <FieldError errors={[errors.name]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                <InfoBlock
                    titleWidth={150}
                    title={<LabelWithInfo label="Environment" />}
                >
                    <FieldGroup>
                        <Field>
                            <Select
                                value={env.value}
                                onValueChange={env.onChange}
                                disabled={envs.length === 0}
                            >
                                <SelectTrigger
                                    aria-invalid={isEnvInvalid}
                                    className="px-2"
                                >
                                    {selectedEnv ? (
                                        <ProjectEnvBadge
                                            name={selectedEnv.name}
                                            color={selectedEnv.color}
                                        />
                                    ) : (
                                        <span className="text-muted-foreground">
                                            {envs.length === 0 ? "No environments" : "Select environment"}
                                        </span>
                                    )}
                                </SelectTrigger>
                                <SelectContent>
                                    {envs.map(projectEnv => (
                                        <SelectItem
                                            key={projectEnv.name}
                                            value={projectEnv.name}
                                        >
                                            <ProjectEnvBadge
                                                name={projectEnv.name}
                                                color={projectEnv.color}
                                            />
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            <FieldError errors={[errors.env]} />
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
                                id="note"
                                className="min-h-[120px]"
                                {...note}
                                placeholder=""
                                rows={4}
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
                            Create App
                        </Button>
                    </div>
                </Field>
            </form>
        </div>
    );
}

interface Props {
    envs: ProjectEnvEntity[];
    isPending: boolean;
    onSubmit: (values: CreateProjectAppFormOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
}

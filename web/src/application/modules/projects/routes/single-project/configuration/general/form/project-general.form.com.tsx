import * as React from "react";
import { type PropsWithChildren, useImperativeHandle } from "react";

import { FieldError, Input, TagInput } from "@components/ui";
import { Avatar } from "@components/ui/avatar";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { FormProvider, useController, useForm } from "react-hook-form";
import { type ProjectDetailsEntity, type ProjectEnvEntity } from "~/projects/domain";

import { InfoBlock } from "@application/shared/components";

import { ProjectEnvInput, ProjectStatusBadge } from "@application/modules/projects/module-shared/components";

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
            envs: defaultValues.envs,
            tags: defaultValues.tags,
            note: defaultValues.note,
            ownerId: defaultValues.owner.id,
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
    const envs = watch("envs");
    const ownerOptions = [
        ...(!defaultValues.userAccesses.some(user => user.id === defaultValues.owner.id) ? [defaultValues.owner] : []),
        ...defaultValues.userAccesses,
    ];

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
    const {
        field: ownerId,
        fieldState: { invalid: isOwnerInvalid },
    } = useController({
        control,
        name: "ownerId",
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

                    {/* Project Owner */}
                    <InfoBlock title="Project Owner">
                        <div className="max-w-[400px]">
                            <Select
                                value={ownerId.value}
                                onValueChange={ownerId.onChange}
                                disabled={ownerOptions.length === 0}
                            >
                                <SelectTrigger aria-invalid={isOwnerInvalid}>
                                    <SelectValue placeholder="Select project owner">
                                        {ownerOptions
                                            .filter(user => user.id === ownerId.value)
                                            .map(user => (
                                                <div
                                                    key={user.id}
                                                    className="flex min-w-0 items-center gap-2"
                                                >
                                                    <Avatar
                                                        name={user.fullName}
                                                        src={user.photo}
                                                        className="size-6"
                                                        borderless
                                                    />
                                                    <span className="truncate">{user.fullName}</span>
                                                </div>
                                            ))}
                                    </SelectValue>
                                </SelectTrigger>
                                <SelectContent>
                                    {ownerOptions.map(user => (
                                        <SelectItem
                                            key={user.id}
                                            value={user.id}
                                        >
                                            <div className="flex min-w-0 items-center gap-2">
                                                <Avatar
                                                    name={user.fullName}
                                                    src={user.photo}
                                                    className="size-6"
                                                    borderless
                                                />
                                                <div className="flex min-w-0 flex-col">
                                                    <span className="truncate">{user.fullName}</span>
                                                    <span className="truncate text-xs text-muted-foreground">
                                                        {user.email || user.username}
                                                    </span>
                                                </div>
                                            </div>
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            <FieldError errors={[errors.ownerId]} />
                        </div>
                    </InfoBlock>

                    {/* Environments */}
                    <InfoBlock title="Environments">
                        <div>
                            <ProjectEnvInput
                                envs={envs}
                                onCreate={handleCreateEnv}
                                onDelete={handleDeleteEnv}
                                onUpdateColor={handleUpdateEnvColor}
                                placeholder="Enter env"
                            />
                            <FieldError errors={[errors.envs]} />
                        </div>
                    </InfoBlock>

                    {/* Tags */}
                    <InfoBlock title="Tags">
                        <div>
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

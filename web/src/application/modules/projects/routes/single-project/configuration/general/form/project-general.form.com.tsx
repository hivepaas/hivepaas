import * as React from "react";
import { type PropsWithChildren, useImperativeHandle, useMemo, useState } from "react";

import { Button, FieldError, Input, TagInput } from "@components/ui";
import { Avatar } from "@components/ui/avatar";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { ImageService } from "@infrastructure/services";
import { Pencil } from "lucide-react";
import { FormProvider, useController, useForm } from "react-hook-form";
import { toast } from "sonner";
import { type ProjectDetailsEntity, type ProjectEnvEntity } from "~/projects/domain";

import { Combobox, type ComboboxOption, InfoBlock } from "@application/shared/components";
import { UsersPublicQueries } from "@application/shared/data-public/queries";
import { PhotoUploadDialog } from "@application/shared/dialogs";

import { ProjectEnvInput, ProjectStatusBadge } from "@application/modules/projects/module-shared/components";

import {
    ProjectGeneralFormSchema,
    type ProjectGeneralFormSchemaInput,
    type ProjectGeneralFormSchemaOutput,
} from "../schemas";
import { type ProjectGeneralFormRef } from "../types";

type OwnerOption = Record<string, unknown> & {
    id: string;
    username: string;
    email: string;
    fullName: string;
    photo: string | null;
};

function getOwnerDisplayName(user: Pick<OwnerOption, "email" | "fullName" | "username">) {
    return user.fullName || user.email || user.username;
}

function getOwnerIdentity(user: Pick<OwnerOption, "email" | "username">) {
    return Array.from(new Set([user.username, user.email].filter(Boolean))).join(" • ");
}

function toOwnerOption(user: Pick<OwnerOption, "id" | "username" | "email" | "fullName" | "photo">): OwnerOption {
    return {
        id: user.id,
        username: user.username,
        email: user.email,
        fullName: user.fullName,
        photo: user.photo,
    };
}

function OwnerSelectedOption({ user }: { user: OwnerOption }) {
    return (
        <div className="flex min-w-0 items-center gap-2">
            <Avatar
                name={getOwnerDisplayName(user)}
                src={user.photo}
                className="size-6"
                borderless
            />
            <span className="truncate">{getOwnerDisplayName(user)}</span>
        </div>
    );
}

function OwnerOptionRow({ user }: { user: OwnerOption }) {
    const identity = getOwnerIdentity(user);

    return (
        <div className="flex min-w-0 flex-1 items-center gap-2">
            <Avatar
                name={getOwnerDisplayName(user)}
                src={user.photo}
                className="size-6"
                borderless
            />
            <div className="flex min-w-0 flex-col">
                <span className="truncate">{getOwnerDisplayName(user)}</span>
                {identity && <span className="truncate text-xs text-muted-foreground">{identity}</span>}
            </div>
        </div>
    );
}

export function ProjectGeneralForm({ ref, defaultValues, onSubmit, readOnly = false, children }: Props) {
    const [openPhotoUpload, setOpenPhotoUpload] = useState(false);
    const [ownerSearch, setOwnerSearch] = useState("");
    const [selectedOwner, setSelectedOwner] = useState<OwnerOption>(() => toOwnerOption(defaultValues.owner));

    const { data: ownerUsersData, isFetching: isFetchingOwnerUsers } = UsersPublicQueries.useFindManyBase({
        search: ownerSearch,
    });

    const methods = useForm<ProjectGeneralFormSchemaInput, unknown, ProjectGeneralFormSchemaOutput>({
        defaultValues: {
            photo: defaultValues.photo === "" ? null : defaultValues.photo,
            photoUpload: null,
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
    const photoUrl = watch("photo");
    const photoPreviewUrl = photoUrl === "" ? null : photoUrl;
    const ownerUsers = ownerUsersData?.data;
    const ownerOptions = useMemo<OwnerOption[]>(() => {
        const optionMap = new Map<string, OwnerOption>();
        const currentOwner = toOwnerOption(defaultValues.owner);

        optionMap.set(selectedOwner.id, selectedOwner);
        optionMap.set(currentOwner.id, currentOwner);

        (ownerUsers ?? []).forEach(user => {
            optionMap.set(user.id, toOwnerOption(user));
        });

        return Array.from(optionMap.values());
    }, [defaultValues.owner, ownerUsers, selectedOwner]);

    const ownerComboboxOptions = useMemo<ComboboxOption<OwnerOption>[]>(
        () =>
            ownerOptions.map(user => ({
                value: user,
                label: getOwnerDisplayName(user),
            })),
        [ownerOptions],
    );

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

    async function handlePhotoUpload(result: File | null) {
        if (readOnly) {
            return;
        }

        if (!result) {
            setValue("photo", null, { shouldDirty: true });
            setValue("photoUpload", { delete: true }, { shouldDirty: true });
            return;
        }

        try {
            const base64String = await ImageService.convertFileToBase64(result);

            setValue("photo", base64String, { shouldDirty: true });
            setValue(
                "photoUpload",
                {
                    fileName: result.name,
                    dataBase64: base64String,
                },
                { shouldDirty: true },
            );
        } catch (error) {
            console.error("Error converting file to base64:", error);
            toast.error("Failed to process image");
        }
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
                        {/* Photo */}
                        <InfoBlock title="Photo">
                            <div className="relative size-24 rounded-full border">
                                <Avatar
                                    key={photoPreviewUrl ?? "no-photo"}
                                    name={defaultValues.name}
                                    className="size-full text-2xl"
                                    src={photoPreviewUrl}
                                />
                                <Button
                                    type="button"
                                    size="icon-sm"
                                    className="absolute -bottom-1 -right-1 rounded-full border"
                                    onClick={() => {
                                        if (readOnly) {
                                            return;
                                        }

                                        setOpenPhotoUpload(true);
                                    }}
                                    disabled={readOnly}
                                    aria-label="Edit photo"
                                    title="Edit photo"
                                >
                                    <Pencil />
                                </Button>
                            </div>
                        </InfoBlock>

                        {/* ID - Read Only */}
                        <InfoBlock title="ID">
                            <Input
                                value={defaultValues.id}
                                type="text"
                                className="max-w-[400px]"
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

                        {/* Owner */}
                        <InfoBlock title="Owner">
                            <div className="max-w-[400px]">
                                <Combobox
                                    options={ownerComboboxOptions}
                                    value={ownerId.value}
                                    onChange={(value, option) => {
                                        if (readOnly || !value || !option) {
                                            return;
                                        }

                                        setSelectedOwner(option);
                                        ownerId.onChange(value);
                                    }}
                                    onSearch={setOwnerSearch}
                                    placeholder="Select project owner"
                                    searchable
                                    emptyText="No users available"
                                    valueKey="id"
                                    loading={isFetchingOwnerUsers}
                                    disabled={readOnly || (ownerOptions.length === 0 && !isFetchingOwnerUsers)}
                                    aria-invalid={isOwnerInvalid}
                                    renderSelectedOption={option => <OwnerSelectedOption user={option.value} />}
                                    renderOption={option => <OwnerOptionRow user={option.value} />}
                                />
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
                                    disabled={readOnly}
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
                                className="w-[100%] min-h-[120px]"
                                placeholder="Enter project notes"
                                rows={4}
                                aria-invalid={isNoteInvalid}
                            />
                            <FieldError errors={[errors.note]} />
                        </InfoBlock>

                        {children}
                    </fieldset>
                </form>
            </FormProvider>
            <PhotoUploadDialog
                open={openPhotoUpload}
                onOpenChange={setOpenPhotoUpload}
                onSubmit={result => {
                    void handlePhotoUpload(result);
                }}
                initialImage={photoPreviewUrl}
                title="Project Photo"
                description="Adjust the crop area. Use tools to rotate/zoom."
                filename="project-photo"
            />
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<ProjectGeneralFormRef>;
    defaultValues: ProjectDetailsEntity;
    onSubmit: (values: ProjectGeneralFormSchemaOutput) => void;
    readOnly?: boolean;
}>;

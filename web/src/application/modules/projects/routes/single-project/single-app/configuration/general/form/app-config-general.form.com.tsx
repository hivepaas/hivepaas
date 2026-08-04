import * as React from "react";
import { type PropsWithChildren, useImperativeHandle, useState } from "react";

import { Button, FieldError, Input, TagInput } from "@components/ui";
import { Avatar } from "@components/ui/avatar";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { ImageService } from "@infrastructure/services";
import { Pencil } from "lucide-react";
import { FormProvider, useController, useForm } from "react-hook-form";
import { toast } from "sonner";
import { ProjectAppsCommands } from "~/projects/data/commands";
import { DetectAppIconDialog } from "~/projects/dialogs/detect-app-icon";
import { type ProjectAppDetails, type ProjectEnvEntity } from "~/projects/domain";
import { ProjectAppStatusBadge, ProjectEnvBadge } from "~/projects/module-shared/components";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock } from "@application/shared/components";
import { PhotoUploadDialog } from "@application/shared/dialogs";

import {
    AppConfigGeneralFormSchema,
    type AppConfigGeneralFormSchemaInput,
    type AppConfigGeneralFormSchemaOutput,
} from "../schemas";
import { type AppConfigGeneralFormRef } from "../types";

export function AppConfigGeneralForm({
    ref,
    defaultValues,
    envs,
    projectID,
    env,
    appID,
    onSubmit,
    readOnly = false,
    children,
}: Props) {
    const [openPhotoUpload, setOpenPhotoUpload] = useState(false);
    const [detectedIconUrl, setDetectedIconUrl] = useState<string | null>(null);
    const [openDetectConfirm, setOpenDetectConfirm] = useState(false);

    const { mutate: detectPhoto, isPending: isDetecting } = ProjectAppsCommands.useDetectPhoto({});
    const { mutate: updatePhoto, isPending: isApplyingPresetIcon } = ProjectAppsCommands.useUpdatePhoto({});

    const methods = useForm<AppConfigGeneralFormSchemaInput, unknown, AppConfigGeneralFormSchemaOutput>({
        defaultValues: {
            photo: defaultValues.photo === "" ? null : defaultValues.photo,
            photoUpload: null,
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
    const photoUrl = watch("photo");
    const photoPreviewUrl = photoUrl === "" ? null : photoUrl;
    const selectedEnv = envs.find(projectEnv => projectEnv.name === defaultValues.env);
    const parentAppName = defaultValues.parentApp
        ? defaultValues.parentApp.name.trim() || defaultValues.parentApp.id
        : null;

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

    function handleAutoDetectIcon() {
        if (readOnly || isDetecting) {
            return;
        }

        detectPhoto(
            {
                projectID,
                env,
                appID,
            },
            {
                onSuccess: ({ data: { url } }) => {
                    setDetectedIconUrl(url || null);
                    setOpenDetectConfirm(true);
                },
            },
        );
    }

    function handleApplyDetectedIcon() {
        if (!detectedIconUrl || readOnly) {
            return;
        }

        updatePhoto(
            {
                projectID,
                env,
                appID,
                photo: {
                    fileName: detectedIconUrl,
                    isPresetIcon: true,
                },
            },
            {
                onSuccess: () => {
                    setValue("photo", detectedIconUrl, { shouldDirty: false });
                    setValue("photoUpload", null, { shouldDirty: false });
                    setOpenDetectConfirm(false);
                    setDetectedIconUrl(null);
                    toast.success("App icon updated");
                },
            },
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
                        {/* Photo */}
                        <InfoBlock title="Photo">
                            <div className="flex items-center gap-4">
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
                                <Button
                                    type="button"
                                    variant="link"
                                    className="h-auto p-0"
                                    onClick={handleAutoDetectIcon}
                                    disabled={readOnly || isDetecting}
                                >
                                    Auto Detect Icon
                                </Button>
                            </div>
                        </InfoBlock>

                        {parentAppName && (
                            <InfoBlock title="Parent App">
                                <Input
                                    value={parentAppName}
                                    type="text"
                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                    disabled
                                    readOnly
                                />
                            </InfoBlock>
                        )}

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
            <PhotoUploadDialog
                open={openPhotoUpload}
                onOpenChange={setOpenPhotoUpload}
                onSubmit={result => {
                    void handlePhotoUpload(result);
                }}
                initialImage={photoPreviewUrl}
                title="App Photo"
                description="Adjust the crop area. Use tools to rotate/zoom."
                filename="app-photo"
            />
            <DetectAppIconDialog
                open={openDetectConfirm}
                onOpenChange={open => {
                    setOpenDetectConfirm(open);
                    if (!open) {
                        setDetectedIconUrl(null);
                    }
                }}
                appName={defaultValues.name}
                iconUrl={detectedIconUrl}
                isApplying={isApplyingPresetIcon}
                onUseIt={handleApplyDetectedIcon}
            />
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigGeneralFormRef>;
    defaultValues: ProjectAppDetails;
    envs: ProjectEnvEntity[];
    projectID: string;
    env: string;
    appID: string;
    onSubmit: (values: AppConfigGeneralFormSchemaOutput) => void;
    readOnly?: boolean;
}>;

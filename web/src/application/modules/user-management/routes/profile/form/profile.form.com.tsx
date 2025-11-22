import React, { type PropsWithChildren, useImperativeHandle, useState } from "react";

import { Avatar, Button } from "@components/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { ImageService } from "@infrastructure/services";
import { format } from "date-fns";
import { Pencil } from "lucide-react";
import { type FieldErrors, FormProvider, useForm } from "react-hook-form";
import { toast } from "sonner";
import { UserRoleBadge, UserSecurityBadge } from "~/user-management/module-shared/components";
import { UserInput } from "~/user-management/module-shared/form/user-input";
import { mapModuleAccesses } from "~/user-management/module-shared/utils";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { PhotoUploadDialog } from "@application/shared/dialogs";
import { type Profile } from "@application/shared/entities";
import { ESecuritySettings, EUserRole } from "@application/shared/enums";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { Information, Security } from "../form-components";
import { ProfileFormSchema, type ProfileFormSchemaInput, type ProfileFormSchemaOutput } from "../schemas";
import { type ProfileFormRef } from "../types";

const DEFAULTS: ProfileFormSchemaInput = {
    fullName: "",
    email: "",
    username: "",
    position: "",
    notes: "",
    role: EUserRole.Member,
    accessExpireAt: null,
    securityOption: ESecuritySettings.PasswordOnly,
    projectAccesses: [],
    moduleAccesses: [],
    photo: null,
    photoUpload: null,
};

type SchemaInput = ProfileFormSchemaInput;
type SchemaOutput = ProfileFormSchemaOutput;

export function ProfileForm({ ref, defaultValues, onSubmit, children }: Props) {
    const [openPhotoUpload, setOpenPhotoUpload] = useState(false);

    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: {
            ...DEFAULTS,
            ...defaultValues,
            fullName: defaultValues.fullName ?? "",
            email: defaultValues.email ?? "",
            moduleAccesses: mapModuleAccesses(defaultValues.moduleAccesses),
            photo: defaultValues.photo,
            photoUpload: null,
        },
        resolver: zodResolver(ProfileFormSchema),
        mode: "onSubmit",
    });

    const { isDirty } = methods.formState;

    function onValid(values: SchemaOutput) {
        if (!isDirty) {
            toast.info("No changes to save");
            return;
        }

        onSubmit({
            ...values,
        });
    }

    function onInvalid(errors: FieldErrors<SchemaInput>) {
        console.error(errors);
    }

    useImperativeHandle(
        ref,
        () => ({
            setValues: (values: SchemaInput) => {
                methods.reset({
                    ...DEFAULTS,
                    ...values,
                    moduleAccesses: mapModuleAccesses(values.moduleAccesses),
                });
            },
            onError(_error: ValidationException) {
                // TODO handle validation error
            },
        }),
        [methods],
    );

    const isAdmin = methods.watch("role") === EUserRole.Admin;

    const photoUrl = methods.watch("photo");

    async function handlePhotoUpload(result: File | null) {
        console.log(result);
        if (!result) {
            methods.setValue("photo", null, { shouldDirty: true });
            methods.setValue("photoUpload", null, { shouldDirty: true });
            return;
        }

        try {
            const base64String = await ImageService.convertFileToBase64(result);

            methods.setValue("photo", base64String, { shouldDirty: true });

            methods.setValue(
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
        <div className="single-user-form">
            <FormProvider {...methods}>
                <form
                    onSubmit={event => {
                        event.preventDefault();

                        void methods.handleSubmit(onValid, onInvalid)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    <InfoBlock title="Photo">
                        <div className="relative size-24 rounded-full border">
                            <Avatar
                                key={photoUrl ?? "no-photo"}
                                name={defaultValues.fullName ?? ""}
                                className="size-full"
                                src={photoUrl}
                            />
                            <Button
                                type="button"
                                size="icon-sm"
                                className="absolute -bottom-1 -right-1 rounded-full border"
                                onClick={() => {
                                    setOpenPhotoUpload(true);
                                }}
                                aria-label="Edit photo"
                                title="Edit photo"
                            >
                                <Pencil />
                            </Button>
                        </div>
                    </InfoBlock>
                    <Information />

                    {/* Role */}
                    <div className="h-px bg-border" />
                    <InfoBlock
                        titleWidth={150}
                        title="Role"
                    >
                        <UserRoleBadge role={defaultValues.role} />
                    </InfoBlock>

                    {/* Joining date */}
                    <div className="h-px bg-border" />
                    <InfoBlock title={<LabelWithInfo label="Joining date" />}>
                        <span className="text-sm">{format(defaultValues.createdAt, "yyyy-MM-dd HH:mm:ss")}</span>
                    </InfoBlock>

                    {/* Access Expiration */}
                    <div className="h-px bg-border" />
                    <InfoBlock title="Access Expiration">
                        {defaultValues.accessExpireAt
                            ? format(defaultValues.accessExpireAt, "yyyy-MM-dd HH:mm:ss")
                            : "-"}
                    </InfoBlock>

                    {/* Security Option */}
                    <div className="h-px bg-border" />
                    <InfoBlock title="Security Option">
                        <UserSecurityBadge securityOption={defaultValues.securityOption} />
                    </InfoBlock>

                    {/* Security */}
                    <Security defaultValues={defaultValues} />

                    {/* Project Access */}
                    <div className="h-px bg-border" />
                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Project access"
                                content="Project access description"
                            />
                        }
                    >
                        <UserInput.ProjectAccess<ProfileFormSchemaInput>
                            name="projectAccesses"
                            isAdmin={isAdmin}
                            disabled
                        />
                    </InfoBlock>

                    {/* Module Access */}
                    <div className="h-px bg-border" />
                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Module access"
                                content="Module access description"
                            />
                        }
                    >
                        <UserInput.ModuleAccess<ProfileFormSchemaInput>
                            name="moduleAccesses"
                            isAdmin={isAdmin}
                            disabled
                        />
                    </InfoBlock>
                    {children}
                </form>
            </FormProvider>
            <PhotoUploadDialog
                open={openPhotoUpload}
                onOpenChange={setOpenPhotoUpload}
                onSubmit={result => {
                    void handlePhotoUpload(result);
                }}
                initialImage={photoUrl}
            />
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<ProfileFormRef>;
    defaultValues: Profile;
    onSubmit: (values: SchemaOutput) => void;
}>;

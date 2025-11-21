import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { format } from "date-fns";
import { type FieldErrors, FormProvider, useForm } from "react-hook-form";
import type { UserBase } from "~/user-management/domain";
import { UserRoleBadge, UserSecurityBadge } from "~/user-management/module-shared/components";
import { UserInput } from "~/user-management/module-shared/form/user-input";
import { mapModuleAccesses } from "~/user-management/module-shared/utils";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
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
};

type SchemaInput = ProfileFormSchemaInput;
type SchemaOutput = ProfileFormSchemaOutput;

export function ProfileForm({ ref, defaultValues, onSubmit, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: {
            ...DEFAULTS,
            ...defaultValues,
            moduleAccesses: mapModuleAccesses(defaultValues.moduleAccesses),
        },
        resolver: zodResolver(ProfileFormSchema),
        mode: "onSubmit",
    });

    function onValid(values: SchemaOutput) {
        onSubmit(values);
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
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<ProfileFormRef>;
    defaultValues: UserBase;
    onSubmit: (values: SchemaOutput) => void;
}>;

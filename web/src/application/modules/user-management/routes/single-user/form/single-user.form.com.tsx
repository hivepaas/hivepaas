import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { format } from "date-fns";
import { type FieldErrors, FormProvider, useForm } from "react-hook-form";
import { mapModuleAccesses } from "~/user-management/module-shared/utils/map-module-accesses";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ESecuritySettings, EUserRole } from "@application/shared/enums";

import { UserInput } from "@application/modules/user-management/module-shared/form/user-input";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { Information } from "../form-components";
import { SingleUserFormSchema, type SingleUserFormSchemaInput, type SingleUserFormSchemaOutput } from "../schemas";
import { type SingleUserFormRef } from "../types";

const DEFAULTS: SingleUserFormSchemaInput = {
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

type SchemaInput = SingleUserFormSchemaInput;
type SchemaOutput = SingleUserFormSchemaOutput;

export function SingleUserForm({ ref, defaultValues, onSubmit, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: {
            ...DEFAULTS,
            ...defaultValues,
            moduleAccesses: mapModuleAccesses(defaultValues.moduleAccesses),
        },
        resolver: zodResolver(SingleUserFormSchema),
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
            onError(error: ValidationException) {
                if (error.errors.length === 0) {
                    return;
                }

                error.errors.forEach(({ path, message }, index) => {
                    methods.setError(
                        path as keyof SchemaInput,
                        { message, type: "manual" },
                        { shouldFocus: index === 0 },
                    );
                });
            },
        }),
        [methods],
    );

    const isAdmin = methods.watch("role") === EUserRole.Admin;

    return (
        <div className="pt-2">
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
                    <InfoBlock title="Role">
                        <UserInput.Role<SingleUserFormSchemaInput> name="role" />
                    </InfoBlock>

                    {/* Joining date */}
                    <div className="h-px bg-border" />
                    <InfoBlock title={<LabelWithInfo label="Joining Date" />}>
                        <span className="text-sm">{format(defaultValues.createdAt, "yyyy-MM-dd HH:mm:ss")}</span>
                    </InfoBlock>

                    {/* Access Expiration */}
                    <div className="h-px bg-border" />
                    <InfoBlock title="Access Expiration">
                        <UserInput.AccessExpiration<SingleUserFormSchemaInput>
                            className="md:min-w-[400px] w-fit"
                            name="accessExpireAt"
                        />
                    </InfoBlock>

                    {/* Security Option */}
                    <div className="h-px bg-border" />
                    <InfoBlock title="Security Option">
                        <UserInput.SecurityOption<SingleUserFormSchemaInput> name="securityOption" />
                    </InfoBlock>

                    {/* Project Access */}
                    <div className="h-px bg-border" />
                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Project Access"
                                content="Project access description"
                            />
                        }
                    >
                        <UserInput.ProjectAccess<SingleUserFormSchemaInput>
                            name="projectAccesses"
                            isAdmin={isAdmin}
                        />
                    </InfoBlock>

                    {/* Module Access */}
                    <div className="h-px bg-border" />
                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Module Access"
                                content="Module access description"
                            />
                        }
                    >
                        <UserInput.ModuleAccess<SingleUserFormSchemaInput>
                            name="moduleAccesses"
                            isAdmin={isAdmin}
                        />
                    </InfoBlock>
                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<SingleUserFormRef>;
    defaultValues: Partial<SchemaInput> & { createdAt: Date };
    onSubmit: (values: SchemaOutput) => void;
}>;

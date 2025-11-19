import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { format } from "date-fns";
import { type FieldErrors, FormProvider, useForm } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { MODULES } from "@application/shared/constants";
import { ESecuritySettings, EUserRole } from "@application/shared/enums";

import { UserInput } from "@application/modules/user-management/module-shared/form/user-input";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { AccessExpiration, Information, Role, SecurityOption } from "../form-components";
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

const DEFAULT_ACCESS = {
    read: false,
    write: false,
    delete: false,
} as const;

type SchemaInput = SingleUserFormSchemaInput;
type SchemaOutput = SingleUserFormSchemaOutput;

function mergeModuleAccess(existingModuleAccess?: SchemaInput["moduleAccesses"]): SchemaInput["moduleAccesses"] {
    return MODULES.map(module => {
        const existingModule = existingModuleAccess?.find(m => m.id === module.id);
        return (
            existingModule ?? {
                id: module.id,
                name: module.name,
                access: DEFAULT_ACCESS,
            }
        );
    });
}

export function SingleUserForm({ ref, defaultValues, onSubmit, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: {
            ...DEFAULTS,
            ...defaultValues,
            moduleAccesses: mergeModuleAccess(defaultValues.moduleAccesses),
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
                    moduleAccesses: mergeModuleAccess(values.moduleAccesses),
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
                    <Role />

                    {/* Joining date */}
                    <div className="h-px bg-border" />
                    <InfoBlock title={<LabelWithInfo label="Joining date" />}>
                        <span>{format(defaultValues.createdAt, "yyyy-MM-dd HH:mm:ss")}</span>
                    </InfoBlock>

                    {/* Access Expiration */}
                    <div className="h-px bg-border" />
                    <AccessExpiration />

                    {/* Security Option */}
                    <div className="h-px bg-border" />
                    <SecurityOption />

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
                                label="Module access"
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

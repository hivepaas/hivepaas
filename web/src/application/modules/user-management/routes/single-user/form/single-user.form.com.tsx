import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useForm } from "react-hook-form";

import { MODULES } from "@application/shared/constants";
import { ESecuritySettings, EUserRole } from "@application/shared/enums";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { AccessExpiration, Information, ModuleAccess, ProjectAccess, Role, SecurityOption } from "../form-components";
import { SingleUserFormSchema, type SingleUserFormSchemaInput, type SingleUserFormSchemaOutput } from "../schemas";
import { type SingleUserFormRef } from "../types";

const DEFAULTS: SingleUserFormSchemaInput = {
    fullName: "",
    email: "",
    username: "",
    position: "",
    role: EUserRole.Member,
    accessExpireAt: null,
    securityOption: ESecuritySettings.PasswordOnly,
    projectAccess: [],
    moduleAccess: [],
};

const DEFAULT_ACCESS = {
    read: false,
    write: false,
    delete: false,
} as const;

type SchemaInput = SingleUserFormSchemaInput;
type SchemaOutput = SingleUserFormSchemaOutput;

function mergeModuleAccess(existingModuleAccess?: SchemaInput["moduleAccess"]): SchemaInput["moduleAccess"] {
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
            moduleAccess: mergeModuleAccess(defaultValues.moduleAccess),
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
                    moduleAccess: mergeModuleAccess(values.moduleAccess),
                });
            },
            onError(_error: ValidationException) {
                // TODO handle validation error
            },
        }),
        [methods],
    );

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

                    {/* Access Expiration */}
                    <div className="h-px bg-border" />
                    <AccessExpiration />

                    {/* Security Option */}
                    <div className="h-px bg-border" />
                    <SecurityOption />

                    {/* Project Access */}
                    <div className="h-px bg-border" />
                    <ProjectAccess />

                    {/* Module Access */}
                    <div className="h-px bg-border" />
                    <ModuleAccess />
                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<SingleUserFormRef>;
    defaultValues: Partial<SchemaInput>;
    onSubmit: (values: SchemaOutput) => void;
}>;

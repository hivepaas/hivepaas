import React, { type PropsWithChildren, forwardRef } from "react";

import { Field, FieldError, FieldGroup, Input } from "@components/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useForm } from "react-hook-form";
import { useController } from "react-hook-form";
import { UserInput } from "~/user-management/module-shared/form/user-input";
import { AccessExpiration, Role, SecurityOption } from "~/user-management/routes/single-user/form-components";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { MODULES } from "@application/shared/constants";
import { ESecuritySettings, EUserRole } from "@application/shared/enums";

import { type InviteUserFormInput, type InviteUserFormOutput, InviteUserFormSchema } from "../schemas";

const DEFAULTS: InviteUserFormInput = {
    email: "",
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

function mergeModuleAccess(
    existingModuleAccess?: InviteUserFormInput["moduleAccess"],
): InviteUserFormInput["moduleAccess"] {
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

export const InviteUserForm = forwardRef<HTMLFormElement, Props>(
    ({ defaultValues = {}, onSubmit, onHasChanges, children }, ref) => {
        const methods = useForm<InviteUserFormInput, unknown, InviteUserFormOutput>({
            defaultValues: {
                ...DEFAULTS,
                ...defaultValues,
                moduleAccess: mergeModuleAccess(defaultValues.moduleAccess),
            },
            resolver: zodResolver(InviteUserFormSchema),
            mode: "onSubmit",
        });

        const {
            control,
            formState: { errors, isDirty },
            handleSubmit,
        } = methods;

        React.useEffect(() => {
            onHasChanges?.(isDirty);
        }, [isDirty, onHasChanges]);

        function onValid(values: InviteUserFormOutput) {
            onSubmit(values);
        }

        function onInvalid(_errors: FieldErrors<InviteUserFormInput>) {
            // Optional: log errors or show notification
        }

        const {
            field: email,
            fieldState: { invalid: isEmailInvalid },
        } = useController({
            control,
            name: "email",
        });

        return (
            <div className="invite-user-form">
                <FormProvider {...methods}>
                    <form
                        ref={ref}
                        onSubmit={event => {
                            event.preventDefault();
                            void handleSubmit(onValid, onInvalid)(event);
                        }}
                        className="flex flex-col gap-6"
                    >
                        {/* Email */}
                        <InfoBlock
                            titleWidth={150}
                            title="Email"
                        >
                            <FieldGroup>
                                <Field>
                                    <Input
                                        id="email"
                                        {...email}
                                        placeholder="abc@domain.com"
                                        aria-invalid={isEmailInvalid}
                                    />
                                    <FieldError errors={[errors.email]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>

                        {/* Role */}
                        <Role />

                        {/* Access Expiration */}
                        <AccessExpiration />

                        {/* Security Option */}
                        <SecurityOption />

                        {/* Project Access */}
                        <InfoBlock
                            title={
                                <LabelWithInfo
                                    label="Project Access"
                                    content="Project access description"
                                />
                            }
                            titleWidth={150}
                        >
                            <UserInput.ProjectAccess<InviteUserFormInput> name="projectAccess" />
                        </InfoBlock>

                        {/* Module Access */}
                        <InfoBlock
                            title={
                                <LabelWithInfo
                                    label="Module Access"
                                    content="Module access description"
                                />
                            }
                            titleWidth={150}
                        >
                            <UserInput.ModuleAccess<InviteUserFormInput> name="moduleAccess" />
                        </InfoBlock>

                        {children}
                    </form>
                </FormProvider>
            </div>
        );
    },
);

InviteUserForm.displayName = "InviteUserForm";

interface Props extends PropsWithChildren {
    defaultValues?: Partial<InviteUserFormInput>;
    onSubmit: (values: InviteUserFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
}

import React, { type PropsWithChildren, forwardRef } from "react";

import { Field, FieldError, FieldGroup, Input } from "@components/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useForm } from "react-hook-form";
import { useController } from "react-hook-form";
import { UserInput } from "~/user-management/module-shared/form/user-input";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ESecuritySettings, EUserRole } from "@application/shared/enums";

import { mapModuleAccesses } from "@application/modules/user-management/module-shared/utils";

import { type InviteUserFormInput, type InviteUserFormOutput, InviteUserFormSchema } from "../schemas";

const DEFAULTS: InviteUserFormInput = {
    email: "",
    role: EUserRole.Member,
    accessExpireAt: null,
    securityOption: ESecuritySettings.PasswordOnly,
    projectAccesses: [],
    moduleAccesses: [],
};

export const InviteUserForm = forwardRef<HTMLFormElement, Props>(
    ({ defaultValues = {}, onSubmit, onHasChanges, children }, ref) => {
        const methods = useForm<InviteUserFormInput, unknown, InviteUserFormOutput>({
            defaultValues: {
                ...DEFAULTS,
                ...defaultValues,
                moduleAccesses: mapModuleAccesses(defaultValues.moduleAccesses),
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
            methods.reset(values);
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

        const isAdmin = methods.watch("role") === EUserRole.Admin;

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
                        <InfoBlock
                            titleWidth={150}
                            title="Role"
                        >
                            <UserInput.Role<InviteUserFormInput> name="role" />
                        </InfoBlock>

                        {/* Access Expiration */}
                        <InfoBlock
                            titleWidth={150}
                            title="Access Expiration"
                        >
                            <UserInput.AccessExpiration<InviteUserFormInput>
                                name="accessExpireAt"
                                className="md:min-w-[400px] w-fit"
                            />
                        </InfoBlock>

                        {/* Security Option */}
                        <InfoBlock
                            titleWidth={150}
                            title="Security Option"
                        >
                            <UserInput.SecurityOption<InviteUserFormInput> name="securityOption" />
                        </InfoBlock>

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
                            <UserInput.ProjectAccess<InviteUserFormInput>
                                name="projectAccesses"
                                isAdmin={isAdmin}
                            />
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
                            <UserInput.ModuleAccess<InviteUserFormInput>
                                name="moduleAccesses"
                                isAdmin={isAdmin}
                            />
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

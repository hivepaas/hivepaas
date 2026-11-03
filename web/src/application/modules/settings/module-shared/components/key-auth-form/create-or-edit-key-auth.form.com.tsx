import { useEffect } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS } from "~/settings/module-shared/constants/settings-form-layout.constants";

import { AvailableInAppsWarning, FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";

import { InheritedSettingReadonlyNotice } from "../inherited-setting-readonly-notice.com";
import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";

import type {
    CreateOrEditKeyAuthFormInput,
    CreateOrEditKeyAuthFormOutput,
} from "./create-or-edit-key-auth.form.schema";
import { CreateOrEditKeyAuthFormSchema } from "./create-or-edit-key-auth.form.schema";

export function CreateOrEditKeyAuthForm({
    isPending,
    onSubmit,
    onHasChanges,
    savedVersion = 0,
    initialValues,
    showAvailableInProjects = true,
    isProjectScope = false,
    readOnlyInherited = false,
    readOnly = false,
    onClose,
}: Props) {
    const isReadOnly = readOnlyInherited || readOnly;
    const isInheritableDisabled = isReadOnly;
    const inheritableLabel = isProjectScope ? "Available in Apps" : "Available in Projects";

    const {
        handleSubmit,
        control,
        getValues,
        reset,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditKeyAuthFormInput, unknown, CreateOrEditKeyAuthFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            keyId: initialValues?.keyId ?? "",
            secretKey: initialValues?.secretKey ?? "",
            inheritable: initialValues?.inheritable ?? (isProjectScope ? true : false),
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditKeyAuthFormSchema),
        mode: "onSubmit",
    });

    useEffect(() => {
        if (savedVersion === 0) {
            return;
        }

        reset(getValues());
        onHasChanges?.(false);
    }, [getValues, onHasChanges, reset, savedVersion]);

    useEffect(() => {
        onHasChanges?.(isReadOnly ? false : isDirty);
    }, [isDirty, onHasChanges, isReadOnly]);

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const {
        field: keyId,
        fieldState: { invalid: isKeyIdInvalid },
    } = useController({ name: "keyId", control });
    const {
        field: secretKey,
        fieldState: { invalid: isSecretKeyInvalid },
    } = useController({ name: "secretKey", control });
    const { field: inheritable } = useController({ name: "inheritable", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditKeyAuthFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditKeyAuthFormOutput>) {
        console.error(_errors);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <div className="">
                {readOnlyInherited && <InheritedSettingReadonlyNotice />}
                {readOnly && !readOnlyInherited && <PermissionReadonlyNotice />}
                <fieldset
                    disabled={isReadOnly}
                    className={`flex flex-col gap-6 border-0 p-0 m-0 min-w-0 ${SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS}`}
                >
                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Name" />}
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...name}
                                    aria-invalid={isNameInvalid}
                                />
                                <FieldError errors={[errors.name]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Key ID"
                                content="The key's public half: an access key ID, such as AWS's or Cloudflare R2's."
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...keyId}
                                    aria-invalid={isKeyIdInvalid}
                                />
                                <FieldError errors={[errors.keyId]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Secret Key"
                                content="The key's secret half. It is stored encrypted, and shown only to whoever may reveal secrets."
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <PasswordInput
                                    value={secretKey.value}
                                    onChange={secretKey.onChange}
                                    aria-invalid={isSecretKeyInvalid}
                                />
                                <FieldError errors={[errors.secretKey]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    {showAvailableInProjects && (
                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label={inheritableLabel} />}
                        >
                            <div className="flex items-center gap-3">
                                <Checkbox
                                    disabled={isInheritableDisabled}
                                    checked={inheritable.value}
                                    onCheckedChange={checked => {
                                        inheritable.onChange(Boolean(checked));
                                    }}
                                />
                                {isProjectScope && !inheritable.value ? <AvailableInAppsWarning /> : null}
                            </div>
                        </InfoBlock>
                    )}

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Default" />}
                    >
                        <Checkbox
                            checked={defaultField.value}
                            onCheckedChange={checked => {
                                defaultField.onChange(Boolean(checked));
                            }}
                        />
                    </InfoBlock>
                </fieldset>
            </div>
            {!isReadOnly && (
                <FormActionBar>
                    <SettingsFormCancelAction
                        onCancel={onClose}
                        disabled={isPending}
                    />
                    <Button
                        type="submit"
                        isLoading={isPending}
                        className="min-w-[100px]"
                    >
                        Save
                    </Button>
                </FormActionBar>
            )}
            {isReadOnly && (
                <FormActionBar>
                    <Button
                        type="button"
                        onClick={onClose}
                        className="min-w-[100px]"
                    >
                        Close
                    </Button>
                </FormActionBar>
            )}
        </form>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: CreateOrEditKeyAuthFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    initialValues?: Partial<CreateOrEditKeyAuthFormInput>;
    showAvailableInProjects?: boolean;
    isProjectScope?: boolean;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}

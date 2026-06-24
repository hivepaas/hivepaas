import { useEffect } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import {
    SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS,
    SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS,
} from "~/settings/module-shared/constants/settings-form-layout.constants";
import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { EAccessTokenKind } from "@application/shared/enums";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";

import { InheritedSettingReadonlyNotice } from "../inherited-setting-readonly-notice.com";
import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";

import type {
    CreateOrEditAccessTokenFormInput,
    CreateOrEditAccessTokenFormOutput,
} from "./create-or-edit-access-token.form.schema";
import { CreateOrEditAccessTokenFormSchema } from "./create-or-edit-access-token.form.schema";

const kindOptions = Object.values(EAccessTokenKind);

export function CreateOrEditAccessTokenForm({
    isPending,
    isTesting,
    testStatus,
    onSubmit,
    onTestConnection,
    onHasChanges,
    initialValues,
    showAvailableInProjects,
    readOnlyInherited = false,
    readOnly = false,
    onClose,
}: Props) {
    const isReadOnly = readOnlyInherited || readOnly;

    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditAccessTokenFormInput, unknown, CreateOrEditAccessTokenFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            kind: initialValues?.kind ?? EAccessTokenKind.Github,
            user: initialValues?.user ?? "",
            token: initialValues?.token ?? "",
            baseURL: initialValues?.baseURL ?? "",
            expireAt: initialValues?.expireAt ?? null,
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditAccessTokenFormSchema),
        mode: "onSubmit",
    });

    useEffect(() => {
        onHasChanges?.(isReadOnly ? false : isDirty);
    }, [isDirty, onHasChanges, isReadOnly]);

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const {
        field: kind,
        fieldState: { invalid: isKindInvalid },
    } = useController({ name: "kind", control });
    const {
        field: user,
        fieldState: { invalid: isUserInvalid },
    } = useController({ name: "user", control });
    const {
        field: token,
        fieldState: { invalid: isTokenInvalid },
    } = useController({ name: "token", control });
    const {
        field: baseURL,
        fieldState: { invalid: isBaseURLInvalid },
    } = useController({ name: "baseURL", control });
    const {
        field: expireAt,
        fieldState: { invalid: isExpireAtInvalid },
    } = useController({ name: "expireAt", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditAccessTokenFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onTestValid(values: CreateOrEditAccessTokenFormOutput) {
        onTestConnection(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditAccessTokenFormOutput>) {
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
                                label="Type"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Select
                                    value={kind.value}
                                    onValueChange={kind.onChange}
                                >
                                    <SelectTrigger aria-invalid={isKindInvalid}>
                                        <SelectValue placeholder="Select type" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {kindOptions.map(option => (
                                            <SelectItem
                                                key={option}
                                                value={option}
                                            >
                                                {option}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                                <FieldError errors={[errors.kind]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="User"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...user}
                                    aria-invalid={isUserInvalid}
                                />
                                <FieldError errors={[errors.user]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Token"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <PasswordInput
                                    value={token.value}
                                    onChange={token.onChange}
                                    aria-invalid={isTokenInvalid}
                                />
                                <FieldError errors={[errors.token]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Base URL" />}
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...baseURL}
                                    aria-invalid={isBaseURLInvalid}
                                />
                                <FieldError errors={[errors.baseURL]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Access expiration" />}
                    >
                        <div className={SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS}>
                            <DateTimePicker
                                value={expireAt.value ?? undefined}
                                onChange={date => {
                                    expireAt.onChange(date ?? null);
                                }}
                                displayFormat={{ hour24: "yyyy-MM-dd HH:mm:ss" }}
                                granularity="second"
                                showClearButton
                                aria-invalid={isExpireAtInvalid}
                            />
                            <FieldError errors={[errors.expireAt]} />
                        </div>
                    </InfoBlock>

                    {showAvailableInProjects && (
                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label="Available in Projects" />}
                        >
                            <Checkbox
                                checked={availableInProjects.value}
                                onCheckedChange={checked => {
                                    availableInProjects.onChange(Boolean(checked));
                                }}
                            />
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
                <div className="pb-6 flex justify-between mt-6">
                    <div className="flex items-center gap-3">
                        <Button
                            type="button"
                            variant="secondary"
                            isLoading={isTesting}
                            onClick={() => {
                                void handleSubmit(onTestValid, onInvalid)();
                            }}
                        >
                            Test Access
                        </Button>
                        {testStatus === "succeeded" && <span className="text-sm text-green-600">Succeeded</span>}
                        {testStatus === "failed" && <span className="text-sm text-destructive">Failed</span>}
                    </div>
                    <div className="flex items-center gap-3">
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
                    </div>
                </div>
            )}
            {isReadOnly && (
                <div className="shrink-0 px-0 mt-6 pb-6 flex justify-end">
                    <Button
                        type="button"
                        onClick={onClose}
                    >
                        Close
                    </Button>
                </div>
            )}
        </form>
    );
}

interface Props {
    isPending: boolean;
    isTesting: boolean;
    testStatus: "idle" | "succeeded" | "failed";
    onSubmit: (values: CreateOrEditAccessTokenFormOutput) => void;
    onTestConnection: (values: CreateOrEditAccessTokenFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    initialValues?: Partial<CreateOrEditAccessTokenFormInput>;
    showAvailableInProjects: boolean;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}

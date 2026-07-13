import { useEffect } from "react";

import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useWatch } from "react-hook-form";
import { SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS } from "~/settings/module-shared/constants/settings-form-layout.constants";

import { FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { EImServiceKind } from "@application/shared/enums";

import {
    Button,
    Checkbox,
    Field,
    FieldError,
    FieldGroup,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui";

import { InheritedSettingReadonlyNotice } from "../inherited-setting-readonly-notice.com";
import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";

import type {
    CreateOrEditImPlatformFormInput,
    CreateOrEditImPlatformFormOutput,
} from "./create-or-edit-im-platform.form.schema";
import { CreateOrEditImPlatformFormSchema } from "./create-or-edit-im-platform.form.schema";

export function CreateOrEditImPlatformForm({
    isPending,
    isTesting,
    testStatus,
    onSubmit,
    onTestSendMsg,
    onHasChanges,
    savedVersion = 0,
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
        getValues,
        reset,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditImPlatformFormInput, unknown, CreateOrEditImPlatformFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            kind: initialValues?.kind ?? EImServiceKind.Slack,
            webhook: initialValues?.webhook ?? "",
            botToken: initialValues?.botToken ?? "",
            chatId: initialValues?.chatId ?? "",
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditImPlatformFormSchema),
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

    const kindValue = useWatch({ control, name: "kind" });

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const {
        field: kind,
        fieldState: { invalid: isKindInvalid },
    } = useController({ name: "kind", control });
    const {
        field: webhook,
        fieldState: { invalid: isWebhookInvalid },
    } = useController({ name: "webhook", control });
    const {
        field: botToken,
        fieldState: { invalid: isBotTokenInvalid },
    } = useController({ name: "botToken", control });
    const {
        field: chatId,
        fieldState: { invalid: isChatIdInvalid },
    } = useController({ name: "chatId", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: CreateOrEditImPlatformFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onTestValid(values: CreateOrEditImPlatformFormOutput) {
        onTestSendMsg(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditImPlatformFormOutput>) {
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
                    <FieldGroup>
                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label="Name" />}
                        >
                            <Field>
                                <Input
                                    {...name}
                                    aria-invalid={isNameInvalid}
                                />
                                <FieldError errors={[errors.name]} />
                            </Field>
                        </InfoBlock>

                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label="Type" />}
                        >
                            <Field>
                                <Select
                                    value={kind.value}
                                    onValueChange={value => {
                                        kind.onChange(value);
                                    }}
                                >
                                    <SelectTrigger aria-invalid={isKindInvalid}>
                                        <SelectValue placeholder="Select type" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value={EImServiceKind.Slack}>Slack Webhook</SelectItem>
                                        <SelectItem value={EImServiceKind.Discord}>Discord Webhook</SelectItem>
                                        <SelectItem value={EImServiceKind.Telegram}>Telegram Bot</SelectItem>
                                    </SelectContent>
                                </Select>
                                <FieldError errors={[errors.kind]} />
                            </Field>
                        </InfoBlock>

                        {kindValue === EImServiceKind.Telegram ? (
                            <>
                                <InfoBlock
                                    titleWidth={220}
                                    title={
                                        <LabelWithInfo
                                            label="Bot Token"
                                            isRequired
                                        />
                                    }
                                >
                                    <Field>
                                        <PasswordInput
                                            {...botToken}
                                            aria-invalid={isBotTokenInvalid}
                                            placeholder="Telegram bot token"
                                        />
                                        <FieldError errors={[errors.botToken]} />
                                    </Field>
                                </InfoBlock>

                                <InfoBlock
                                    titleWidth={220}
                                    title={
                                        <LabelWithInfo
                                            label="Chat ID"
                                            isRequired
                                        />
                                    }
                                >
                                    <Field>
                                        <Input
                                            {...chatId}
                                            aria-invalid={isChatIdInvalid}
                                            placeholder="Telegram chat ID"
                                        />
                                        <FieldError errors={[errors.chatId]} />
                                    </Field>
                                </InfoBlock>
                            </>
                        ) : (
                            <InfoBlock
                                titleWidth={220}
                                title={
                                    <LabelWithInfo
                                        label="Webhook URL"
                                        isRequired
                                    />
                                }
                            >
                                <Field>
                                    <PasswordInput
                                        {...webhook}
                                        aria-invalid={isWebhookInvalid}
                                        placeholder={
                                            kindValue === EImServiceKind.Slack
                                                ? "Slack webhook URL"
                                                : "Discord webhook URL"
                                        }
                                    />
                                    <FieldError errors={[errors.webhook]} />
                                </Field>
                            </InfoBlock>
                        )}

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
                    </FieldGroup>
                </fieldset>
            </div>
            {!isReadOnly && (
                <FormActionBar contentClassName="justify-between">
                    <div className="flex items-center gap-3">
                        <Button
                            type="button"
                            variant="secondary"
                            isLoading={isTesting}
                            onClick={() => {
                                void handleSubmit(onTestValid, onInvalid)();
                            }}
                        >
                            Test Send Msg
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
    isTesting: boolean;
    testStatus: "idle" | "succeeded" | "failed";
    onSubmit: (values: CreateOrEditImPlatformFormOutput) => void;
    onTestSendMsg: (values: CreateOrEditImPlatformFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    initialValues?: Partial<CreateOrEditImPlatformFormInput>;
    showAvailableInProjects: boolean;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}

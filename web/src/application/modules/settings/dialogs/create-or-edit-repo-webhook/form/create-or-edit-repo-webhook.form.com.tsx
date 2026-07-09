import { useEffect } from "react";

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Clipboard } from "lucide-react";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { toast } from "sonner";
import { InheritedSettingReadonlyNotice, PermissionReadonlyNotice } from "~/settings/module-shared/components";
import { SettingsFormCancelAction } from "~/settings/module-shared/components/settings-form-cancel-action";
import {
    SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS,
    SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS,
} from "~/settings/module-shared/constants/settings-form-layout.constants";
import { ERepoWebhookKind } from "~/settings/module-shared/enums";

import { FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";

import type { CreateOrEditRepoWebhookFormInput, CreateOrEditRepoWebhookFormOutput } from "../schemas";
import { CreateOrEditRepoWebhookFormSchema } from "../schemas";

const kindOptions = [
    ERepoWebhookKind.Github,
    ERepoWebhookKind.Gitlab,
    ERepoWebhookKind.Gitea,
    ERepoWebhookKind.Bitbucket,
    ERepoWebhookKind.Gogs,
] satisfies string[];

export function CreateOrEditRepoWebhookForm({
    isPending,
    onSubmit,
    onHasChanges,
    savedVersion = 0,
    initialValues,
    webhookURL,
    showAvailableInProjects,
    readOnlyInherited = false,
    readOnly = false,
    stickyActions = false,
    onClose,
}: Props) {
    const isReadOnly = readOnlyInherited || readOnly;

    const {
        handleSubmit,
        control,
        getValues,
        reset,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditRepoWebhookFormInput, unknown, CreateOrEditRepoWebhookFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            kind: initialValues?.kind ?? ERepoWebhookKind.Github,
            secret: initialValues?.secret ?? "",
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditRepoWebhookFormSchema),
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
    }, [isDirty, isReadOnly, onHasChanges]);

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const {
        field: kind,
        fieldState: { invalid: isKindInvalid },
    } = useController({ name: "kind", control });
    const {
        field: secret,
        fieldState: { invalid: isSecretInvalid },
    } = useController({ name: "secret", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });
    const resolvedKindOptions =
        typeof kind.value === "string" && kind.value.length > 0 && !kindOptions.includes(kind.value)
            ? [...kindOptions, kind.value]
            : kindOptions;

    function onValid(values: CreateOrEditRepoWebhookFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditRepoWebhookFormOutput>) {
        console.error(_errors);
    }

    function handleCopyWebhookURL() {
        if (!webhookURL) {
            return;
        }

        void navigator.clipboard
            .writeText(webhookURL)
            .then(() => {
                toast.success("Webhook URL copied to clipboard");
            })
            .catch(() => {
                toast.error("Failed to copy webhook URL");
            });
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <div className="flex flex-col gap-6">
                {readOnlyInherited && <InheritedSettingReadonlyNotice />}
                {readOnly && !readOnlyInherited && <PermissionReadonlyNotice />}
                <fieldset
                    disabled={isReadOnly}
                    className={cn(
                        "flex flex-col gap-6 border-0 p-0 m-0 min-w-0",
                        SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS,
                    )}
                >
                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Name"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...name}
                                    placeholder="webhook name"
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
                                        {resolvedKindOptions.map(option => (
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
                        title={<LabelWithInfo label="Secret" />}
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...secret}
                                    placeholder="auto-generate if empty"
                                    aria-invalid={isSecretInvalid}
                                />
                                <FieldError errors={[errors.secret]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Webhook URL" />}
                    >
                        {webhookURL ? (
                            <div
                                className={cn(
                                    dashedBorderBox,
                                    SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS,
                                    "flex min-h-12 items-center justify-center gap-3 break-all text-center text-sm",
                                )}
                            >
                                <span>{webhookURL}</span>
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="shrink-0 text-link"
                                    onClick={handleCopyWebhookURL}
                                >
                                    <Clipboard className="size-4" />
                                    <span className="sr-only">Copy webhook URL</span>
                                </Button>
                            </div>
                        ) : (
                            <div
                                className={cn(
                                    dashedBorderBox,
                                    SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS,
                                    "text-center text-sm p-2",
                                )}
                            >
                                please create a webhook first
                            </div>
                        )}
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

                    <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
                        After creating it, you can use the information above to configure the webhook on GitHub, GitLab,
                        or wherever your source code is hosted.
                    </div>
                </fieldset>
            </div>
            {!isReadOnly && (
                <FormActionBar sticky={stickyActions}>
                    <SettingsFormCancelAction
                        onCancel={onClose}
                        disabled={isPending}
                    />
                    <Button
                        type="submit"
                        className="min-w-[100px]"
                        isLoading={isPending}
                    >
                        Save
                    </Button>
                </FormActionBar>
            )}
            {isReadOnly && (
                <FormActionBar sticky={stickyActions}>
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
    onSubmit: (values: CreateOrEditRepoWebhookFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    initialValues?: Partial<CreateOrEditRepoWebhookFormInput>;
    webhookURL?: string;
    showAvailableInProjects: boolean;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    stickyActions?: boolean;
    onClose?: () => void;
}

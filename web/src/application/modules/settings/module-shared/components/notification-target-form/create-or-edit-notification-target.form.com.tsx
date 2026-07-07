import { useEffect, useMemo, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { ProjectEmailQueries, ProjectImServiceQueries } from "~/projects/data/queries";
import { EmailQueries, ImServiceQueries } from "~/settings/data/queries";
import type { SettingEmail, SettingImService } from "~/settings/domain";
import { SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS } from "~/settings/module-shared/constants/settings-form-layout.constants";

import { AppLink, Combobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { EImServiceKind } from "@application/shared/enums";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";

import { InheritedSettingReadonlyNotice } from "../inherited-setting-readonly-notice.com";
import type { NotificationTargetTableScope } from "../notification-target-table";
import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";

import type {
    CreateOrEditNotificationTargetFormInput,
    CreateOrEditNotificationTargetFormOutput,
} from "./create-or-edit-notification-target.form.schema";
import { CreateOrEditNotificationTargetFormSchema } from "./create-or-edit-notification-target.form.schema";

const selectPagination = { page: 1, size: 100 };
type NotificationSection = "email" | "slack" | "discord" | "telegram";

const initialOpenSections: NotificationSection[] = ["email", "slack", "discord", "telegram"];

function mergeOpenSections(current: NotificationSection[], additions: NotificationSection[]) {
    return Array.from(new Set([...current, ...additions]));
}

function getInvalidSections(errors: FieldErrors<CreateOrEditNotificationTargetFormOutput>): NotificationSection[] {
    const sections: NotificationSection[] = [];

    if (errors.senderEmailAccountId || errors.customAddresses) {
        sections.push("email");
    }
    if (errors.slackWebhookId) {
        sections.push("slack");
    }
    if (errors.discordWebhookId) {
        sections.push("discord");
    }
    if (errors.telegramSettingId) {
        sections.push("telegram");
    }

    return sections;
}

function buildEmailOptions(emails: SettingEmail[]) {
    return emails.map(item => ({
        value: { id: item.id, name: item.name },
        label: item.name,
    }));
}

function buildImOptions(items: SettingImService[], kind: EImServiceKind) {
    return items
        .filter(item => item.kind === kind)
        .map(item => ({
            value: { id: item.id, name: item.name },
            label: item.name,
        }));
}

function useNotificationSourceQueries(scope: NotificationTargetTableScope) {
    const settingsEmails = EmailQueries.useFindManyPaginated(
        { pagination: selectPagination },
        { enabled: scope.type === "settings" },
    );
    const settingsImServices = ImServiceQueries.useFindManyPaginated(
        { pagination: selectPagination },
        { enabled: scope.type === "settings" },
    );
    const projectEmails = ProjectEmailQueries.useFindManyPaginated(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            pagination: selectPagination,
        },
        { enabled: scope.type === "project" },
    );
    const projectImServices = ProjectImServiceQueries.useFindManyPaginated(
        {
            projectID: scope.type === "project" ? scope.projectId : "",
            pagination: selectPagination,
        },
        { enabled: scope.type === "project" },
    );

    return {
        emailQuery: scope.type === "project" ? projectEmails : settingsEmails,
        imServiceQuery: scope.type === "project" ? projectImServices : settingsImServices,
    };
}

export function CreateOrEditNotificationTargetForm({
    scope,
    isPending,
    onSubmit,
    onHasChanges,
    savedVersion = 0,
    initialValues,
    showAvailableInProjects,
    readOnlyInherited = false,
    readOnly = false,
    onClose,
}: Props) {
    const isReadOnly = readOnlyInherited || readOnly;
    const [openSections, setOpenSections] = useState<NotificationSection[]>(initialOpenSections);

    const { emailQuery, imServiceQuery } = useNotificationSourceQueries(scope);
    const {
        handleSubmit,
        control,
        getValues,
        reset,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditNotificationTargetFormInput, unknown, CreateOrEditNotificationTargetFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            emailEnabled: initialValues?.emailEnabled ?? false,
            emailUseDefault: initialValues?.emailUseDefault ?? true,
            senderEmailAccountId: initialValues?.senderEmailAccountId ?? "",
            notifyAdmins: initialValues?.notifyAdmins ?? false,
            notifyProjectOwners: initialValues?.notifyProjectOwners ?? false,
            notifyProjectMembers: initialValues?.notifyProjectMembers ?? false,
            customAddresses: initialValues?.customAddresses ?? "",
            slackEnabled: initialValues?.slackEnabled ?? false,
            slackUseDefault: initialValues?.slackUseDefault ?? true,
            slackWebhookId: initialValues?.slackWebhookId ?? "",
            discordEnabled: initialValues?.discordEnabled ?? false,
            discordUseDefault: initialValues?.discordUseDefault ?? true,
            discordWebhookId: initialValues?.discordWebhookId ?? "",
            telegramEnabled: initialValues?.telegramEnabled ?? false,
            telegramUseDefault: initialValues?.telegramUseDefault ?? true,
            telegramSettingId: initialValues?.telegramSettingId ?? "",
            minSendInterval: initialValues?.minSendInterval ?? "3m",
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditNotificationTargetFormSchema),
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
    const { field: emailEnabled } = useController({ name: "emailEnabled", control });
    const { field: emailUseDefault } = useController({ name: "emailUseDefault", control });
    const {
        field: senderEmailAccountId,
        fieldState: { invalid: isSenderEmailAccountInvalid },
    } = useController({ name: "senderEmailAccountId", control });
    const { field: notifyAdmins } = useController({ name: "notifyAdmins", control });
    const { field: notifyProjectOwners } = useController({ name: "notifyProjectOwners", control });
    const { field: notifyProjectMembers } = useController({ name: "notifyProjectMembers", control });
    const {
        field: customAddresses,
        fieldState: { invalid: isCustomAddressesInvalid },
    } = useController({ name: "customAddresses", control });
    const { field: slackEnabled } = useController({ name: "slackEnabled", control });
    const { field: slackUseDefault } = useController({ name: "slackUseDefault", control });
    const {
        field: slackWebhookId,
        fieldState: { invalid: isSlackWebhookInvalid },
    } = useController({ name: "slackWebhookId", control });
    const { field: discordEnabled } = useController({ name: "discordEnabled", control });
    const { field: discordUseDefault } = useController({ name: "discordUseDefault", control });
    const {
        field: discordWebhookId,
        fieldState: { invalid: isDiscordWebhookInvalid },
    } = useController({ name: "discordWebhookId", control });
    const { field: telegramEnabled } = useController({ name: "telegramEnabled", control });
    const { field: telegramUseDefault } = useController({ name: "telegramUseDefault", control });
    const {
        field: telegramSettingId,
        fieldState: { invalid: isTelegramSettingInvalid },
    } = useController({ name: "telegramSettingId", control });
    const {
        field: minSendInterval,
        fieldState: { invalid: isMinSendIntervalInvalid },
    } = useController({ name: "minSendInterval", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    const isEmailEnabled = emailEnabled.value;
    const isEmailUseDefault = emailUseDefault.value;
    const isSlackEnabled = slackEnabled.value;
    const isSlackUseDefault = slackUseDefault.value;
    const isDiscordEnabled = discordEnabled.value;
    const isDiscordUseDefault = discordUseDefault.value;
    const isTelegramEnabled = telegramEnabled.value;
    const isTelegramUseDefault = telegramUseDefault.value;

    const emailOptions = useMemo(() => buildEmailOptions(emailQuery.data?.data ?? []), [emailQuery.data?.data]);
    const slackOptions = useMemo(
        () => buildImOptions(imServiceQuery.data?.data ?? [], EImServiceKind.Slack),
        [imServiceQuery.data?.data],
    );
    const discordOptions = useMemo(
        () => buildImOptions(imServiceQuery.data?.data ?? [], EImServiceKind.Discord),
        [imServiceQuery.data?.data],
    );
    const telegramOptions = useMemo(
        () => buildImOptions(imServiceQuery.data?.data ?? [], EImServiceKind.Telegram),
        [imServiceQuery.data?.data],
    );

    useEffect(() => {
        const option = emailOptions[0];
        if (
            isEmailEnabled &&
            !isEmailUseDefault &&
            !senderEmailAccountId.value &&
            emailOptions.length === 1 &&
            option
        ) {
            senderEmailAccountId.onChange(option.value.id);
        }
    }, [emailOptions, isEmailEnabled, isEmailUseDefault, senderEmailAccountId]);

    useEffect(() => {
        const option = slackOptions[0];
        if (isSlackEnabled && !isSlackUseDefault && !slackWebhookId.value && slackOptions.length === 1 && option) {
            slackWebhookId.onChange(option.value.id);
        }
    }, [isSlackEnabled, isSlackUseDefault, slackOptions, slackWebhookId]);

    useEffect(() => {
        const option = discordOptions[0];
        if (
            isDiscordEnabled &&
            !isDiscordUseDefault &&
            !discordWebhookId.value &&
            discordOptions.length === 1 &&
            option
        ) {
            discordWebhookId.onChange(option.value.id);
        }
    }, [discordOptions, discordWebhookId, isDiscordEnabled, isDiscordUseDefault]);

    useEffect(() => {
        const option = telegramOptions[0];
        if (
            isTelegramEnabled &&
            !isTelegramUseDefault &&
            !telegramSettingId.value &&
            telegramOptions.length === 1 &&
            option
        ) {
            telegramSettingId.onChange(option.value.id);
        }
    }, [isTelegramEnabled, isTelegramUseDefault, telegramOptions, telegramSettingId]);

    const emailAccountsRoute =
        scope.type === "project"
            ? ROUTE.projects.single.providerConfiguration.emailAccounts.$route(scope.projectId)
            : ROUTE.settings.emailAccounts.$route;
    const imPlatformsRoute =
        scope.type === "project"
            ? ROUTE.projects.single.providerConfiguration.imPlatforms.$route(scope.projectId)
            : ROUTE.settings.imPlatforms.$route;

    function onValid(values: CreateOrEditNotificationTargetFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditNotificationTargetFormOutput>) {
        const invalidSections = getInvalidSections(_errors);
        if (invalidSections.length > 0) {
            setOpenSections(current => mergeOpenSections(current, invalidSections));
        }
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

                    <Accordion
                        type="multiple"
                        value={openSections}
                        onValueChange={value => {
                            setOpenSections(value as NotificationSection[]);
                        }}
                        className="w-full flex flex-col gap-4"
                    >
                        <AccordionItem
                            value="email"
                            className="border-none"
                        >
                            <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent">
                                Email Notification
                            </AccordionTrigger>
                            <AccordionContent className="pt-4 pb-0 pl-4">
                                <div className="flex flex-col gap-6">
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Enabled" />}
                                    >
                                        <Checkbox
                                            checked={isEmailEnabled}
                                            onCheckedChange={checked => {
                                                emailEnabled.onChange(Boolean(checked));
                                            }}
                                        />
                                    </InfoBlock>
                                    {isEmailEnabled && (
                                        <>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Use Default Email Account" />}
                                            >
                                                <Checkbox
                                                    checked={isEmailUseDefault}
                                                    onCheckedChange={checked => {
                                                        emailUseDefault.onChange(Boolean(checked));
                                                    }}
                                                />
                                            </InfoBlock>
                                            {!isEmailUseDefault && (
                                                <InfoBlock
                                                    titleWidth={220}
                                                    title={<LabelWithInfo label="Sender Email Account" />}
                                                >
                                                    <FieldGroup>
                                                        <Field>
                                                            <Combobox
                                                                options={emailOptions}
                                                                value={senderEmailAccountId.value}
                                                                onChange={value => {
                                                                    senderEmailAccountId.onChange(value ?? "");
                                                                }}
                                                                placeholder="None"
                                                                emptyText="No email accounts available"
                                                                searchable={false}
                                                                allowClear
                                                                loading={emailQuery.isLoading}
                                                                onRefresh={() => {
                                                                    void emailQuery.refetch();
                                                                }}
                                                                isRefreshing={emailQuery.isRefetching}
                                                                aria-invalid={isSenderEmailAccountInvalid}
                                                            />
                                                            <AppLink.Basic
                                                                className="text-sm text-link"
                                                                to={emailAccountsRoute}
                                                                ignorePrevPath
                                                            >
                                                                Manage email account Settings
                                                            </AppLink.Basic>
                                                            <FieldError errors={[errors.senderEmailAccountId]} />
                                                        </Field>
                                                    </FieldGroup>
                                                </InfoBlock>
                                            )}
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Notify Admins" />}
                                            >
                                                <Checkbox
                                                    checked={notifyAdmins.value}
                                                    onCheckedChange={checked => {
                                                        notifyAdmins.onChange(Boolean(checked));
                                                    }}
                                                />
                                            </InfoBlock>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Notify Project Owners" />}
                                            >
                                                <Checkbox
                                                    checked={notifyProjectOwners.value}
                                                    onCheckedChange={checked => {
                                                        notifyProjectOwners.onChange(Boolean(checked));
                                                    }}
                                                />
                                            </InfoBlock>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Notify Project Members" />}
                                            >
                                                <Checkbox
                                                    checked={notifyProjectMembers.value}
                                                    onCheckedChange={checked => {
                                                        notifyProjectMembers.onChange(Boolean(checked));
                                                    }}
                                                />
                                            </InfoBlock>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Custom Addresses" />}
                                            >
                                                <FieldGroup>
                                                    <Field>
                                                        <Input
                                                            {...customAddresses}
                                                            placeholder="email1, email2"
                                                            aria-invalid={isCustomAddressesInvalid}
                                                        />
                                                        <FieldError errors={[errors.customAddresses]} />
                                                    </Field>
                                                </FieldGroup>
                                            </InfoBlock>
                                        </>
                                    )}
                                </div>
                            </AccordionContent>
                        </AccordionItem>

                        <AccordionItem
                            value="slack"
                            className="border-none"
                        >
                            <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent">
                                Slack Notification
                            </AccordionTrigger>
                            <AccordionContent className="pt-4 pb-0 pl-4">
                                <div className="flex flex-col gap-6">
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Enabled" />}
                                    >
                                        <Checkbox
                                            checked={isSlackEnabled}
                                            onCheckedChange={checked => {
                                                slackEnabled.onChange(Boolean(checked));
                                            }}
                                        />
                                    </InfoBlock>
                                    {isSlackEnabled && (
                                        <>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Use Default Slack Webhook" />}
                                            >
                                                <Checkbox
                                                    checked={isSlackUseDefault}
                                                    onCheckedChange={checked => {
                                                        slackUseDefault.onChange(Boolean(checked));
                                                    }}
                                                />
                                            </InfoBlock>
                                            {!isSlackUseDefault && (
                                                <InfoBlock
                                                    titleWidth={220}
                                                    title={<LabelWithInfo label="Webhook" />}
                                                >
                                                    <FieldGroup>
                                                        <Field>
                                                            <Combobox
                                                                options={slackOptions}
                                                                value={slackWebhookId.value}
                                                                onChange={value => {
                                                                    slackWebhookId.onChange(value ?? "");
                                                                }}
                                                                placeholder="None"
                                                                emptyText="No Slack webhooks available"
                                                                searchable={false}
                                                                allowClear
                                                                loading={imServiceQuery.isLoading}
                                                                onRefresh={() => {
                                                                    void imServiceQuery.refetch();
                                                                }}
                                                                isRefreshing={imServiceQuery.isRefetching}
                                                                aria-invalid={isSlackWebhookInvalid}
                                                            />
                                                            <AppLink.Basic
                                                                className="text-sm text-link"
                                                                to={imPlatformsRoute}
                                                                ignorePrevPath
                                                            >
                                                                Manage IM Platform Settings
                                                            </AppLink.Basic>
                                                            <FieldError errors={[errors.slackWebhookId]} />
                                                        </Field>
                                                    </FieldGroup>
                                                </InfoBlock>
                                            )}
                                        </>
                                    )}
                                </div>
                            </AccordionContent>
                        </AccordionItem>

                        <AccordionItem
                            value="discord"
                            className="border-none"
                        >
                            <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent">
                                Discord Notification
                            </AccordionTrigger>
                            <AccordionContent className="pt-4 pb-0 pl-4">
                                <div className="flex flex-col gap-6">
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Enabled" />}
                                    >
                                        <Checkbox
                                            checked={isDiscordEnabled}
                                            onCheckedChange={checked => {
                                                discordEnabled.onChange(Boolean(checked));
                                            }}
                                        />
                                    </InfoBlock>
                                    {isDiscordEnabled && (
                                        <>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Use Default Discord Webhook" />}
                                            >
                                                <Checkbox
                                                    checked={isDiscordUseDefault}
                                                    onCheckedChange={checked => {
                                                        discordUseDefault.onChange(Boolean(checked));
                                                    }}
                                                />
                                            </InfoBlock>
                                            {!isDiscordUseDefault && (
                                                <InfoBlock
                                                    titleWidth={220}
                                                    title={<LabelWithInfo label="Webhook" />}
                                                >
                                                    <FieldGroup>
                                                        <Field>
                                                            <Combobox
                                                                options={discordOptions}
                                                                value={discordWebhookId.value}
                                                                onChange={value => {
                                                                    discordWebhookId.onChange(value ?? "");
                                                                }}
                                                                placeholder="None"
                                                                emptyText="No Discord webhooks available"
                                                                searchable={false}
                                                                allowClear
                                                                loading={imServiceQuery.isLoading}
                                                                onRefresh={() => {
                                                                    void imServiceQuery.refetch();
                                                                }}
                                                                isRefreshing={imServiceQuery.isRefetching}
                                                                aria-invalid={isDiscordWebhookInvalid}
                                                            />
                                                            <AppLink.Basic
                                                                className="text-sm text-link"
                                                                to={imPlatformsRoute}
                                                                ignorePrevPath
                                                            >
                                                                Manage IM Platform Settings
                                                            </AppLink.Basic>
                                                            <FieldError errors={[errors.discordWebhookId]} />
                                                        </Field>
                                                    </FieldGroup>
                                                </InfoBlock>
                                            )}
                                        </>
                                    )}
                                </div>
                            </AccordionContent>
                        </AccordionItem>

                        <AccordionItem
                            value="telegram"
                            className="border-none"
                        >
                            <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent">
                                Telegram Notification
                            </AccordionTrigger>
                            <AccordionContent className="pt-4 pb-0 pl-4">
                                <div className="flex flex-col gap-6">
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Enabled" />}
                                    >
                                        <Checkbox
                                            checked={isTelegramEnabled}
                                            onCheckedChange={checked => {
                                                telegramEnabled.onChange(Boolean(checked));
                                            }}
                                        />
                                    </InfoBlock>
                                    {isTelegramEnabled && (
                                        <>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Use Default Telegram Bot" />}
                                            >
                                                <Checkbox
                                                    checked={isTelegramUseDefault}
                                                    onCheckedChange={checked => {
                                                        telegramUseDefault.onChange(Boolean(checked));
                                                    }}
                                                />
                                            </InfoBlock>
                                            {!isTelegramUseDefault && (
                                                <InfoBlock
                                                    titleWidth={220}
                                                    title={<LabelWithInfo label="Bot Setting" />}
                                                >
                                                    <FieldGroup>
                                                        <Field>
                                                            <Combobox
                                                                options={telegramOptions}
                                                                value={telegramSettingId.value}
                                                                onChange={value => {
                                                                    telegramSettingId.onChange(value ?? "");
                                                                }}
                                                                placeholder="None"
                                                                emptyText="No Telegram bots available"
                                                                searchable={false}
                                                                allowClear
                                                                loading={imServiceQuery.isLoading}
                                                                onRefresh={() => {
                                                                    void imServiceQuery.refetch();
                                                                }}
                                                                isRefreshing={imServiceQuery.isRefetching}
                                                                aria-invalid={isTelegramSettingInvalid}
                                                            />
                                                            <AppLink.Basic
                                                                className="text-sm text-link"
                                                                to={imPlatformsRoute}
                                                                ignorePrevPath
                                                            >
                                                                Manage IM Platform Settings
                                                            </AppLink.Basic>
                                                            <FieldError errors={[errors.telegramSettingId]} />
                                                        </Field>
                                                    </FieldGroup>
                                                </InfoBlock>
                                            )}
                                        </>
                                    )}
                                </div>
                            </AccordionContent>
                        </AccordionItem>
                    </Accordion>
                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Override Min Send Interval" />}
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...minSendInterval}
                                    placeholder="3m"
                                    aria-invalid={isMinSendIntervalInvalid}
                                />
                                <FieldError errors={[errors.minSendInterval]} />
                            </Field>
                        </FieldGroup>
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
                <div className="shrink-0 px-0 mt-6 pb-6">
                    <div className="flex items-center justify-end gap-3">
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
                <div className="shrink-0 px-0 mt-6 pb-6">
                    <div className="flex justify-end">
                        <Button
                            type="button"
                            onClick={onClose}
                        >
                            Close
                        </Button>
                    </div>
                </div>
            )}
        </form>
    );
}

interface Props {
    scope: NotificationTargetTableScope;
    isPending: boolean;
    onSubmit: (values: CreateOrEditNotificationTargetFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    initialValues?: Partial<CreateOrEditNotificationTargetFormInput>;
    showAvailableInProjects: boolean;
    readOnlyInherited?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}

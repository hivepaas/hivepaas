import { useEffect, useMemo } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { ProjectEmailQueries, ProjectImServiceQueries } from "~/projects/data/queries";
import { EmailQueries, ImServiceQueries } from "~/settings/data/queries";
import type { SettingEmail, SettingImService } from "~/settings/domain";
import type { NotificationTargetTableScope } from "~/settings/module-shared/components";
import { InheritedSettingReadonlyNotice } from "~/settings/module-shared/components/inherited-setting-readonly-notice.com";

import { AppLink, Combobox, ContentBlock, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { EImServiceKind } from "@application/shared/enums";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";

import type { CreateOrEditNotificationTargetFormInput, CreateOrEditNotificationTargetFormOutput } from "../schemas";
import { CreateOrEditNotificationTargetFormSchema } from "../schemas";

const selectPagination = { page: 1, size: 100 };

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
    initialValues,
    showAvailableInProjects,
    readOnlyInherited = false,
    onClose,
}: Props) {
    const { emailQuery, imServiceQuery } = useNotificationSourceQueries(scope);
    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditNotificationTargetFormInput, unknown, CreateOrEditNotificationTargetFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            emailEnabled: initialValues?.emailEnabled ?? false,
            senderEmailAccountId: initialValues?.senderEmailAccountId ?? "",
            notifyAdmins: initialValues?.notifyAdmins ?? false,
            notifyProjectOwners: initialValues?.notifyProjectOwners ?? false,
            notifyProjectMembers: initialValues?.notifyProjectMembers ?? false,
            customAddresses: initialValues?.customAddresses ?? "",
            slackEnabled: initialValues?.slackEnabled ?? false,
            slackWebhookId: initialValues?.slackWebhookId ?? "",
            discordEnabled: initialValues?.discordEnabled ?? false,
            discordWebhookId: initialValues?.discordWebhookId ?? "",
            minSendInterval: initialValues?.minSendInterval ?? "3m",
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditNotificationTargetFormSchema),
        mode: "onSubmit",
    });

    useEffect(() => {
        onHasChanges?.(readOnlyInherited ? false : isDirty);
    }, [isDirty, onHasChanges, readOnlyInherited]);

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const { field: emailEnabled } = useController({ name: "emailEnabled", control });
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
    const {
        field: slackWebhookId,
        fieldState: { invalid: isSlackWebhookInvalid },
    } = useController({ name: "slackWebhookId", control });
    const { field: discordEnabled } = useController({ name: "discordEnabled", control });
    const {
        field: discordWebhookId,
        fieldState: { invalid: isDiscordWebhookInvalid },
    } = useController({ name: "discordWebhookId", control });
    const {
        field: minSendInterval,
        fieldState: { invalid: isMinSendIntervalInvalid },
    } = useController({ name: "minSendInterval", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    const emailOptions = useMemo(() => buildEmailOptions(emailQuery.data?.data ?? []), [emailQuery.data?.data]);
    const slackOptions = useMemo(
        () => buildImOptions(imServiceQuery.data?.data ?? [], EImServiceKind.Slack),
        [imServiceQuery.data?.data],
    );
    const discordOptions = useMemo(
        () => buildImOptions(imServiceQuery.data?.data ?? [], EImServiceKind.Discord),
        [imServiceQuery.data?.data],
    );

    useEffect(() => {
        const option = emailOptions[0];
        if (!senderEmailAccountId.value && emailOptions.length === 1 && option) {
            senderEmailAccountId.onChange(option.value.id);
        }
    }, [emailOptions, senderEmailAccountId]);

    useEffect(() => {
        const option = slackOptions[0];
        if (!slackWebhookId.value && slackOptions.length === 1 && option) {
            slackWebhookId.onChange(option.value.id);
        }
    }, [slackOptions, slackWebhookId]);

    useEffect(() => {
        const option = discordOptions[0];
        if (!discordWebhookId.value && discordOptions.length === 1 && option) {
            discordWebhookId.onChange(option.value.id);
        }
    }, [discordOptions, discordWebhookId]);

    const emailAccountsRoute =
        scope.type === "project"
            ? ROUTE.projects.single.configuration.emailAccounts.$route(scope.projectId)
            : ROUTE.settings.emailAccounts.$route;
    const imPlatformsRoute =
        scope.type === "project"
            ? ROUTE.projects.single.configuration.imPlatforms.$route(scope.projectId)
            : ROUTE.settings.imPlatforms.$route;

    function onValid(values: CreateOrEditNotificationTargetFormOutput) {
        if (readOnlyInherited) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditNotificationTargetFormOutput>) {
        console.error(_errors);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="flex flex-col gap-6"
        >
            {readOnlyInherited && <InheritedSettingReadonlyNotice />}
            <fieldset
                disabled={readOnlyInherited}
                className="flex flex-col gap-6 border-0 p-0 m-0 min-w-0"
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

                <ContentBlock label="Email Notification">
                    <div className="flex flex-col gap-6">
                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label="Enabled" />}
                        >
                            <Checkbox
                                checked={emailEnabled.value}
                                onCheckedChange={checked => {
                                    emailEnabled.onChange(Boolean(checked));
                                }}
                            />
                        </InfoBlock>
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
                                        Configure email accounts
                                    </AppLink.Basic>
                                    <FieldError errors={[errors.senderEmailAccountId]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>
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
                    </div>
                </ContentBlock>

                <ContentBlock label="Slack Notification">
                    <div className="flex flex-col gap-6">
                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label="Enabled" />}
                        >
                            <Checkbox
                                checked={slackEnabled.value}
                                onCheckedChange={checked => {
                                    slackEnabled.onChange(Boolean(checked));
                                }}
                            />
                        </InfoBlock>
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
                                        Configure IM platforms
                                    </AppLink.Basic>
                                    <FieldError errors={[errors.slackWebhookId]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>
                    </div>
                </ContentBlock>

                {/* <SectionTitle>Discord Notification</SectionTitle> */}
                <ContentBlock label="Discord Notification">
                    <div className="flex flex-col gap-6">
                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label="Enabled" />}
                        >
                            <Checkbox
                                checked={discordEnabled.value}
                                onCheckedChange={checked => {
                                    discordEnabled.onChange(Boolean(checked));
                                }}
                            />
                        </InfoBlock>
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
                                        Configure IM platforms
                                    </AppLink.Basic>
                                    <FieldError errors={[errors.discordWebhookId]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>
                    </div>
                </ContentBlock>
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

                {!readOnlyInherited && (
                    <Field>
                        <div className="flex justify-end">
                            <Button
                                type="submit"
                                isLoading={isPending}
                            >
                                Save
                            </Button>
                        </div>
                    </Field>
                )}
            </fieldset>
            {readOnlyInherited && (
                <Field>
                    <div className="flex justify-end">
                        <Button
                            type="button"
                            onClick={onClose}
                        >
                            Close
                        </Button>
                    </div>
                </Field>
            )}
        </form>
    );
}

interface Props {
    scope: NotificationTargetTableScope;
    isPending: boolean;
    onSubmit: (values: CreateOrEditNotificationTargetFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    initialValues?: Partial<CreateOrEditNotificationTargetFormInput>;
    showAvailableInProjects: boolean;
    readOnlyInherited?: boolean;
    onClose?: () => void;
}

import { useEffect } from "react";

import { InputNumber } from "@components/ui/input-number";
import { PasswordInput } from "@components/ui/input-password";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { RefreshCcw } from "lucide-react";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { InheritedSettingReadonlyNotice } from "~/settings/module-shared/components/inherited-setting-readonly-notice.com";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";
import { Textarea } from "@/components/ui/textarea";

import type { CreateOrEditGithubAppFormInput, CreateOrEditGithubAppFormOutput } from "../schemas";
import { CreateOrEditGithubAppFormSchema } from "../schemas";

export function CreateOrEditGithubAppForm({
    isPending,
    isTesting,
    testStatus,
    isReprovisioning,
    onSubmit,
    onTestConnection,
    onReprovision,
    onHasChanges,
    initialValues,
    readonlyValues,
    showAvailableInProjects,
    showTestConnection,
    readOnlyInherited = false,
    onClose,
}: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditGithubAppFormInput, unknown, CreateOrEditGithubAppFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            organization: initialValues?.organization ?? "",
            appId: initialValues?.appId,
            installationId: initialValues?.installationId,
            clientId: initialValues?.clientId ?? "",
            clientSecret: initialValues?.clientSecret ?? "",
            privateKey: initialValues?.privateKey ?? "",
            ssoEnabled: initialValues?.ssoEnabled ?? true,
            availableInProjects: initialValues?.availableInProjects ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditGithubAppFormSchema),
        mode: "onSubmit",
    });

    useEffect(() => {
        onHasChanges?.(readOnlyInherited ? false : isDirty);
    }, [isDirty, onHasChanges, readOnlyInherited]);

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const {
        field: organization,
        fieldState: { invalid: isOrganizationInvalid },
    } = useController({ name: "organization", control });
    const {
        field: appId,
        fieldState: { invalid: isAppIdInvalid },
    } = useController({ name: "appId", control });
    const {
        field: installationId,
        fieldState: { invalid: isInstallationIdInvalid },
    } = useController({ name: "installationId", control });
    const {
        field: clientId,
        fieldState: { invalid: isClientIdInvalid },
    } = useController({ name: "clientId", control });
    const {
        field: clientSecret,
        fieldState: { invalid: isClientSecretInvalid },
    } = useController({ name: "clientSecret", control });
    const {
        field: privateKey,
        fieldState: { invalid: isPrivateKeyInvalid },
    } = useController({ name: "privateKey", control });
    const { field: ssoEnabled } = useController({ name: "ssoEnabled", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    const hasReadonlyWebhookValues = [
        readonlyValues?.callbackURL,
        readonlyValues?.webhookURL,
        readonlyValues?.webhookSecret,
    ].some(value => Boolean(value));

    function onValid(values: CreateOrEditGithubAppFormOutput) {
        if (readOnlyInherited) {
            return;
        }

        onSubmit(values);
    }

    function onTestValid(values: CreateOrEditGithubAppFormOutput) {
        if (readOnlyInherited) {
            return;
        }

        onTestConnection(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditGithubAppFormOutput>) {
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
            {onReprovision && !readOnlyInherited && (
                <div className={cn(dashedBorderBox, "flex flex-col gap-4 text-center text-sm leading-6")}>
                    <div>
                        <span className="text-orange-500">Important:</span> It is recommended that you do not manually
                        modify the information below unless you are certain it is correct. You can use the reprovision
                        feature to reset the GitHub App setup.
                    </div>
                    <div className="flex justify-start">
                        <Button
                            type="button"
                            onClick={onReprovision}
                            isLoading={isReprovisioning}
                        >
                            <RefreshCcw className="size-4" />
                            Reprovision Github App
                        </Button>
                    </div>
                </div>
            )}
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

                <InfoBlock
                    titleWidth={220}
                    title={<LabelWithInfo label="Organization" />}
                >
                    <FieldGroup>
                        <Field>
                            <Input
                                {...organization}
                                aria-invalid={isOrganizationInvalid}
                            />
                            <FieldError errors={[errors.organization]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="App ID"
                            isRequired
                        />
                    }
                >
                    <FieldGroup>
                        <Field>
                            <InputNumber
                                value={appId.value}
                                onValueChange={appId.onChange}
                                min={1}
                                useGrouping={false}
                                showControls={false}
                                aria-invalid={isAppIdInvalid}
                            />
                            <FieldError errors={[errors.appId]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Installation ID"
                            isRequired
                        />
                    }
                >
                    <FieldGroup>
                        <Field>
                            <InputNumber
                                value={installationId.value}
                                onValueChange={installationId.onChange}
                                min={1}
                                useGrouping={false}
                                showControls={false}
                                aria-invalid={isInstallationIdInvalid}
                            />
                            <FieldError errors={[errors.installationId]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Client ID"
                            isRequired
                        />
                    }
                >
                    <FieldGroup>
                        <Field>
                            <Input
                                {...clientId}
                                aria-invalid={isClientIdInvalid}
                            />
                            <FieldError errors={[errors.clientId]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Client Secret"
                            isRequired
                        />
                    }
                >
                    <FieldGroup>
                        <Field>
                            <PasswordInput
                                value={clientSecret.value}
                                onChange={clientSecret.onChange}
                                aria-invalid={isClientSecretInvalid}
                            />
                            <FieldError errors={[errors.clientSecret]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                {hasReadonlyWebhookValues && (
                    <>
                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label="Callback URL" />}
                        >
                            <Input
                                value={readonlyValues?.callbackURL ?? ""}
                                readOnly
                            />
                        </InfoBlock>
                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label="Webhook URL" />}
                        >
                            <Input
                                value={readonlyValues?.webhookURL ?? ""}
                                readOnly
                            />
                        </InfoBlock>
                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label="Webhook Secret" />}
                        >
                            <PasswordInput
                                value={readonlyValues?.webhookSecret ?? ""}
                                readOnly
                            />
                        </InfoBlock>
                    </>
                )}

                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Private Key"
                            isRequired
                        />
                    }
                >
                    <FieldGroup>
                        <Field>
                            <Textarea
                                {...privateKey}
                                minRows={5}
                                maxRows={10}
                                aria-invalid={isPrivateKeyInvalid}
                            />
                            <FieldError errors={[errors.privateKey]} />
                        </Field>
                    </FieldGroup>
                </InfoBlock>

                <InfoBlock
                    titleWidth={220}
                    title={<LabelWithInfo label="SSO Enabled" />}
                >
                    <Checkbox
                        checked={ssoEnabled.value}
                        onCheckedChange={checked => {
                            ssoEnabled.onChange(Boolean(checked));
                        }}
                    />
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
                        <div className="flex flex-wrap items-center justify-between gap-3">
                            <div className="flex items-center gap-3">
                                {showTestConnection && (
                                    <Button
                                        type="button"
                                        variant="secondary"
                                        onClick={() => void handleSubmit(onTestValid, onInvalid)()}
                                        isLoading={isTesting}
                                    >
                                        Test Access
                                    </Button>
                                )}
                                {testStatus === "succeeded" && (
                                    <span className="text-sm text-green-600">Succeeded</span>
                                )}
                                {testStatus === "failed" && <span className="text-sm text-red-600">Failed</span>}
                            </div>
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
    isPending: boolean;
    isTesting: boolean;
    testStatus: "idle" | "succeeded" | "failed";
    isReprovisioning: boolean;
    onSubmit: (values: CreateOrEditGithubAppFormOutput) => void;
    onTestConnection: (values: CreateOrEditGithubAppFormOutput) => void;
    onReprovision?: () => void;
    onHasChanges?: (dirty: boolean) => void;
    initialValues?: Partial<CreateOrEditGithubAppFormInput>;
    readonlyValues?: {
        callbackURL?: string;
        webhookURL?: string;
        webhookSecret?: string;
    };
    showAvailableInProjects: boolean;
    showTestConnection: boolean;
    readOnlyInherited?: boolean;
    onClose?: () => void;
}

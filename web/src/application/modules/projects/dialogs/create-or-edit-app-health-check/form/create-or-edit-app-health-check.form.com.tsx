import React, { useEffect } from "react";

import { Field, FieldError } from "@components/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldErrors, FormProvider, useController, useForm, useWatch } from "react-hook-form";
import type { AppHealthCheck } from "~/projects/domain";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import {
    EAppHealthCheckGrpcStatus,
    EAppHealthCheckGrpcVersion,
    EAppHealthCheckRestMethod,
    EAppHealthCheckReturnBodyMode,
    EAppHealthCheckType,
} from "~/projects/module-shared/enums";
import { useProjectNotificationSettingsSources } from "~/projects/module-shared/hooks";

import { ContentBlock, FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { NotificationSettings } from "@application/shared/form";

import { Button, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";
import { InputNumber } from "@/components/ui/input-number";
import { Textarea } from "@/components/ui/textarea";

import type { CreateOrEditAppHealthCheckFormInput, CreateOrEditAppHealthCheckFormOutput } from "../schemas";
import { CreateOrEditAppHealthCheckFormSchema } from "../schemas";

const BODY_METHODS = [EAppHealthCheckRestMethod.POST, EAppHealthCheckRestMethod.PUT] as const;
// eslint-disable-next-line quotes
const JSON_PLACEHOLDER = '{\n  "status": "up"\n}';

function getReturnBodyMode(rest: AppHealthCheck["rest"]): EAppHealthCheckReturnBodyMode {
    if (rest?.returnText) {
        return EAppHealthCheckReturnBodyMode.Text;
    }

    if (rest?.returnJSON) {
        return EAppHealthCheckReturnBodyMode.JSON;
    }

    return EAppHealthCheckReturnBodyMode.Skipped;
}

function mapInitialValues(healthCheck?: AppHealthCheck): CreateOrEditAppHealthCheckFormInput {
    return {
        name: healthCheck?.name ?? "",
        interval: healthCheck?.interval ?? "30s",
        timeout: healthCheck?.timeout ?? "",
        maxRetry: healthCheck?.maxRetry,
        retryDelay: healthCheck?.retryDelay ?? "",
        healthcheckType: healthCheck?.healthcheckType ?? EAppHealthCheckType.REST,
        rest: {
            url: healthCheck?.rest?.url ?? "",
            method: healthCheck?.rest?.method ?? EAppHealthCheckRestMethod.GET,
            contentType: healthCheck?.rest?.contentType ?? "application/json",
            body: healthCheck?.rest?.body ?? "",
            returnCode: healthCheck?.rest?.returnCode ?? "",
            returnBodyMode: getReturnBodyMode(healthCheck?.rest ?? null),
            textExact: healthCheck?.rest?.returnText?.exact ?? "",
            textRegex: healthCheck?.rest?.returnText?.regex ?? "",
            jsonExact: healthCheck?.rest?.returnJSON?.exact ?? "",
            jsonContain: healthCheck?.rest?.returnJSON?.contain ?? "",
        },
        grpc: {
            version: healthCheck?.grpc?.version ?? EAppHealthCheckGrpcVersion.V1,
            addr: healthCheck?.grpc?.addr ?? "",
            service: healthCheck?.grpc?.service ?? "",
            returnStatus: healthCheck?.grpc?.returnStatus ?? EAppHealthCheckGrpcStatus.Serving,
        },
        notification: {
            successUseDefault: healthCheck?.notification?.successUseDefault ?? true,
            success: healthCheck?.notification?.success,
            failureUseDefault: healthCheck?.notification?.failureUseDefault ?? true,
            failure: healthCheck?.notification?.failure,
            minSendInterval: healthCheck?.notification?.minSendInterval ?? "",
        },
    };
}

export function CreateOrEditAppHealthCheckForm({
    projectId,
    isPending,
    onSubmit,
    onHasChanges,
    initialValues,
    readOnly = false,
    stickyActions = false,
    onClose,
}: Props) {
    const methods = useForm<CreateOrEditAppHealthCheckFormInput, unknown, CreateOrEditAppHealthCheckFormOutput>({
        defaultValues: mapInitialValues(initialValues),
        resolver: zodResolver(CreateOrEditAppHealthCheckFormSchema),
        mode: "onSubmit",
    });

    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
    } = methods;

    const healthcheckType = useWatch({ control, name: "healthcheckType" });
    const restMethod = useWatch({ control, name: "rest.method" });
    const returnBodyMode = useWatch({ control, name: "rest.returnBodyMode" });
    const { sources: notificationSources, manageLink: notificationManageLink } =
        useProjectNotificationSettingsSources(projectId);

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ control, name: "name" });
    const {
        field: interval,
        fieldState: { invalid: isIntervalInvalid },
    } = useController({
        control,
        name: "interval",
    });
    const {
        field: timeout,
        fieldState: { invalid: isTimeoutInvalid },
    } = useController({ control, name: "timeout" });
    const {
        field: maxRetry,
        fieldState: { invalid: isMaxRetryInvalid },
    } = useController({
        control,
        name: "maxRetry",
    });
    const {
        field: retryDelay,
        fieldState: { invalid: isRetryDelayInvalid },
    } = useController({
        control,
        name: "retryDelay",
    });
    const { field: typeField } = useController({ control, name: "healthcheckType" });
    const {
        field: restUrl,
        fieldState: { invalid: isRestUrlInvalid },
    } = useController({
        control,
        name: "rest.url",
    });
    const { field: restMethodField } = useController({ control, name: "rest.method" });
    const { field: restContentType } = useController({ control, name: "rest.contentType" });
    const { field: restBody } = useController({ control, name: "rest.body" });
    const {
        field: restReturnCode,
        fieldState: { invalid: isRestReturnCodeInvalid },
    } = useController({
        control,
        name: "rest.returnCode",
    });
    const { field: returnBodyModeField } = useController({ control, name: "rest.returnBodyMode" });
    const { field: textExact } = useController({ control, name: "rest.textExact" });
    const {
        field: textRegex,
        fieldState: { invalid: isTextRegexInvalid },
    } = useController({
        control,
        name: "rest.textRegex",
    });
    const {
        field: jsonExact,
        fieldState: { invalid: isJsonExactInvalid },
    } = useController({
        control,
        name: "rest.jsonExact",
    });
    const {
        field: jsonContain,
        fieldState: { invalid: isJsonContainInvalid },
    } = useController({
        control,
        name: "rest.jsonContain",
    });
    const { field: grpcVersion } = useController({ control, name: "grpc.version" });
    const {
        field: grpcAddr,
        fieldState: { invalid: isGrpcAddrInvalid },
    } = useController({
        control,
        name: "grpc.addr",
    });
    const { field: grpcService } = useController({ control, name: "grpc.service" });
    const { field: grpcReturnStatus } = useController({ control, name: "grpc.returnStatus" });
    const {
        field: minSendInterval,
        fieldState: { invalid: isMinSendIntervalInvalid },
    } = useController({
        control,
        name: "notification.minSendInterval",
    });

    useEffect(() => {
        onHasChanges?.(readOnly ? false : isDirty);
    }, [isDirty, onHasChanges, readOnly]);

    function onValid(values: CreateOrEditAppHealthCheckFormOutput) {
        if (readOnly) {
            return;
        }

        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditAppHealthCheckFormOutput>) {
        console.error(_errors);
    }

    const showBody = BODY_METHODS.includes(restMethod as (typeof BODY_METHODS)[number]);

    return (
        <FormProvider {...methods}>
            <form
                onSubmit={event => {
                    event.preventDefault();
                    if (readOnly) {
                        return;
                    }

                    void handleSubmit(onValid, onInvalid)(event);
                }}
                className="min-h-0 flex flex-1 flex-col"
            >
                <fieldset
                    disabled={readOnly}
                    className="contents"
                >
                    <div className="flex flex-col gap-6">
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
                                        id="app-health-check-name"
                                        {...name}
                                        placeholder="health check name"
                                        aria-invalid={isNameInvalid}
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                    />
                                    <FieldError errors={[errors.name]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>

                        <ContentBlock label="Scheduling">
                            <div className="flex flex-col gap-6">
                                <InfoBlock
                                    titleWidth={220}
                                    title={
                                        <LabelWithInfo
                                            label="Interval"
                                            isRequired
                                        />
                                    }
                                >
                                    <Field>
                                        <Input
                                            id="app-health-check-interval"
                                            {...interval}
                                            placeholder="30s"
                                            aria-invalid={isIntervalInvalid}
                                            className="max-w-[260px]"
                                        />
                                        <FieldError errors={[errors.interval]} />
                                    </Field>
                                </InfoBlock>

                                <InfoBlock
                                    titleWidth={220}
                                    title={<LabelWithInfo label="Timeout" />}
                                >
                                    <Field>
                                        <Input
                                            id="app-health-check-timeout"
                                            {...timeout}
                                            placeholder="15s"
                                            aria-invalid={isTimeoutInvalid}
                                            className="max-w-[260px]"
                                        />
                                        <FieldError errors={[errors.timeout]} />
                                    </Field>
                                </InfoBlock>

                                <InfoBlock
                                    titleWidth={220}
                                    title={<LabelWithInfo label="Max Retry" />}
                                >
                                    <Field>
                                        <InputNumber
                                            id="app-health-check-max-retry"
                                            value={maxRetry.value}
                                            onValueChange={value => {
                                                const nextValue =
                                                    value !== undefined && Number.isFinite(value) ? value : undefined;

                                                maxRetry.onChange(nextValue);
                                            }}
                                            min={0}
                                            useGrouping={false}
                                            placeholder="1"
                                            aria-invalid={isMaxRetryInvalid}
                                            className="max-w-[100px]"
                                        />
                                        <FieldError errors={[errors.maxRetry]} />
                                    </Field>
                                </InfoBlock>

                                <InfoBlock
                                    titleWidth={220}
                                    title={<LabelWithInfo label="Retry Delay" />}
                                >
                                    <Field>
                                        <Input
                                            id="app-health-check-retry-delay"
                                            {...retryDelay}
                                            placeholder="5s"
                                            aria-invalid={isRetryDelayInvalid}
                                            className="max-w-[260px]"
                                        />
                                        <FieldError errors={[errors.retryDelay]} />
                                    </Field>
                                </InfoBlock>
                            </div>
                        </ContentBlock>

                        <ContentBlock label="Health Check">
                            <div className="flex flex-col gap-6">
                                <InfoBlock
                                    titleWidth={220}
                                    title={
                                        <LabelWithInfo
                                            label="Type"
                                            isRequired
                                        />
                                    }
                                >
                                    <Tabs
                                        value={healthcheckType}
                                        onValueChange={nextValue => {
                                            typeField.onChange(nextValue);
                                        }}
                                    >
                                        <TabsList>
                                            <TabsTrigger value={EAppHealthCheckType.REST}>REST</TabsTrigger>
                                            <TabsTrigger value={EAppHealthCheckType.GRPC}>GRPC</TabsTrigger>
                                        </TabsList>
                                    </Tabs>
                                </InfoBlock>

                                {healthcheckType === EAppHealthCheckType.REST ? (
                                    <>
                                        <InfoBlock
                                            titleWidth={220}
                                            title={
                                                <LabelWithInfo
                                                    label="URL"
                                                    isRequired
                                                />
                                            }
                                        >
                                            <Field>
                                                <Input
                                                    id="app-health-check-rest-url"
                                                    {...restUrl}
                                                    placeholder="https://your-addr/path"
                                                    aria-invalid={isRestUrlInvalid}
                                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                />
                                                <FieldError errors={[errors.rest?.url]} />
                                            </Field>
                                        </InfoBlock>

                                        <InfoBlock
                                            titleWidth={220}
                                            title={
                                                <LabelWithInfo
                                                    label="Method"
                                                    isRequired
                                                />
                                            }
                                        >
                                            <Tabs
                                                value={restMethod}
                                                onValueChange={nextValue => {
                                                    restMethodField.onChange(nextValue);
                                                }}
                                            >
                                                <TabsList>
                                                    <TabsTrigger value={EAppHealthCheckRestMethod.GET}>GET</TabsTrigger>
                                                    <TabsTrigger value={EAppHealthCheckRestMethod.POST}>
                                                        POST
                                                    </TabsTrigger>
                                                    <TabsTrigger value={EAppHealthCheckRestMethod.PUT}>PUT</TabsTrigger>
                                                </TabsList>
                                            </Tabs>
                                        </InfoBlock>

                                        <InfoBlock
                                            titleWidth={220}
                                            title={<LabelWithInfo label="Content Type" />}
                                        >
                                            <Input
                                                id="app-health-check-rest-content-type"
                                                {...restContentType}
                                                placeholder="application/json"
                                                className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                            />
                                        </InfoBlock>

                                        {showBody && (
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Body" />}
                                            >
                                                <Textarea
                                                    id="app-health-check-rest-body"
                                                    {...restBody}
                                                    rows={5}
                                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                />
                                            </InfoBlock>
                                        )}

                                        <InfoBlock
                                            titleWidth={220}
                                            title={<LabelWithInfo label="Return Code Must Be" />}
                                        >
                                            <Field>
                                                <Input
                                                    id="app-health-check-rest-return-code"
                                                    {...restReturnCode}
                                                    placeholder="200,201,202"
                                                    aria-invalid={isRestReturnCodeInvalid}
                                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                />
                                                <FieldError errors={[errors.rest?.returnCode]} />
                                            </Field>
                                        </InfoBlock>

                                        <InfoBlock
                                            titleWidth={220}
                                            title={<LabelWithInfo label="Return Body Must Be" />}
                                        >
                                            <Tabs
                                                value={returnBodyMode}
                                                onValueChange={nextValue => {
                                                    returnBodyModeField.onChange(nextValue);
                                                }}
                                            >
                                                <TabsList>
                                                    <TabsTrigger value={EAppHealthCheckReturnBodyMode.Skipped}>
                                                        Skipped
                                                    </TabsTrigger>
                                                    <TabsTrigger value={EAppHealthCheckReturnBodyMode.Text}>
                                                        Text
                                                    </TabsTrigger>
                                                    <TabsTrigger value={EAppHealthCheckReturnBodyMode.JSON}>
                                                        JSON
                                                    </TabsTrigger>
                                                </TabsList>
                                            </Tabs>
                                        </InfoBlock>

                                        {returnBodyMode === EAppHealthCheckReturnBodyMode.Text && (
                                            <>
                                                <InfoBlock
                                                    titleWidth={220}
                                                    title={<LabelWithInfo label="Text Exact" />}
                                                >
                                                    <Textarea
                                                        id="app-health-check-rest-text-exact"
                                                        {...textExact}
                                                        placeholder="value"
                                                        rows={4}
                                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                    />
                                                </InfoBlock>

                                                <InfoBlock
                                                    titleWidth={220}
                                                    title={<LabelWithInfo label="Text Regex" />}
                                                >
                                                    <Field>
                                                        <Input
                                                            id="app-health-check-rest-text-regex"
                                                            {...textRegex}
                                                            placeholder="regular expression"
                                                            aria-invalid={isTextRegexInvalid}
                                                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                        />
                                                        <FieldError errors={[errors.rest?.textRegex]} />
                                                    </Field>
                                                </InfoBlock>
                                            </>
                                        )}

                                        {returnBodyMode === EAppHealthCheckReturnBodyMode.JSON && (
                                            <>
                                                <InfoBlock
                                                    titleWidth={220}
                                                    title={<LabelWithInfo label="JSON Exact" />}
                                                >
                                                    <Field>
                                                        <Textarea
                                                            id="app-health-check-rest-json-exact"
                                                            {...jsonExact}
                                                            placeholder={JSON_PLACEHOLDER}
                                                            rows={5}
                                                            aria-invalid={isJsonExactInvalid}
                                                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                        />
                                                        <FieldError errors={[errors.rest?.jsonExact]} />
                                                    </Field>
                                                </InfoBlock>

                                                <InfoBlock
                                                    titleWidth={220}
                                                    title={<LabelWithInfo label="JSON Contains" />}
                                                >
                                                    <Field>
                                                        <Textarea
                                                            id="app-health-check-rest-json-contain"
                                                            {...jsonContain}
                                                            placeholder={JSON_PLACEHOLDER}
                                                            rows={4}
                                                            aria-invalid={isJsonContainInvalid}
                                                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                        />
                                                        <FieldError errors={[errors.rest?.jsonContain]} />
                                                    </Field>
                                                </InfoBlock>
                                            </>
                                        )}
                                    </>
                                ) : (
                                    <>
                                        <InfoBlock
                                            titleWidth={220}
                                            title={<LabelWithInfo label="Healthcheck Version" />}
                                        >
                                            <Tabs
                                                value={grpcVersion.value}
                                                onValueChange={nextValue => {
                                                    grpcVersion.onChange(nextValue);
                                                }}
                                            >
                                                <TabsList>
                                                    <TabsTrigger value={EAppHealthCheckGrpcVersion.V1}>v1</TabsTrigger>
                                                </TabsList>
                                            </Tabs>
                                        </InfoBlock>

                                        <InfoBlock
                                            titleWidth={220}
                                            title={
                                                <LabelWithInfo
                                                    label="Address"
                                                    isRequired
                                                />
                                            }
                                        >
                                            <Field>
                                                <Input
                                                    id="app-health-check-grpc-address"
                                                    {...grpcAddr}
                                                    placeholder="grpc address"
                                                    aria-invalid={isGrpcAddrInvalid}
                                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                />
                                                <FieldError errors={[errors.grpc?.addr]} />
                                            </Field>
                                        </InfoBlock>

                                        <InfoBlock
                                            titleWidth={220}
                                            title={<LabelWithInfo label="Service" />}
                                        >
                                            <Input
                                                id="app-health-check-grpc-service"
                                                {...grpcService}
                                                placeholder="grpc service"
                                                className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                            />
                                        </InfoBlock>

                                        <InfoBlock
                                            titleWidth={220}
                                            title={
                                                <LabelWithInfo
                                                    label="Return Status Must Be"
                                                    isRequired
                                                />
                                            }
                                        >
                                            <InputNumber
                                                id="app-health-check-grpc-return-status"
                                                value={grpcReturnStatus.value}
                                                onValueChange={value => {
                                                    grpcReturnStatus.onChange(
                                                        value ?? EAppHealthCheckGrpcStatus.Serving,
                                                    );
                                                }}
                                                min={1}
                                                useGrouping={false}
                                                className="max-w-[260px]"
                                            />
                                        </InfoBlock>
                                    </>
                                )}
                            </div>
                        </ContentBlock>

                        <ContentBlock label="Notification Configuration">
                            <NotificationSettings<CreateOrEditAppHealthCheckFormInput>
                                names={{
                                    successUseDefault: "notification.successUseDefault",
                                    success: "notification.success",
                                    failureUseDefault: "notification.failureUseDefault",
                                    failure: "notification.failure",
                                }}
                                sources={notificationSources}
                                manageLink={notificationManageLink}
                                readOnly={readOnly}
                                titleWidth={220}
                            >
                                <div className={cn(dashedBorderBox, "text-center")}>
                                    <span className="text-orange-500">Note:</span> If you don&apos;t want to receive
                                    continuous notifications for identical results, use this configuration. For example,
                                    if you set <span className="text-orange-500"> Min Send Interval = 10m</span>, and
                                    the health check result remains success or failure within 10 minutes, you will only
                                    receive one notification.
                                </div>

                                <InfoBlock
                                    titleWidth={220}
                                    title={<LabelWithInfo label="Min Send Interval" />}
                                >
                                    <Field>
                                        <Input
                                            id="app-health-check-min-send-interval"
                                            {...minSendInterval}
                                            placeholder="10m"
                                            aria-invalid={isMinSendIntervalInvalid}
                                            className="max-w-[260px]"
                                        />
                                        <FieldError errors={[errors.notification?.minSendInterval]} />
                                    </Field>
                                </InfoBlock>
                            </NotificationSettings>
                        </ContentBlock>
                    </div>
                    {!readOnly && (
                        <FormActionBar sticky={stickyActions}>
                            <Button
                                type="button"
                                variant="outline"
                                className="min-w-[100px]"
                                disabled={isPending}
                                onClick={onClose}
                            >
                                Cancel
                            </Button>
                            <Button
                                type="submit"
                                isLoading={isPending}
                                className="min-w-[100px]"
                            >
                                Save
                            </Button>
                        </FormActionBar>
                    )}
                    {readOnly && (
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
                </fieldset>
            </form>
        </FormProvider>
    );
}

interface Props {
    projectId: string;
    isPending: boolean;
    onSubmit: (values: CreateOrEditAppHealthCheckFormOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
    initialValues?: AppHealthCheck;
    readOnly?: boolean;
    stickyActions?: boolean;
    onClose?: () => void;
}

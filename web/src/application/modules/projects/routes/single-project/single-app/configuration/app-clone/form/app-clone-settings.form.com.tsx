import React, { type PropsWithChildren, type ReactNode, useImperativeHandle, useMemo, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { ArrowRight, Plus, Trash2 } from "lucide-react";
import {
    type FieldPath,
    FormProvider,
    type FieldError as HookFormFieldError,
    useController,
    useFieldArray,
    useForm,
    useFormContext,
    useWatch,
} from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { toast } from "sonner";
import { ProjectCommandPipeQueries, ProjectSslCertQueries } from "~/projects/data";
import type { AppCloneSettings } from "~/projects/domain";
import type { ProjectEnvEntity } from "~/projects/domain";
import { ProjectAppStatusBadge, ProjectEnvBadge } from "~/projects/module-shared/components";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

import {
    AppLink,
    Combobox,
    ComboboxWithAddon,
    ContentBlock,
    InfoBlock,
    LabelWithInfo,
    PopConfirm,
} from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";

import type { ValidationException } from "@infrastructure/exceptions/validation";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Field, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger } from "@/components/ui/select";

import { AppCloneNotificationFields } from "../building-blocks";
import {
    type AppCloneSettingsFormSchemaInput,
    type AppCloneSettingsFormSchemaOutput,
    createAppCloneSettingsFormSchema,
    emptyAppCloneSettingsFormDefaults,
} from "../schemas";
import type { AppCloneSettingsFormRef } from "../types";

import { mapAppCloneSettingsToFormInput } from "./app-clone-settings.form-mappers";

type SchemaInput = AppCloneSettingsFormSchemaInput;
type SchemaOutput = AppCloneSettingsFormSchemaOutput;

const FALLBACK_ENV_COLOR = "#64748b";

function ReadOnlyValue({ children }: { children: ReactNode }) {
    return (
        <div className="flex min-h-9 min-w-0 items-center rounded-md border border-transparent px-0 py-2 text-sm">
            <div className="min-w-0 truncate">{children}</div>
        </div>
    );
}

function MappingRow({
    label,
    source,
    error,
    children,
}: {
    label: string;
    source: ReactNode;
    error?: HookFormFieldError;
    children: ReactNode;
}) {
    return (
        <InfoBlock
            titleWidth={220}
            title={<LabelWithInfo label={label} />}
        >
            <div className="grid min-w-0 grid-cols-[minmax(0,1fr)_1.5rem_minmax(0,1.25fr)] items-start gap-3">
                <ReadOnlyValue>{source}</ReadOnlyValue>
                <div className="flex h-9 items-center justify-center text-muted-foreground">
                    <ArrowRight className="size-4" />
                </div>
                <FieldGroup>
                    <Field>
                        {children}
                        {error ? <FieldError errors={[error]} /> : null}
                    </Field>
                </FieldGroup>
            </div>
        </InfoBlock>
    );
}

function SectionEnabledField({ name }: { name: FieldPath<SchemaInput> }) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field } = useController({ control, name });

    return (
        <InfoBlock
            titleWidth={220}
            title="Enabled"
        >
            <Checkbox
                checked={field.value === true}
                onCheckedChange={value => {
                    field.onChange(value === true);
                }}
            />
        </InfoBlock>
    );
}

function BooleanFlagField({ name, label }: { name: FieldPath<SchemaInput>; label: string }) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field } = useController({ control, name });

    return (
        <InfoBlock
            titleWidth={220}
            title={label}
        >
            <Checkbox
                checked={field.value === true}
                onCheckedChange={value => {
                    field.onChange(value === true);
                }}
            />
        </InfoBlock>
    );
}

function TargetEnvSelect({
    envs,
    readOnly,
    error,
}: {
    envs: ProjectEnvEntity[];
    readOnly: boolean;
    error?: HookFormFieldError;
}) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field } = useController({ control, name: "targetEnv" });
    const targetEnv = field.value.trim();
    const envOptions = useMemo(() => {
        const nonEmptyEnvs = envs.filter(env => env.name.trim() !== "");

        if (!targetEnv || nonEmptyEnvs.some(env => env.name === targetEnv)) {
            return nonEmptyEnvs;
        }

        return [{ name: targetEnv, color: FALLBACK_ENV_COLOR }, ...nonEmptyEnvs];
    }, [envs, targetEnv]);
    const selectedEnv = targetEnv ? envOptions.find(env => env.name === targetEnv) : null;

    return (
        <InfoBlock
            titleWidth={220}
            title="Target Environment"
        >
            <Select
                value={targetEnv}
                disabled={readOnly || envOptions.length === 0}
                onValueChange={field.onChange}
            >
                <SelectTrigger
                    aria-invalid={Boolean(error)}
                    className={cn(PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS, "px-2")}
                >
                    {selectedEnv ? (
                        <ProjectEnvBadge
                            name={selectedEnv.name}
                            color={selectedEnv.color}
                        />
                    ) : (
                        <span className="text-muted-foreground">
                            {envOptions.length === 0 ? "No environments" : "Select environment"}
                        </span>
                    )}
                </SelectTrigger>
                <SelectContent>
                    {envOptions.map(env => (
                        <SelectItem
                            key={env.name}
                            value={env.name}
                        >
                            <ProjectEnvBadge
                                name={env.name}
                                color={env.color}
                            />
                        </SelectItem>
                    ))}
                </SelectContent>
            </Select>
            {error ? <FieldError errors={[error]} /> : null}
        </InfoBlock>
    );
}

function TargetStatusSelect({ readOnly, error }: { readOnly: boolean; error?: HookFormFieldError }) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field } = useController({ control, name: "targetStatus" });

    return (
        <InfoBlock
            titleWidth={220}
            title="Target Status"
        >
            <Select
                value={field.value}
                disabled={readOnly}
                onValueChange={field.onChange}
            >
                <SelectTrigger
                    aria-invalid={Boolean(error)}
                    className={cn(PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS, "px-2")}
                >
                    <ProjectAppStatusBadge status={field.value} />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value={EProjectAppStatus.Active}>
                        <ProjectAppStatusBadge status={EProjectAppStatus.Active} />
                    </SelectItem>
                    <SelectItem value={EProjectAppStatus.Disabled}>
                        <ProjectAppStatusBadge status={EProjectAppStatus.Disabled} />
                    </SelectItem>
                </SelectContent>
            </Select>
            {error ? <FieldError errors={[error]} /> : null}
        </InfoBlock>
    );
}

function RoutingDomainFields({ projectId, env, readOnly }: { projectId: string; env?: string; readOnly: boolean }) {
    const { control, formState } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { fields } = useFieldArray({ control, name: "cloneRoutingDomains" });

    return (
        <div className="flex flex-col gap-4">
            {fields.map((field, index) => (
                <RoutingDomainRow
                    key={field.id}
                    projectId={projectId}
                    env={env}
                    index={index}
                    readOnly={readOnly}
                    domainError={formState.errors.cloneRoutingDomains?.[index]?.targetDomain}
                />
            ))}
        </div>
    );
}

function RoutingDomainRow({
    projectId,
    env,
    index,
    readOnly,
    domainError,
}: {
    projectId: string;
    env?: string;
    index: number;
    readOnly: boolean;
    domainError?: HookFormFieldError;
}) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const sourceDomain = useWatch({ control, name: `cloneRoutingDomains.${index}.sourceDomain` });
    const sourceSslCert = useWatch({ control, name: `cloneRoutingDomains.${index}.sourceSslCert` });
    const targetDomain = useWatch({ control, name: `cloneRoutingDomains.${index}.targetDomain` });
    const { field: targetDomainField } = useController({ control, name: `cloneRoutingDomains.${index}.targetDomain` });
    const { field: targetSslCertField } = useController({
        control,
        name: `cloneRoutingDomains.${index}.targetSslCert`,
    });

    const [searchQuery, setSearchQuery] = useState("");
    const normalizedTargetDomain = typeof targetDomain === "string" ? targetDomain.trim() : "";
    const hasTargetDomain = normalizedTargetDomain !== "";

    const {
        data: { data: sslCerts } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = ProjectSslCertQueries.useFindManyPaginated(
        {
            projectID: projectId,
            env,
            search: searchQuery,
            domain: normalizedTargetDomain,
        },
        {
            enabled: Boolean(projectId) && hasTargetDomain,
        },
    );

    const comboboxOptions = useMemo(() => {
        if (!hasTargetDomain) {
            return [];
        }

        const options = sslCerts.map(cert => ({
            value: { id: cert.id, name: cert.name, domain: cert.domain },
            label: cert.name,
        }));

        if (targetSslCertField.value?.id && !options.some(option => option.value.id === targetSslCertField.value?.id)) {
            return [
                {
                    value: {
                        id: targetSslCertField.value.id,
                        name: targetSslCertField.value.name,
                        domain: "",
                    },
                    label: targetSslCertField.value.name,
                },
                ...options,
            ];
        }

        return options;
    }, [hasTargetDomain, sslCerts, targetSslCertField.value]);

    return (
        <div className="flex flex-col gap-4">
            <MappingRow
                label="Domain"
                source={sourceDomain}
                error={domainError}
            >
                <Input
                    {...targetDomainField}
                    disabled={readOnly}
                    aria-invalid={Boolean(domainError)}
                />
            </MappingRow>
            <MappingRow
                label="SSL Cert"
                source={sourceSslCert?.name ?? <span className="text-muted-foreground">-</span>}
            >
                <Combobox
                    options={comboboxOptions}
                    value={hasTargetDomain ? (targetSslCertField.value?.id ?? null) : null}
                    onChange={(_, option) => {
                        if (readOnly || !hasTargetDomain) {
                            return;
                        }

                        targetSslCertField.onChange(option ? { id: option.id, name: option.name } : null);
                    }}
                    onSearch={setSearchQuery}
                    placeholder="Select SSL certificate"
                    searchable
                    closeOnSelect
                    emptyText="No SSL certificates available"
                    valueKey="id"
                    loading={isFetching}
                    onRefresh={hasTargetDomain ? () => void refetch() : undefined}
                    isRefreshing={isRefetching}
                    disabled={readOnly || !hasTargetDomain}
                />
            </MappingRow>
        </div>
    );
}

function CommandPipesSection({ projectId, env, readOnly }: { projectId: string; env?: string; readOnly: boolean }) {
    const { control, setValue } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const commandPipes = useWatch({ control, name: "commandPipes" });
    const { field: enabledField } = useController({ control, name: "postCloneCommandsEnabled" });
    const [searchQuery, setSearchQuery] = useState("");
    const [selectedCommand, setSelectedCommand] = useState<{ id: string; name: string } | null>(null);

    const {
        data: { data: commandPipesList } = DEFAULT_PAGINATED_DATA,
        isLoading,
        refetch,
        isRefetching,
    } = ProjectCommandPipeQueries.useFindManyPaginated({
        projectID: projectId,
        env,
        search: searchQuery,
    });

    const availableOptions = useMemo(() => {
        const selectedIds = new Set(commandPipes.map(pipe => pipe.id));

        return commandPipesList
            .filter(pipe => !selectedIds.has(pipe.id))
            .map(pipe => ({
                value: { id: pipe.id, name: pipe.name },
                label: pipe.name,
            }));
    }, [commandPipes, commandPipesList]);

    function handleEnabledChange(checked: boolean) {
        enabledField.onChange(checked);

        if (!checked) {
            setValue("commandPipes", [], { shouldDirty: true });
            setSelectedCommand(null);
            setSearchQuery("");
        }
    }

    function handleAddCommand() {
        if (!selectedCommand || readOnly) {
            return;
        }

        if (commandPipes.some(pipe => pipe.id === selectedCommand.id)) {
            toast.error(`"${selectedCommand.name}" already exists`);
            return;
        }

        setValue("commandPipes", [...commandPipes, selectedCommand], { shouldDirty: true });
        setSelectedCommand(null);
        setSearchQuery("");
    }

    function handleRemoveCommand(id: string) {
        if (readOnly) {
            return;
        }

        setValue(
            "commandPipes",
            commandPipes.filter(pipe => pipe.id !== id),
            { shouldDirty: true },
        );
    }

    return (
        <>
            <InfoBlock
                titleWidth={220}
                title="Enabled"
            >
                <Checkbox
                    checked={enabledField.value}
                    disabled={readOnly}
                    onCheckedChange={value => {
                        handleEnabledChange(value === true);
                    }}
                />
            </InfoBlock>

            {enabledField.value ? (
                <>
                    <div className={cn(dashedBorderBox, "leading-6 mb-6")}>
                        <p>
                            <span className="font-semibold text-orange-500">Note:</span>{" "}
                            <span className="font-semibold">Post-Clone Commands</span> offer higher data consistency
                            compared to raw Volume Cloning. For instance, you can use{" "}
                            <code className="text-orange-500">pg_dump</code> on the source app paired with{" "}
                            <code className="text-orange-500">pg_restore</code> on the target app to clone a PostgreSQL
                            database without stopping the source app.
                        </p>
                        <p className="mt-2 italic">
                            💡 If you choose this approach, you should uncheck{" "}
                            <span className="text-orange-500">Clone Volume Data</span> while enabling{" "}
                            <span className="text-orange-500">Clone Volumes</span> (to ensure empty volumes are created
                            for the new app).
                        </p>
                    </div>

                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Command Pipes"
                                content="Command pipes to execute on the target app after cloning."
                            />
                        }
                    >
                        <div className="flex max-w-[545px] flex-col gap-2">
                            <div className="flex gap-2">
                                <ComboboxWithAddon<{ id: string; name: string }>
                                    addonLeft="Pipe"
                                    value={selectedCommand?.id}
                                    onChange={(_, option) => {
                                        setSelectedCommand(option);
                                    }}
                                    onSearch={setSearchQuery}
                                    onRefresh={() => void refetch()}
                                    isRefreshing={isRefetching}
                                    loading={isLoading}
                                    valueKey="id"
                                    options={availableOptions}
                                    placeholder="select command pipe to add"
                                    classNameContainer="max-w-[460px]"
                                    disabled={readOnly}
                                />

                                <Button
                                    type="button"
                                    variant="outline"
                                    onClick={handleAddCommand}
                                    disabled={readOnly || !selectedCommand}
                                >
                                    <Plus className="size-4" /> Add
                                </Button>
                            </div>

                            <div className="text-xs">
                                <AppLink.Basic
                                    to={ROUTE.projects.single.providerConfiguration.commandPipes.$route(projectId)}
                                    className="text-link"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    Configure Command Pipes
                                </AppLink.Basic>
                            </div>

                            {commandPipes.length > 0 && (
                                <div className="flex w-full flex-col divide-y">
                                    {commandPipes.map(pipe => (
                                        <div
                                            key={pipe.id}
                                            className="grid grid-cols-[minmax(0,1fr)_auto_76px] items-center gap-2 py-1.5"
                                        >
                                            <div className="grid min-w-0 grid-cols-1 items-center">
                                                <span className="break-words text-sm">{pipe.name}</span>
                                            </div>
                                            <AppLink.Basic
                                                to={ROUTE.projects.single.providerConfiguration.commandPipes.edit.$route(
                                                    projectId,
                                                    pipe.id,
                                                )}
                                                className="shrink-0 text-xs text-link"
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                ignorePrevPath
                                            >
                                                View
                                            </AppLink.Basic>
                                            <div className="flex w-[76px] justify-end">
                                                <PopConfirm
                                                    title="Remove command pipe"
                                                    variant="destructive"
                                                    confirmText="Remove"
                                                    cancelText="Cancel"
                                                    description="Are you sure you want to remove this command pipe?"
                                                    onConfirm={() => {
                                                        if (!readOnly) {
                                                            handleRemoveCommand(pipe.id);
                                                        }
                                                    }}
                                                >
                                                    <Button
                                                        type="button"
                                                        variant="ghost"
                                                        size="icon"
                                                        className="h-8 w-8 rounded-md text-zinc-400 hover:bg-red-50 hover:text-red-500"
                                                        disabled={readOnly}
                                                    >
                                                        <Trash2 className="size-3.5" />
                                                    </Button>
                                                </PopConfirm>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>
                    </InfoBlock>
                </>
            ) : null}
        </>
    );
}

function ConditionalSection({ enabled, children }: PropsWithChildren<{ enabled: boolean }>) {
    if (!enabled) {
        return null;
    }

    return children;
}

export function AppCloneSettingsForm({
    ref,
    projectId,
    env,
    envs,
    defaultValues,
    onSubmit,
    readOnly = false,
    children,
}: Props) {
    const envNames = useMemo(() => envs.map(item => item.name).filter(name => name.trim() !== ""), [envs]);
    const schema = useMemo(() => createAppCloneSettingsFormSchema(envNames), [envNames]);

    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapAppCloneSettingsToFormInput(defaultValues)
            : emptyAppCloneSettingsFormDefaults,
        resolver: zodResolver(schema),
        mode: "onSubmit",
    });

    const cloneDeploymentSettings = useWatch({ control: methods.control, name: "cloneDeploymentSettings" });
    const cloneRoutingSettings = useWatch({ control: methods.control, name: "cloneRoutingSettings" });
    const cloneVolumes = useWatch({ control: methods.control, name: "cloneVolumes" });

    useUpdateEffect(() => {
        methods.reset(
            defaultValues ? mapAppCloneSettingsToFormInput(defaultValues) : emptyAppCloneSettingsFormDefaults,
        );
    }, [defaultValues]);

    useImperativeHandle(
        ref,
        () => ({
            submit: () => {
                void methods.handleSubmit(onSubmit)();
            },
            setValues: (values: Partial<SchemaInput>) => {
                methods.reset({
                    ...methods.getValues(),
                    ...values,
                } as SchemaInput);
            },
            onError(error: ValidationException) {
                if (error.errors.length === 0) {
                    return;
                }

                error.errors.forEach(({ path, message }, index) => {
                    methods.setError(
                        path as FieldPath<SchemaInput>,
                        { message, type: "manual" },
                        {
                            shouldFocus: index === 0,
                        },
                    );
                });
            },
        }),
        [methods, onSubmit],
    );

    return (
        <div className="pt-2">
            <FormProvider {...methods}>
                <form
                    onSubmit={event => {
                        event.preventDefault();
                        if (readOnly) {
                            return;
                        }

                        void methods.handleSubmit(onSubmit)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    <fieldset
                        disabled={readOnly}
                        className="contents"
                    >
                        <ContentBlock label="General">
                            <div className="flex flex-col gap-6">
                                <InfoBlock
                                    titleWidth={220}
                                    title="Target Name"
                                >
                                    <Input
                                        {...methods.register("targetName")}
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                        aria-invalid={Boolean(methods.formState.errors.targetName)}
                                    />
                                    {methods.formState.errors.targetName ? (
                                        <FieldError errors={[methods.formState.errors.targetName]} />
                                    ) : null}
                                </InfoBlock>

                                <TargetEnvSelect
                                    envs={envs}
                                    readOnly={readOnly}
                                    error={methods.formState.errors.targetEnv}
                                />

                                <TargetStatusSelect
                                    readOnly={readOnly}
                                    error={methods.formState.errors.targetStatus}
                                />

                                <BooleanFlagField
                                    name="cloneEnvVars"
                                    label="Clone Env Variables"
                                />
                                <BooleanFlagField
                                    name="cloneSecrets"
                                    label="Clone Secrets"
                                />
                                <BooleanFlagField
                                    name="cloneConfigFiles"
                                    label="Clone Config Files"
                                />
                                <BooleanFlagField
                                    name="clonePeriodicJobs"
                                    label="Clone Periodic Jobs"
                                />
                                <BooleanFlagField
                                    name="cloneSchedJobs"
                                    label="Clone Scheduled Jobs"
                                />
                            </div>
                        </ContentBlock>

                        <ContentBlock label="Clone Deployment Settings">
                            <div className="flex flex-col gap-6">
                                <SectionEnabledField name="cloneDeploymentSettings" />
                                <ConditionalSection enabled={cloneDeploymentSettings}>
                                    <InfoBlock
                                        titleWidth={220}
                                        title="Target Replicas"
                                    >
                                        <div className="flex flex-wrap items-center gap-3">
                                            <Input
                                                type="number"
                                                {...methods.register("targetReplicas")}
                                                className="w-24"
                                                aria-invalid={Boolean(methods.formState.errors.targetReplicas)}
                                            />
                                            <span className="text-sm text-muted-foreground">
                                                (-1: inherit from source app)
                                            </span>
                                        </div>
                                        {methods.formState.errors.targetReplicas ? (
                                            <FieldError errors={[methods.formState.errors.targetReplicas]} />
                                        ) : null}
                                    </InfoBlock>
                                </ConditionalSection>
                            </div>
                        </ContentBlock>

                        <ContentBlock label="Clone Routing Settings">
                            <div className="flex flex-col gap-6">
                                <SectionEnabledField name="cloneRoutingSettings" />
                                <ConditionalSection enabled={cloneRoutingSettings}>
                                    <RoutingDomainFields
                                        projectId={projectId}
                                        env={env}
                                        readOnly={readOnly}
                                    />
                                </ConditionalSection>
                            </div>
                        </ContentBlock>

                        <ContentBlock label="Clone Volumes">
                            <div className="flex flex-col gap-6">
                                <SectionEnabledField name="cloneVolumes" />
                                <ConditionalSection enabled={cloneVolumes}>
                                    <div className={cn(dashedBorderBox, "text-sm leading-6")}>
                                        <p>
                                            <span className="font-semibold text-orange-500">Warning:</span>{" "}
                                            <span className="font-semibold">Raw volume cloning</span> is not recommended
                                            for database-type applications (e.g., PostgreSQL, MySQL, Redis, MongoDB)
                                            while actively running with continuous read/write operations and active
                                            memory buffering. Direct file-level copying during active I/O can easily
                                            result in database corruption or inconsistent data states.
                                        </p>
                                        <p className="mt-2 italic">
                                            💡 Recommendation: Use dedicated database dump/restore utilities (e.g.,
                                            pg_dump, mysqldump) via Post-Clone Commands, or stop the source application
                                            before cloning volumes.
                                        </p>
                                    </div>
                                    <BooleanFlagField
                                        name="cloneVolumeData"
                                        label="Clone Volume Data"
                                    />
                                    <BooleanFlagField
                                        name="stopSourceAppBeforeClone"
                                        label="Stop Source App Before Clone"
                                    />
                                </ConditionalSection>
                            </div>
                        </ContentBlock>

                        <ContentBlock label="Post-Clone Commands">
                            <CommandPipesSection
                                projectId={projectId}
                                env={env}
                                readOnly={readOnly}
                            />
                        </ContentBlock>

                        <ContentBlock label="Notification Configuration">
                            <AppCloneNotificationFields readOnly={readOnly} />
                        </ContentBlock>

                        {children}
                    </fieldset>
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppCloneSettingsFormRef>;
    projectId: string;
    env?: string;
    envs: ProjectEnvEntity[];
    defaultValues?: AppCloneSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;

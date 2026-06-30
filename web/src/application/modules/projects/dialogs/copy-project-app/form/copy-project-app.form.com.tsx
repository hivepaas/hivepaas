import React, { type ReactNode, useEffect, useMemo, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { ArrowRight } from "lucide-react";
import {
    type FieldErrors,
    FormProvider,
    type FieldError as HookFormFieldError,
    useController,
    useFieldArray,
    useForm,
    useFormContext,
    useWatch,
} from "react-hook-form";
import type { ProjectEnvEntity } from "~/projects/domain";
import { ProjectAppStatusBadge, ProjectEnvBadge } from "~/projects/module-shared/components";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

import { Combobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";

import { ProjectSslCertQueries } from "@application/modules/projects/data";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DialogActionFooter, DialogBody } from "@/components/ui/dialog";
import { Field, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger } from "@/components/ui/select";

import {
    type CopyProjectAppFormInput,
    type CopyProjectAppFormOutput,
    createCopyProjectAppFormSchema,
} from "../schemas";

const FALLBACK_ENV_COLOR = "#64748b";

function ReadOnlyValue({ children }: { children: ReactNode }) {
    return (
        <div className="flex min-h-9 min-w-0 items-center rounded-md border border-transparent px-3 py-2 text-sm">
            <div className="min-w-0 truncate">{children}</div>
        </div>
    );
}

function CopyFormRow({ label, source, error, children }: CopyFormRowProps) {
    return (
        <InfoBlock
            titleWidth={150}
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

function CopyFlagRow({ label, name, readOnly }: CopyFlagRowProps) {
    const { control } = useFormContext<CopyProjectAppFormInput, unknown, CopyProjectAppFormOutput>();
    const { field } = useController({ control, name });

    return (
        <CopyFormRow
            label={label}
            source={<span className="text-muted-foreground">-</span>}
        >
            <div className="flex h-9 items-center">
                <Checkbox
                    checked={field.value}
                    disabled={readOnly}
                    aria-label={`Copy ${label}`}
                    onCheckedChange={checked => {
                        field.onChange(checked === true);
                    }}
                />
            </div>
        </CopyFormRow>
    );
}

function TargetEnvSelect({ envs, readOnly, error }: TargetEnvSelectProps) {
    const { control } = useFormContext<CopyProjectAppFormInput, unknown, CopyProjectAppFormOutput>();
    const { field } = useController({ control, name: "targetEnv" });
    const targetEnv = field.value.trim();
    const envOptions = useMemo(() => {
        const nonEmptyEnvs = envs.filter(env => env.name.trim() !== "");

        if (!targetEnv || nonEmptyEnvs.some(env => env.name === targetEnv)) {
            return nonEmptyEnvs;
        }

        return [
            {
                name: targetEnv,
                color: FALLBACK_ENV_COLOR,
            },
            ...nonEmptyEnvs,
        ];
    }, [envs, targetEnv]);
    const selectedEnv = targetEnv ? envOptions.find(env => env.name === targetEnv) : null;

    return (
        <CopyFormRow
            label="Env"
            source={<SourceEnvBadge />}
            error={error}
        >
            <Select
                value={targetEnv}
                disabled={readOnly || envOptions.length === 0}
                onValueChange={field.onChange}
            >
                <SelectTrigger
                    aria-invalid={Boolean(error)}
                    className="px-2"
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
        </CopyFormRow>
    );
}

function SourceEnvBadge() {
    const { control } = useFormContext<CopyProjectAppFormInput, unknown, CopyProjectAppFormOutput>();
    const sourceEnv = useWatch({ control, name: "sourceEnv" });

    return sourceEnv ? <ProjectEnvBadge name={sourceEnv} /> : <span className="text-muted-foreground">-</span>;
}

function TargetStatusSelect({ readOnly, error }: TargetStatusSelectProps) {
    const { control } = useFormContext<CopyProjectAppFormInput, unknown, CopyProjectAppFormOutput>();
    const { field } = useController({ control, name: "targetStatus" });
    const sourceStatus = useWatch({ control, name: "sourceStatus" });

    return (
        <CopyFormRow
            label="Status"
            source={<ProjectAppStatusBadge status={sourceStatus} />}
            error={error}
        >
            <Select
                value={field.value}
                disabled={readOnly}
                onValueChange={field.onChange}
            >
                <SelectTrigger
                    aria-invalid={Boolean(error)}
                    className="px-2"
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
        </CopyFormRow>
    );
}

function DomainTargetField({ index, readOnly, error }: DomainTargetFieldProps) {
    const { control } = useFormContext<CopyProjectAppFormInput, unknown, CopyProjectAppFormOutput>();
    const sourceDomain = useWatch({
        control,
        name: `copyHttpSettings.copyDomainSettings.${index}.sourceDomain`,
    });
    const { field } = useController({
        control,
        name: `copyHttpSettings.copyDomainSettings.${index}.targetDomain`,
    });

    return (
        <CopyFormRow
            label="Domain"
            source={sourceDomain}
            error={error}
        >
            <Input
                {...field}
                disabled={readOnly}
                aria-invalid={Boolean(error)}
            />
        </CopyFormRow>
    );
}

function DomainSslCertField({ projectId, index, readOnly }: DomainSslCertFieldProps) {
    const [searchQuery, setSearchQuery] = useState("");
    const { control } = useFormContext<CopyProjectAppFormInput, unknown, CopyProjectAppFormOutput>();
    const sourceSslCert = useWatch({
        control,
        name: `copyHttpSettings.copyDomainSettings.${index}.sourceSslCert`,
    });
    const targetDomain = useWatch({
        control,
        name: `copyHttpSettings.copyDomainSettings.${index}.targetDomain`,
    });
    const normalizedTargetDomain = typeof targetDomain === "string" ? targetDomain.trim() : "";
    const hasTargetDomain = normalizedTargetDomain !== "";
    const { field } = useController({
        control,
        name: `copyHttpSettings.copyDomainSettings.${index}.targetSslCert`,
    });

    const {
        data: { data: sslCerts } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = ProjectSslCertQueries.useFindManyPaginated(
        {
            projectID: projectId,
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

        if (field.value?.id && !options.some(option => option.value.id === field.value?.id)) {
            return [
                {
                    value: { id: field.value.id, name: field.value.name, domain: "" },
                    label: field.value.name,
                },
                ...options,
            ];
        }

        return options;
    }, [field.value, hasTargetDomain, sslCerts]);

    return (
        <CopyFormRow
            label="SSL Cert"
            source={sourceSslCert?.name ?? <span className="text-muted-foreground">-</span>}
        >
            <Combobox
                options={comboboxOptions}
                value={hasTargetDomain ? (field.value?.id ?? null) : null}
                onChange={(_, option) => {
                    if (readOnly || !hasTargetDomain) {
                        return;
                    }

                    field.onChange(option ? { id: option.id, name: option.name } : null);
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
                allowClear
                disabled={readOnly || !hasTargetDomain}
            />
        </CopyFormRow>
    );
}

export function CopyProjectAppForm({
    projectId,
    envs,
    defaultValues,
    isPending,
    readOnly = false,
    onSubmit,
    onHasChanges,
}: Props) {
    const envNames = useMemo(() => envs.map(env => env.name).filter(name => name.trim() !== ""), [envs]);
    const schema = useMemo(() => createCopyProjectAppFormSchema(envNames), [envNames]);

    const methods = useForm<CopyProjectAppFormInput, unknown, CopyProjectAppFormOutput>({
        defaultValues,
        resolver: zodResolver(schema),
        mode: "onSubmit",
    });

    const {
        control,
        handleSubmit,
        reset,
        formState: { errors, isDirty },
    } = methods;
    const { fields } = useFieldArray({
        control,
        name: "copyHttpSettings.copyDomainSettings",
    });
    const { field: targetName } = useController({ control, name: "targetName" });
    const sourceName = useWatch({ control, name: "sourceName" });

    useEffect(() => {
        reset(defaultValues);
    }, [defaultValues, reset]);

    useEffect(() => {
        onHasChanges?.(readOnly ? false : isDirty);
    }, [isDirty, onHasChanges, readOnly]);

    function onValid(values: CopyProjectAppFormOutput) {
        if (readOnly) {
            return;
        }

        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CopyProjectAppFormInput>) {
        // React Hook Form handles field rendering.
    }

    return (
        <FormProvider {...methods}>
            <form
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onValid, onInvalid)(event);
                }}
                className="min-h-0 flex flex-1 flex-col"
            >
                <DialogBody className="flex max-h-[70vh] flex-col gap-5 overflow-y-auto">
                    <fieldset
                        disabled={readOnly}
                        className="m-0 flex min-w-0 flex-col gap-5 border-0 p-0"
                    >
                        <CopyFormRow
                            label="Name"
                            source={sourceName}
                            error={errors.targetName}
                        >
                            <Input
                                {...targetName}
                                disabled={readOnly}
                                aria-invalid={Boolean(errors.targetName)}
                            />
                        </CopyFormRow>

                        <TargetEnvSelect
                            envs={envs}
                            readOnly={readOnly}
                            error={errors.targetEnv}
                        />

                        <TargetStatusSelect
                            readOnly={readOnly}
                            error={errors.targetStatus}
                        />

                        {fields.map((field, index) => (
                            <React.Fragment key={field.id}>
                                <DomainTargetField
                                    index={index}
                                    readOnly={readOnly}
                                    error={errors.copyHttpSettings?.copyDomainSettings?.[index]?.targetDomain}
                                />
                                <DomainSslCertField
                                    projectId={projectId}
                                    index={index}
                                    readOnly={readOnly}
                                />
                            </React.Fragment>
                        ))}

                        <CopyFlagRow
                            label="Deployment Settings"
                            name="copyDeploymentSettings.copy"
                            readOnly={readOnly}
                        />
                        <CopyFlagRow
                            label="Env Variables"
                            name="copyEnvVars.copy"
                            readOnly={readOnly}
                        />
                        <CopyFlagRow
                            label="Config Files"
                            name="copyConfigFiles.copy"
                            readOnly={readOnly}
                        />
                        <CopyFlagRow
                            label="Secrets"
                            name="copySecrets.copy"
                            readOnly={readOnly}
                        />
                        <CopyFlagRow
                            label="Health Checks"
                            name="copyHealthChecks.copy"
                            readOnly={readOnly}
                        />
                        <CopyFlagRow
                            label="Scheduled Jobs"
                            name="copySchedJobs.copy"
                            readOnly={readOnly}
                        />
                    </fieldset>
                </DialogBody>
                <DialogActionFooter>
                    <Button
                        type="submit"
                        isLoading={isPending}
                        disabled={readOnly}
                    >
                        Copy
                    </Button>
                </DialogActionFooter>
            </form>
        </FormProvider>
    );
}

interface CopyFormRowProps {
    label: string;
    source: ReactNode;
    error?: HookFormFieldError;
    children: ReactNode;
}

interface CopyFlagRowProps {
    label: string;
    name:
        | "copyDeploymentSettings.copy"
        | "copyEnvVars.copy"
        | "copyConfigFiles.copy"
        | "copySecrets.copy"
        | "copyHealthChecks.copy"
        | "copySchedJobs.copy";
    readOnly: boolean;
}

interface TargetEnvSelectProps {
    envs: ProjectEnvEntity[];
    readOnly: boolean;
    error?: HookFormFieldError;
}

interface TargetStatusSelectProps {
    readOnly: boolean;
    error?: HookFormFieldError;
}

interface DomainTargetFieldProps {
    index: number;
    readOnly: boolean;
    error?: HookFormFieldError;
}

interface DomainSslCertFieldProps {
    projectId: string;
    index: number;
    readOnly: boolean;
}

interface Props {
    projectId: string;
    envs: ProjectEnvEntity[];
    defaultValues: CopyProjectAppFormInput;
    isPending: boolean;
    readOnly?: boolean;
    onSubmit: (values: CopyProjectAppFormOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
}

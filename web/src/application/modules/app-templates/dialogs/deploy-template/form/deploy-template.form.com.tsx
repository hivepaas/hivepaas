import React, { useEffect, useMemo, useRef, useState } from "react";

import { cn } from "@/lib/utils";
import { AlertTriangle, Boxes, Database, Globe, HardDrive, Loader2, Rocket, Search, Sparkles } from "lucide-react";
import { type Control, Controller, type UseFormSetValue, useForm } from "react-hook-form";
import { type ProjectEnvEntity } from "~/projects/domain";
import { ProjectEnvBadge } from "~/projects/module-shared/components";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DialogBody } from "@/components/ui/dialog";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/input-password";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import {
    type AppTemplateDependency,
    type AppTemplateDetail,
    type AppTemplateImageTag,
    type AppTemplateParam,
    type AppTemplateParamOption,
    type AppTemplateVariant,
    type AppTemplateVersionSummary,
    type CreateAppFromTemplateReq,
    appTemplatesApi,
} from "../../../api";

export interface DeployTemplateFormProps {
    template: AppTemplateDetail;
    projectId: string;
    envs: ProjectEnvEntity[];
    clusterVolumes?: { id: string; name: string }[];
    initialEnv?: string;
    initialVersion?: string;
    initialVariant?: string;
    isPending: boolean;
    readOnly?: boolean;
    onSubmit: (values: CreateAppFromTemplateReq) => void;
    onCancel: () => void;
}

interface FormState {
    name: string;
    env: string;
    version: string;
    variant: string;
    imageOverride: string;
    params: Record<string, unknown>;
    dependencyParams: Record<string, Record<string, string>>;
}

function generateRandomPassword(length: number = 24): string {
    const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_!@#$%";
    const array = new Uint8Array(length);
    window.crypto.getRandomValues(array);
    return Array.from(array, byte => chars[byte % chars.length]).join("");
}

/**
 * Computes default domain for template parameters based on current window domain.
 * - Current domain `abc.xyz.tuv` -> `<template-name>.xyz.tuv`
 * - Current domain `tuv` -> `<template-name>.tuv`
 */
function getDefaultDomain(templateName: string): string {
    if (typeof window === "undefined" || !window.location.hostname) {
        return "";
    }
    let hostname = window.location.hostname.trim().toLowerCase();
    if (hostname.includes(":")) {
        hostname = hostname.split(":")[0] ?? "";
    }
    if (!hostname) {
        return "";
    }

    const tName = templateName.trim().toLowerCase() || "app";
    const segments = hostname.split(".");
    if (segments.length > 1) {
        return `${tName}.${segments.slice(1).join(".")}`;
    }
    return `${tName}.${hostname}`;
}

export function DeployTemplateForm({
    template,
    projectId,
    envs,
    clusterVolumes = [],
    initialEnv,
    initialVersion,
    initialVariant,
    isPending,
    readOnly = false,
    onSubmit,
    onCancel,
}: DeployTemplateFormProps) {
    // 1. Initial defaults calculation
    const defaultEnv = useMemo(() => {
        if (initialEnv && envs.some((e: ProjectEnvEntity) => e.name === initialEnv)) {
            return initialEnv;
        }
        return envs[0]?.name ?? "";
    }, [envs, initialEnv]);

    const { versions } = template;
    const defaultVersion = useMemo(() => {
        if (initialVersion && versions.some((v: AppTemplateVersionSummary) => v.name === initialVersion)) {
            return initialVersion;
        }
        return versions.find((v: AppTemplateVersionSummary) => v.default)?.name ?? versions[0]?.name ?? "";
    }, [versions, initialVersion]);

    const { variants } = template;
    const defaultVariant = useMemo(() => {
        if (initialVariant && variants.some((v: AppTemplateVariant) => v.name === initialVariant)) {
            return initialVariant;
        }
        return variants.find((v: AppTemplateVariant) => v.default)?.name ?? variants[0]?.name ?? "";
    }, [variants, initialVariant]);

    // Initial parameter values (domain is left empty by default)
    const defaultParams = useMemo(() => {
        const result: Record<string, unknown> = {};
        const firstVol = clusterVolumes[0];
        for (const param of template.parameters) {
            if (param.default !== undefined && param.default !== null) {
                result[param.name] = param.default;
            } else if (param.type === "bool") {
                result[param.name] = false;
            } else if (param.type === "volume" && firstVol) {
                result[param.name] = firstVol.id;
            } else {
                result[param.name] = "";
            }
        }
        return result;
    }, [template.parameters, clusterVolumes]);

    // Initial dependency params
    const defaultDependencyParams = useMemo(() => {
        const result: Record<string, Record<string, string>> = {};
        const firstVol = clusterVolumes[0];
        for (const dep of template.dependencies ?? []) {
            const depRecord: Record<string, string> = {};
            result[dep.name] = depRecord;
            for (const p of dep.parameters ?? []) {
                if (p.type === "volume" && firstVol) {
                    depRecord[p.name] = firstVol.id;
                } else if (p.default !== undefined && p.default !== null) {
                    const defaultVal =
                        typeof p.default === "string" || typeof p.default === "number" ? String(p.default) : "";
                    depRecord[p.name] = defaultVal;
                } else {
                    depRecord[p.name] = "";
                }
            }
        }
        return result;
    }, [template.dependencies, clusterVolumes]);

    // 2. React Hook Form Setup
    const { control, handleSubmit, watch, setValue } = useForm<FormState>({
        defaultValues: {
            name: template.name,
            env: defaultEnv,
            version: defaultVersion,
            variant: defaultVariant,
            imageOverride: "",
            params: defaultParams,
            dependencyParams: defaultDependencyParams,
        },
        mode: "onSubmit",
    });

    const selectedVersion = watch("version");
    const selectedVariant = watch("variant");
    const imageOverride = watch("imageOverride");

    // 3. Registry Image Tags State & Fetcher
    const [scannedTags, setScannedTags] = useState<AppTemplateImageTag[]>([]);
    const [isFetchingTags, setIsFetchingTags] = useState<boolean>(false);

    // Clear scanned tags whenever variant changes, and reset version if it was a custom scanned tag
    const prevVariantRef = useRef<string | undefined>(selectedVariant);
    useEffect(() => {
        if (prevVariantRef.current !== undefined && prevVariantRef.current !== selectedVariant) {
            setScannedTags([]);
            const isOfficial = versions.some((v: AppTemplateVersionSummary) => v.name === selectedVersion);
            if (!isOfficial) {
                const fallbackVer =
                    versions.find(
                        (v: AppTemplateVersionSummary) =>
                            (v.variants.length === 0 || v.variants.includes(selectedVariant)) && v.default,
                    )?.name ??
                    versions.find(
                        (v: AppTemplateVersionSummary) =>
                            v.variants.length === 0 || v.variants.includes(selectedVariant),
                    )?.name ??
                    versions.find((v: AppTemplateVersionSummary) => v.default)?.name ??
                    versions[0]?.name ??
                    "";
                setValue("version", fallbackVer);
                setValue("imageOverride", "");
            }
        }
        prevVariantRef.current = selectedVariant;
    }, [selectedVariant, versions, selectedVersion, setValue]);

    const handleFetchMoreTags = async () => {
        if (isFetchingTags) return;
        setIsFetchingTags(true);
        try {
            const officialVer = versions.some((v: AppTemplateVersionSummary) => v.name === selectedVersion)
                ? selectedVersion
                : (versions.find(
                      (v: AppTemplateVersionSummary) =>
                          (v.variants.length === 0 || v.variants.includes(selectedVariant)) && v.default,
                  )?.name ??
                  versions.find(
                      (v: AppTemplateVersionSummary) => v.variants.length === 0 || v.variants.includes(selectedVariant),
                  )?.name ??
                  versions.find((v: AppTemplateVersionSummary) => v.default)?.name ??
                  versions[0]?.name ??
                  "");

            const data = await appTemplatesApi.getImageTags({
                templateName: template.name,
                version: officialVer,
                variant: selectedVariant,
            });
            setScannedTags(data.tags);
        } catch {
            // Silently fall back or keep existing
        } finally {
            setIsFetchingTags(false);
        }
    };

    // Determine if the currently chosen version is a custom tag (not in template versions)
    const isOfficialTemplateVersion = useMemo(() => {
        return versions.some((v: AppTemplateVersionSummary) => v.name === selectedVersion);
    }, [versions, selectedVersion]);

    const isCustomVersion = !isOfficialTemplateVersion || Boolean(imageOverride);

    // Variants available for currently selected version
    const availableVariants = useMemo(() => {
        const verObj = versions.find((v: AppTemplateVersionSummary) => v.name === selectedVersion);
        if (!verObj?.variants || verObj.variants.length === 0) {
            return variants;
        }
        return variants.filter((v: AppTemplateVariant) => verObj.variants.includes(v.name));
    }, [versions, variants, selectedVersion]);

    // Form submission handler
    const handleFormSubmit = (data: FormState) => {
        if (readOnly || isPending) return;

        // Clean params: omit empty optional fields
        const cleanedParams: Record<string, unknown> = {};
        for (const [key, value] of Object.entries(data.params)) {
            if (value !== "" && value !== undefined && value !== null) {
                cleanedParams[key] = value;
            }
        }

        // Clean dependency params
        const cleanedDepParams: Record<string, Record<string, unknown>> = {};
        for (const [depName, depVals] of Object.entries(data.dependencyParams)) {
            const innerClean: Record<string, unknown> = {};
            for (const [k, v] of Object.entries(depVals)) {
                if (v !== "") {
                    innerClean[k] = v;
                }
            }
            if (Object.keys(innerClean).length > 0) {
                cleanedDepParams[depName] = innerClean;
            }
        }

        onSubmit({
            projectID: projectId,
            projectEnv: data.env,
            name: data.name.trim(),
            template: template.name,
            version: isOfficialTemplateVersion ? data.version : undefined,
            variant: data.variant ? data.variant : undefined,
            imageOverride: isOfficialTemplateVersion
                ? undefined
                : data.imageOverride
                  ? data.imageOverride
                  : data.version,
            params: Object.keys(cleanedParams).length > 0 ? cleanedParams : undefined,
            dependencyParams: Object.keys(cleanedDepParams).length > 0 ? cleanedDepParams : undefined,
        });
    };

    return (
        <form
            onSubmit={e => {
                void handleSubmit(handleFormSubmit)(e);
            }}
            className="flex flex-col h-full overflow-hidden"
        >
            <DialogBody className="space-y-6 px-3.5 py-5 overflow-y-auto">
                {/* ========================================================================= */}
                {/* SECTION 1: APPLICATION BASICS & VERSION / VARIANT SELECTION               */}
                {/* ========================================================================= */}
                <div className="space-y-4 rounded-xl border border-border/70 bg-card/60 p-4 shadow-2xs">
                    <div className="flex items-center gap-2 border-b border-border/40 pb-2.5">
                        <Boxes className="size-4 text-amber-500" />
                        <h4 className="text-sm font-semibold text-foreground tracking-tight">Application & Version</h4>
                    </div>

                    {/* Row 1: App Name & Environment */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                        {/* App Name */}
                        <Controller
                            name="name"
                            control={control}
                            rules={{
                                required: "App name is required",
                                maxLength: { value: 100, message: "Maximum 100 characters" },
                                pattern: {
                                    value: /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/,
                                    message: "Lowercase alphanumeric characters and hyphens only",
                                },
                            }}
                            render={({ field, fieldState }) => (
                                <Field data-invalid={fieldState.invalid}>
                                    <FieldLabel className="text-xs font-medium text-foreground">
                                        App Name <span className="text-destructive">*</span>
                                    </FieldLabel>
                                    <Input
                                        {...field}
                                        placeholder="e.g. my-app"
                                        disabled={readOnly || isPending}
                                        className="h-9 text-xs"
                                    />
                                    {fieldState.error && (
                                        <FieldError className="text-[11px] text-destructive">
                                            {fieldState.error.message}
                                        </FieldError>
                                    )}
                                </Field>
                            )}
                        />

                        {/* Environment */}
                        <Controller
                            name="env"
                            control={control}
                            rules={{ required: "Environment is required" }}
                            render={({ field }) => (
                                <Field>
                                    <FieldLabel className="text-xs font-medium text-foreground">
                                        Target Environment <span className="text-destructive">*</span>
                                    </FieldLabel>
                                    <Select
                                        value={field.value}
                                        onValueChange={field.onChange}
                                        disabled={readOnly || isPending}
                                    >
                                        <SelectTrigger className="h-9 text-xs">
                                            <SelectValue placeholder="Select environment" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {envs.map(env => (
                                                <SelectItem
                                                    key={env.id}
                                                    value={env.name}
                                                    className="text-xs"
                                                >
                                                    <div className="flex items-center gap-2">
                                                        <ProjectEnvBadge
                                                            name={env.name}
                                                            color={env.color}
                                                            size="xs"
                                                        />
                                                    </div>
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </Field>
                            )}
                        />
                    </div>

                    {/* Row 2: Variant Selection */}
                    {variants.length > 0 && (
                        <div className="space-y-1.5 pt-1">
                            <span className="text-xs font-medium text-foreground block">
                                Distribution Variant <span className="text-destructive">*</span>
                            </span>
                            <Controller
                                name="variant"
                                control={control}
                                render={({ field }) => (
                                    <div className="flex flex-wrap items-center gap-2">
                                        {variants.map((v: AppTemplateVariant) => {
                                            const isAvailable = availableVariants.some(
                                                (av: AppTemplateVariant) => av.name === v.name,
                                            );
                                            const isSelected = field.value === v.name;

                                            return (
                                                <button
                                                    key={v.name}
                                                    type="button"
                                                    disabled={!isAvailable || readOnly || isPending}
                                                    onClick={() => {
                                                        if (field.value !== v.name) {
                                                            field.onChange(v.name);
                                                            setScannedTags([]);
                                                            const isOfficial = versions.some(
                                                                (ver: AppTemplateVersionSummary) =>
                                                                    ver.name === selectedVersion,
                                                            );
                                                            if (!isOfficial) {
                                                                const fallbackVer =
                                                                    versions.find(
                                                                        (ver: AppTemplateVersionSummary) =>
                                                                            (ver.variants.length === 0 ||
                                                                                ver.variants.includes(v.name)) &&
                                                                            ver.default,
                                                                    )?.name ??
                                                                    versions.find(
                                                                        (ver: AppTemplateVersionSummary) =>
                                                                            ver.variants.length === 0 ||
                                                                            ver.variants.includes(v.name),
                                                                    )?.name ??
                                                                    versions.find(
                                                                        (ver: AppTemplateVersionSummary) => ver.default,
                                                                    )?.name ??
                                                                    versions[0]?.name ??
                                                                    "";
                                                                setValue("version", fallbackVer);
                                                                setValue("imageOverride", "");
                                                            }
                                                        }
                                                    }}
                                                    className={cn(
                                                        "relative inline-flex items-center gap-2 rounded-lg border px-3 py-1.5 text-xs font-medium transition-all",
                                                        isSelected
                                                            ? "border-amber-500 bg-amber-500/10 text-amber-600 dark:text-amber-400 font-semibold shadow-2xs"
                                                            : isAvailable
                                                              ? "border-border/80 bg-muted/40 text-muted-foreground hover:bg-muted/80 hover:text-foreground"
                                                              : "border-border/40 bg-muted/20 text-muted-foreground/40 cursor-not-allowed",
                                                    )}
                                                >
                                                    <span
                                                        className={cn(
                                                            "size-2 rounded-full",
                                                            isSelected
                                                                ? "bg-amber-500"
                                                                : isAvailable
                                                                  ? "bg-muted-foreground/50"
                                                                  : "bg-muted-foreground/20",
                                                        )}
                                                    />
                                                    <span>{v.title ? v.title : v.name}</span>
                                                    {v.description && (
                                                        <span className="text-[11px] text-muted-foreground/80 font-normal">
                                                            ({v.description})
                                                        </span>
                                                    )}
                                                </button>
                                            );
                                        })}
                                    </div>
                                )}
                            />
                        </div>
                    )}

                    {/* Row 3: Version Selection & Tag Override Button */}
                    <div className="space-y-2 pt-1">
                        <span className="text-xs font-medium text-foreground block">
                            Version <span className="text-destructive">*</span>
                        </span>
                        <div className="flex items-center gap-2.5">
                            <div className="flex-1">
                                <Controller
                                    name="version"
                                    control={control}
                                    rules={{ required: "Version is required" }}
                                    render={({ field }) => (
                                        <Select
                                            value={field.value}
                                            onValueChange={val => {
                                                field.onChange(val);
                                                // If chosen from scanned tags, record imageOverride
                                                const matchedTag = scannedTags.find(t => t.tag === val);
                                                if (matchedTag) {
                                                    setValue("imageOverride", matchedTag.image);
                                                } else {
                                                    setValue("imageOverride", "");
                                                }
                                            }}
                                            disabled={readOnly || isPending}
                                        >
                                            <SelectTrigger className="h-9 text-xs w-full bg-background border-border/80">
                                                <SelectValue placeholder="Select version" />
                                            </SelectTrigger>
                                            <SelectContent className="max-h-[280px]">
                                                {/* Official Template Versions */}
                                                <div className="px-2 py-1 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                                                    Template Versions
                                                </div>
                                                {versions.map((v: AppTemplateVersionSummary) => (
                                                    <SelectItem
                                                        key={v.name}
                                                        value={v.name}
                                                        className="text-xs"
                                                    >
                                                        <div className="flex items-center justify-between gap-3 w-full">
                                                            <span>
                                                                v{v.name} {v.release ? `(${v.release})` : ""}
                                                            </span>
                                                            {v.default && (
                                                                <Badge
                                                                    variant="secondary"
                                                                    className="text-[10px] px-1 py-0 h-4 bg-muted/80 text-muted-foreground"
                                                                >
                                                                    latest
                                                                </Badge>
                                                            )}
                                                        </div>
                                                    </SelectItem>
                                                ))}

                                                {/* Scanned Registry Tags */}
                                                {scannedTags.length > 0 && (
                                                    <>
                                                        <div className="mt-2 border-t border-border/40 px-2 pt-2 pb-1 text-[11px] font-semibold uppercase tracking-wider text-amber-600 dark:text-amber-400">
                                                            Scanned Registry Tags ({scannedTags.length})
                                                        </div>
                                                        {scannedTags.map(tagObj => (
                                                            <SelectItem
                                                                key={tagObj.tag}
                                                                value={tagObj.tag}
                                                                className="text-xs"
                                                            >
                                                                <div className="flex items-center justify-between gap-3 w-full">
                                                                    <span>{tagObj.tag}</span>
                                                                    {tagObj.newer && (
                                                                        <Badge
                                                                            variant="outline"
                                                                            className="text-[10px] px-1 py-0 h-4 border-amber-500/40 text-amber-600 dark:text-amber-400"
                                                                        >
                                                                            newer
                                                                        </Badge>
                                                                    )}
                                                                </div>
                                                            </SelectItem>
                                                        ))}
                                                    </>
                                                )}
                                            </SelectContent>
                                        </Select>
                                    )}
                                />
                            </div>

                            {/* Tag scan button on the right: Check More Tags */}
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                disabled={isFetchingTags || readOnly || isPending}
                                onClick={() => {
                                    void handleFetchMoreTags();
                                }}
                                className="h-9 shrink-0 text-xs font-medium border-border/80 hover:bg-muted/70 px-3 min-w-[130px]"
                            >
                                {isFetchingTags ? (
                                    <>
                                        <Loader2 className="size-3.5 mr-1.5 animate-spin" />
                                        Scanning...
                                    </>
                                ) : (
                                    <>
                                        <Search className="size-3.5 mr-1.5 text-muted-foreground" />
                                        Check More Tags
                                    </>
                                )}
                            </Button>
                        </div>

                        {/* Custom Image Version Warning Banner (Style: This volume is not pinned to a node..) */}
                        {isCustomVersion && (
                            <div className="rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-400 flex items-start gap-2.5 transition-all">
                                <AlertTriangle className="size-4 shrink-0 mt-0.5 text-amber-600 dark:text-amber-400" />
                                <div className="text-xs leading-relaxed">
                                    <span className="font-semibold">Using a custom image version.</span> This version is
                                    not tested or maintained as part of the official template. Using a custom version or
                                    moving tag may cause the deployment to fail or the application to behave
                                    unexpectedly.
                                </div>
                            </div>
                        )}
                    </div>
                </div>

                {/* ========================================================================= */}
                {/* SECTION 2: PARAMETERS CONFIGURATION (1 ROW PER PARAMETER)               */}
                {/* ========================================================================= */}
                <div className="space-y-4 rounded-xl border border-border/70 bg-card/60 p-4 shadow-2xs">
                    <div className="flex items-center justify-between border-b border-border/40 pb-2.5">
                        <div className="flex items-center gap-2">
                            <Database className="size-4 text-amber-500" />
                            <h4 className="text-sm font-semibold text-foreground tracking-tight">
                                Parameters Configuration
                            </h4>
                        </div>
                        <span className="text-[11px] text-muted-foreground">
                            {template.parameters.length} parameters defined
                        </span>
                    </div>

                    {template.parameters.length === 0 && (
                        <p className="text-xs text-muted-foreground py-2 italic">
                            This template does not require any additional parameters.
                        </p>
                    )}

                    <FieldGroup className="space-y-4">
                        {template.parameters.map((param: AppTemplateParam) => (
                            <ParameterRow
                                key={param.name}
                                param={param}
                                templateName={template.name}
                                control={control}
                                setValue={setValue}
                                clusterVolumes={clusterVolumes}
                                readOnly={readOnly || isPending}
                            />
                        ))}
                    </FieldGroup>
                </div>

                {/* ========================================================================= */}
                {/* SECTION 3: DEPENDENT SERVICES (IF TEMPLATE DECLARES ANY)                  */}
                {/* ========================================================================= */}
                {template.dependencies && template.dependencies.length > 0 && (
                    <div className="space-y-4 rounded-xl border border-border/70 bg-card/60 p-4 shadow-2xs">
                        <div className="flex items-center gap-2 border-b border-border/40 pb-2.5">
                            <HardDrive className="size-4 text-amber-500" />
                            <h4 className="text-sm font-semibold text-foreground tracking-tight">Dependent Services</h4>
                        </div>
                        <p className="text-xs text-muted-foreground">
                            This template automatically deploys companion services needed for operation.
                        </p>

                        <div className="space-y-4">
                            {template.dependencies.map((dep: AppTemplateDependency) => (
                                <div
                                    key={dep.name}
                                    className="rounded-lg border border-border/60 bg-muted/20 p-3.5 space-y-3"
                                >
                                    <div className="flex items-center justify-between">
                                        <span className="text-xs font-semibold text-foreground">
                                            {dep.title}{" "}
                                            <span className="text-muted-foreground font-normal">({dep.template})</span>
                                        </span>
                                        <Badge
                                            variant="outline"
                                            className="text-[10px] font-mono px-1.5 py-0"
                                        >
                                            {dep.version ?? "default"}
                                        </Badge>
                                    </div>

                                    {dep.parameters?.map((p: AppTemplateParam) => (
                                        <Controller
                                            key={p.name}
                                            name={`dependencyParams.${dep.name}.${p.name}`}
                                            control={control}
                                            render={({ field }) => (
                                                <Field className="space-y-1">
                                                    <div className="space-y-0.5">
                                                        <FieldLabel className="text-xs font-medium text-foreground">
                                                            {p.title ? p.title : p.name}{" "}
                                                            {!p.optional && <span className="text-destructive">*</span>}
                                                        </FieldLabel>
                                                        {p.description && (
                                                            <p className="text-xs text-muted-foreground leading-relaxed">
                                                                {p.description}
                                                            </p>
                                                        )}
                                                    </div>

                                                    {p.type === "volume" && clusterVolumes.length > 0 ? (
                                                        <Select
                                                            value={field.value}
                                                            onValueChange={field.onChange}
                                                            disabled={readOnly || isPending}
                                                        >
                                                            <SelectTrigger className="h-9 text-xs">
                                                                <SelectValue placeholder="Select cluster volume" />
                                                            </SelectTrigger>
                                                            <SelectContent>
                                                                {clusterVolumes.map(vol => (
                                                                    <SelectItem
                                                                        key={vol.id}
                                                                        value={vol.id}
                                                                        textValue={vol.name}
                                                                        className="text-xs"
                                                                    >
                                                                        <div className="flex items-center gap-2">
                                                                            <HardDrive className="size-3.5 text-muted-foreground" />
                                                                            <span>{vol.name}</span>
                                                                        </div>
                                                                    </SelectItem>
                                                                ))}
                                                            </SelectContent>
                                                        </Select>
                                                    ) : (
                                                        <Input
                                                            {...field}
                                                            value={field.value}
                                                            placeholder={p.description ? p.description : "Enter value"}
                                                            disabled={readOnly || isPending}
                                                            className="h-9 text-xs"
                                                        />
                                                    )}
                                                </Field>
                                            )}
                                        />
                                    ))}
                                </div>
                            ))}
                        </div>
                    </div>
                )}
            </DialogBody>

            {/* Fixed Action Footer */}
            <div className="shrink-0 border-t border-border/50 bg-background/95 px-3.5 py-3.5 flex items-center justify-between">
                <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={onCancel}
                    disabled={isPending}
                    className="h-9 px-4 text-xs font-medium"
                >
                    Cancel
                </Button>

                <Button
                    type="submit"
                    size="sm"
                    disabled={readOnly || isPending}
                    className="h-9 px-5 text-xs font-semibold bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 text-white shadow-xs disabled:opacity-50"
                >
                    {isPending ? (
                        <>
                            <Loader2 className="size-3.5 mr-1.5 animate-spin" />
                            Deploying...
                        </>
                    ) : (
                        <>
                            <Rocket className="size-3.5 mr-1.5" />
                            Deploy Template
                        </>
                    )}
                </Button>
            </div>
        </form>
    );
}

// =========================================================================
// SINGLE PARAMETER ROW COMPONENT (1 ROW PER PARAMETER)
// =========================================================================
interface ParameterRowProps {
    param: AppTemplateParam;
    templateName: string;
    control: Control<FormState>;
    setValue: UseFormSetValue<FormState>;
    clusterVolumes: { id: string; name: string }[];
    readOnly?: boolean;
}

function ParameterRow({ param, templateName, control, setValue, clusterVolumes, readOnly }: ParameterRowProps) {
    const isRequired = !param.optional;

    return (
        <Controller
            name={`params.${param.name}`}
            control={control}
            rules={{
                required:
                    isRequired && param.type !== "secret" && param.type !== "bool"
                        ? `${param.title ? param.title : param.name} is required`
                        : false,
                minLength: param.minLength
                    ? { value: param.minLength, message: `Minimum ${param.minLength} characters` }
                    : undefined,
                maxLength: param.maxLength
                    ? { value: param.maxLength, message: `Maximum ${param.maxLength} characters` }
                    : undefined,
                pattern: param.pattern
                    ? { value: new RegExp(param.pattern), message: "Format does not match pattern" }
                    : undefined,
            }}
            render={({ field, fieldState }) => {
                return (
                    <Field
                        data-invalid={fieldState.invalid}
                        className="space-y-1.5 py-1 border-b border-border/30 last:border-0 pb-3"
                    >
                        {/* Title, optional badge, and description */}
                        <div className="space-y-1">
                            <FieldLabel className="text-xs font-semibold text-foreground flex items-center gap-1.5">
                                <span>{param.title ? param.title : param.name}</span>
                                {isRequired && <span className="text-destructive">*</span>}
                                {param.optional && (
                                    <Badge
                                        variant="outline"
                                        className="text-[10px] px-1 py-0 h-4 text-muted-foreground font-normal"
                                    >
                                        optional
                                    </Badge>
                                )}
                            </FieldLabel>

                            {param.description && (
                                <p className="text-xs text-muted-foreground leading-relaxed">{param.description}</p>
                            )}
                        </div>

                        {/* Input Control by Type */}
                        <div>
                            {/* Type: Secret */}
                            {param.type === "secret" && (
                                <div className="space-y-1.5">
                                    <div className="flex items-center gap-2 w-full">
                                        <div className="flex-1 min-w-0">
                                            <PasswordInput
                                                value={typeof field.value === "string" ? field.value : ""}
                                                onChange={field.onChange}
                                                placeholder={
                                                    param.generated ? "Leave empty to auto-generate" : "Enter password"
                                                }
                                                disabled={readOnly}
                                                className="h-9 text-xs w-full"
                                            />
                                        </div>
                                        <Button
                                            type="button"
                                            variant="outline"
                                            size="sm"
                                            title="Generate secure password"
                                            disabled={readOnly}
                                            onClick={() => {
                                                const newPass = generateRandomPassword(32);
                                                setValue(`params.${param.name}`, newPass);
                                            }}
                                            className="h-9 px-3 text-xs font-medium border-border/80 hover:bg-muted/80 text-muted-foreground hover:text-foreground shrink-0"
                                        >
                                            <Sparkles className="size-3.5 mr-1 text-amber-500" />
                                            Generate
                                        </Button>
                                    </div>
                                    {param.generated && (
                                        <p className="text-[11px] text-muted-foreground/70">
                                            ℹ A secure password will be automatically generated by HivePaaS if left
                                            empty.
                                        </p>
                                    )}
                                </div>
                            )}

                            {/* Type: Volume */}
                            {param.type === "volume" && (
                                <div className="space-y-1">
                                    {clusterVolumes.length > 0 ? (
                                        <Select
                                            value={typeof field.value === "string" ? field.value : ""}
                                            onValueChange={field.onChange}
                                            disabled={readOnly}
                                        >
                                            <SelectTrigger className="h-9 text-xs w-full">
                                                <SelectValue placeholder="Select a cluster volume" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                {clusterVolumes.map(vol => (
                                                    <SelectItem
                                                        key={vol.id}
                                                        value={vol.id}
                                                        textValue={vol.name}
                                                        className="text-xs"
                                                    >
                                                        <div className="flex items-center gap-2">
                                                            <HardDrive className="size-3.5 text-muted-foreground" />
                                                            <span>{vol.name}</span>
                                                        </div>
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                    ) : (
                                        <Input
                                            value={typeof field.value === "string" ? field.value : ""}
                                            onChange={field.onChange}
                                            placeholder="Enter volume ID"
                                            disabled={readOnly}
                                            className="h-9 text-xs"
                                        />
                                    )}
                                </div>
                            )}

                            {/* Type: Select */}
                            {param.type === "select" && (
                                <Select
                                    value={typeof field.value === "string" ? field.value : ""}
                                    onValueChange={field.onChange}
                                    disabled={readOnly}
                                >
                                    <SelectTrigger className="h-9 text-xs w-full">
                                        <SelectValue placeholder="Select an option" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {param.options?.map((opt: AppTemplateParamOption) => (
                                            <SelectItem
                                                key={opt.value}
                                                value={opt.value}
                                                className="text-xs"
                                            >
                                                {opt.title ? opt.title : opt.value}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            )}

                            {/* Type: Size */}
                            {param.type === "size" && (
                                <div className="flex items-center gap-2">
                                    <Input
                                        value={typeof field.value === "string" ? field.value : ""}
                                        onChange={field.onChange}
                                        placeholder="e.g. 512MB, 1GB, 2GB"
                                        disabled={readOnly}
                                        className="h-9 text-xs flex-1"
                                    />
                                    <div className="hidden sm:flex items-center gap-1 shrink-0">
                                        {["512MB", "1GB", "2GB", "4GB"].map(preset => (
                                            <button
                                                key={preset}
                                                type="button"
                                                disabled={readOnly}
                                                onClick={() => {
                                                    setValue(`params.${param.name}`, preset);
                                                }}
                                                className={cn(
                                                    "h-7 px-2 text-[11px] rounded border transition-colors",
                                                    field.value === preset
                                                        ? "border-amber-500 bg-amber-500/10 text-amber-600 font-semibold"
                                                        : "border-border/60 bg-muted/40 hover:bg-muted text-muted-foreground",
                                                )}
                                            >
                                                {preset}
                                            </button>
                                        ))}
                                    </div>
                                </div>
                            )}

                            {/* Type: Bool */}
                            {param.type === "bool" && (
                                <div className="flex items-center gap-2 pt-0.5">
                                    <Checkbox
                                        checked={Boolean(field.value)}
                                        onCheckedChange={field.onChange}
                                        disabled={readOnly}
                                    />
                                    <span className="text-xs text-foreground">
                                        Enable {param.title ? param.title : param.name}
                                    </span>
                                </div>
                            )}

                            {/* Type: Int / Number */}
                            {(param.type === "int" || param.type === "integer") && (
                                <Input
                                    type="number"
                                    value={
                                        typeof field.value === "number" || typeof field.value === "string"
                                            ? String(field.value)
                                            : ""
                                    }
                                    onChange={e => {
                                        const num = e.target.value === "" ? "" : Number(e.target.value);
                                        field.onChange(num);
                                    }}
                                    min={param.min !== undefined ? Number(param.min) : undefined}
                                    max={param.max !== undefined ? Number(param.max) : undefined}
                                    disabled={readOnly}
                                    className="h-9 text-xs"
                                />
                            )}

                            {/* Type: Domain */}
                            {(param.type === "domain" || param.name.toLowerCase() === "domain") && (
                                <div className="flex items-center gap-2 w-full">
                                    <div className="flex-1 min-w-0">
                                        <Input
                                            value={typeof field.value === "string" ? field.value : ""}
                                            onChange={field.onChange}
                                            placeholder="e.g. app.example.com"
                                            disabled={readOnly}
                                            className="h-9 text-xs w-full"
                                        />
                                    </div>
                                    <Button
                                        type="button"
                                        variant="outline"
                                        size="sm"
                                        title="Suggest domain based on current hostname"
                                        disabled={readOnly}
                                        onClick={() => {
                                            const suggested = getDefaultDomain(templateName);
                                            if (suggested) {
                                                setValue(`params.${param.name}`, suggested, {
                                                    shouldValidate: true,
                                                    shouldDirty: true,
                                                });
                                            }
                                        }}
                                        className="h-9 px-3 text-xs font-medium border-border/80 hover:bg-muted/80 text-muted-foreground hover:text-foreground shrink-0"
                                    >
                                        <Globe className="size-3.5 mr-1 text-emerald-500" />
                                        Suggest Domain
                                    </Button>
                                </div>
                            )}

                            {/* Type: String or fallback */}
                            {param.type !== "secret" &&
                                param.type !== "volume" &&
                                param.type !== "select" &&
                                param.type !== "size" &&
                                param.type !== "bool" &&
                                param.type !== "int" &&
                                param.type !== "integer" &&
                                param.type !== "domain" &&
                                param.name.toLowerCase() !== "domain" && (
                                    <Input
                                        value={typeof field.value === "string" ? field.value : ""}
                                        onChange={field.onChange}
                                        placeholder={
                                            typeof param.default === "string" || typeof param.default === "number"
                                                ? String(param.default)
                                                : "Enter value"
                                        }
                                        disabled={readOnly}
                                        className="h-9 text-xs"
                                    />
                                )}
                        </div>

                        {fieldState.error && (
                            <FieldError className="text-[11px] text-destructive">{fieldState.error.message}</FieldError>
                        )}
                    </Field>
                );
            }}
        />
    );
}

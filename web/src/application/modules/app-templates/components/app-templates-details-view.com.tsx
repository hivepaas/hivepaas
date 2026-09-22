import React, { useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import {
    ArrowLeft,
    BookOpen,
    Boxes,
    CheckCircle2,
    Code2,
    ExternalLink,
    Globe,
    HardDrive,
    Info,
    KeyRound,
    Lock,
    Rocket,
    Scale,
    ShieldAlert,
    Sliders,
    Tag,
} from "lucide-react";
import ReactMarkdown from "react-markdown";
import { useParams } from "react-router";
import remarkGfm from "remark-gfm";

import { AppLink } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

import type { AppTemplateDependency, AppTemplateDetail, AppTemplateParam, AppTemplateSummary } from "../api";
import { useGetAppTemplate } from "../data";
import { useDeployTemplateDialogState } from "../dialogs";
import { describeCapabilities, grantedByTemplate } from "../utils";

interface AppTemplatesDetailsViewProps {
    templateName: string;
    templateSummary?: AppTemplateSummary;
    onBack: () => void;
    onSelectTag?: (tag: string) => void;
    className?: string;
}

function formatParamValue(val: unknown): string {
    if (val === null || val === undefined) return "";
    if (typeof val === "string") return val;
    if (typeof val === "number" || typeof val === "boolean") return val.toString();
    if (typeof val === "object") {
        try {
            return JSON.stringify(val);
        } catch {
            return "";
        }
    }
    return "";
}

/**
 * Renders template description markdown (paragraphs, lists, bold, inline code, fenced code
 * blocks, links) via react-markdown, styled to match the app's existing look.
 */
function MarkdownRenderer({ content, className }: { content: string; className?: string }) {
    if (!content) return null;

    return (
        <div className={cn("space-y-2 text-sm leading-relaxed text-foreground/90", className)}>
            <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                components={{
                    p: ({ children }) => <p className="m-0">{children}</p>,
                    strong: ({ children }) => <strong className="font-semibold text-foreground">{children}</strong>,
                    ul: ({ children }) => (
                        <ul className="list-disc pl-5 space-y-1 my-1.5 text-foreground/85">{children}</ul>
                    ),
                    ol: ({ children }) => (
                        <ol className="list-decimal pl-5 space-y-1 my-1.5 text-foreground/85">{children}</ol>
                    ),
                    li: ({ children }) => <li>{children}</li>,
                    a: ({ children, href }) => (
                        <a
                            href={href}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-primary underline underline-offset-2 hover:text-primary/80"
                        >
                            {children}
                        </a>
                    ),
                    pre: ({ children }) => (
                        <pre className="rounded-md bg-muted/60 p-3 my-1.5 overflow-x-auto text-[13px] font-mono leading-normal">
                            {children}
                        </pre>
                    ),
                    code: ({ className: codeClassName, children, node: _node, ...rest }) => {
                        // Fenced code blocks carry a `language-xxx` class from remark-gfm; inline
                        // `code` spans don't, so this is how the two get told apart here.
                        if ((codeClassName ?? "").includes("language-")) {
                            return (
                                <code
                                    className={cn("whitespace-pre", codeClassName)}
                                    {...rest}
                                >
                                    {children}
                                </code>
                            );
                        }
                        return (
                            <code
                                className="rounded bg-muted/80 px-1.5 py-0.5 font-mono text-[13px] text-amber-700 dark:text-amber-300 font-medium"
                                {...rest}
                            >
                                {children}
                            </code>
                        );
                    },
                }}
            >
                {content}
            </ReactMarkdown>
        </div>
    );
}

function ParametersTable({ parameters }: { parameters: AppTemplateParam[] }) {
    return (
        <div className="rounded-lg border border-border/60 overflow-hidden">
            <Table>
                <TableHeader className="bg-muted/50 text-xs">
                    <TableRow>
                        <TableHead className="font-semibold text-foreground">Parameter</TableHead>
                        <TableHead className="font-semibold text-foreground">Type</TableHead>
                        <TableHead className="font-semibold text-foreground">Default Value</TableHead>
                        <TableHead className="font-semibold text-foreground">Required</TableHead>
                        <TableHead className="font-semibold text-foreground">Description</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody className="text-xs">
                    {parameters.map(param => {
                        const defaultValueStr = formatParamValue(param.default);
                        const minStr = param.min !== undefined ? formatParamValue(param.min) : "";
                        return (
                            <TableRow
                                key={param.name}
                                className="hover:bg-muted/30"
                            >
                                {/* Name & Title */}
                                <TableCell className="font-mono font-medium text-foreground py-2.5">
                                    <div className="flex flex-col">
                                        <span>{param.name}</span>
                                        {param.title && param.title !== param.name && (
                                            <span className="font-sans text-[11px] text-muted-foreground">
                                                {param.title}
                                            </span>
                                        )}
                                    </div>
                                </TableCell>

                                {/* Type */}
                                <TableCell className="py-2.5">
                                    <Badge
                                        variant="outline"
                                        className={cn(
                                            "text-[11px] font-mono px-1.5 py-0.5",
                                            param.type === "secret"
                                                ? "bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/30"
                                                : param.type === "volume"
                                                  ? "bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/30"
                                                  : param.type === "size"
                                                    ? "bg-purple-500/10 text-purple-700 dark:text-purple-400 border-purple-500/30"
                                                    : param.type === "domain"
                                                      ? "bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border-emerald-500/30"
                                                      : param.type === "app"
                                                        ? "bg-cyan-500/10 text-cyan-700 dark:text-cyan-400 border-cyan-500/30"
                                                        : "bg-muted/60 text-muted-foreground",
                                        )}
                                    >
                                        {param.type === "secret" && <KeyRound className="size-2.5 mr-1 inline" />}
                                        {param.type === "domain" && <Globe className="size-2.5 mr-1 inline" />}
                                        {param.type === "volume" && <HardDrive className="size-2.5 mr-1 inline" />}
                                        {param.type === "app" && <Boxes className="size-2.5 mr-1 inline" />}
                                        {param.type}
                                    </Badge>
                                </TableCell>

                                {/* Default value */}
                                <TableCell className="py-2.5 font-mono text-muted-foreground">
                                    {param.generated ? (
                                        <span className="italic text-amber-700/80 dark:text-amber-400/80 text-[11px]">
                                            Auto-generated
                                        </span>
                                    ) : defaultValueStr !== "" ? (
                                        <span className="rounded bg-muted/60 px-1.5 py-0.5 text-[11px] text-foreground/90">
                                            {defaultValueStr}
                                        </span>
                                    ) : (
                                        <span className="text-muted-foreground/40">—</span>
                                    )}
                                </TableCell>

                                {/* Required / Optional */}
                                <TableCell className="py-2.5">
                                    {param.optional ? (
                                        <span className="text-muted-foreground text-[11px]">Optional</span>
                                    ) : (
                                        <span className="text-amber-700 dark:text-amber-400 font-semibold text-[11px]">
                                            Required
                                        </span>
                                    )}
                                </TableCell>

                                {/* Description & Constraints */}
                                <TableCell className="py-2.5 max-w-[260px] text-muted-foreground">
                                    <div className="space-y-0.5">
                                        {param.description && <p>{param.description}</p>}
                                        {param.pattern && (
                                            <p className="font-mono text-[10px] text-muted-foreground/70">
                                                Pattern: {param.pattern}
                                            </p>
                                        )}
                                        {minStr !== "" && (
                                            <p className="text-[10px] text-muted-foreground/70">Min: {minStr}</p>
                                        )}
                                    </div>
                                </TableCell>
                            </TableRow>
                        );
                    })}
                </TableBody>
            </Table>
        </div>
    );
}

export function AppTemplatesDetailsView({
    templateName,
    templateSummary,
    onBack,
    onSelectTag,
    className,
}: AppTemplatesDetailsViewProps) {
    const [imageError, setImageError] = useState(false);

    // Fetch full template detail
    const { data: templateDetail, isLoading } = useGetAppTemplate(templateName);

    // Current data: use detailed response when available, fallback to summary
    const template: AppTemplateDetail | AppTemplateSummary | undefined = templateDetail ?? templateSummary;

    // Versions
    const versions = template?.versions ?? [];
    const defaultVer = versions.find(v => v.default) ?? versions[0];
    const [selectedVersionName, setSelectedVersionName] = useState<string>(defaultVer?.name ?? "");

    useEffect(() => {
        setSelectedVersionName(defaultVer?.name ?? "");
    }, [templateName, defaultVer?.name]);

    // Variants (if detailed data is loaded)
    const variants = templateDetail?.variants ?? [];

    const primaryCategory = template ? (template.categories[0]?.split("/").pop() ?? template.categories[0]) : undefined;
    const { id: projectId } = useParams<{ id: string }>();
    const { open: openDeployDialog } = useDeployTemplateDialogState();

    const handleDeployClick = () => {
        if (!template || !template.compatible || !projectId) return;
        openDeployDialog(projectId, {
            templateName: template.name,
            initialVersion: selectedVersionName,
        });
    };

    if (!template && isLoading) {
        return (
            <div className={cn("flex flex-col gap-4 w-full", className)}>
                {/* Top Navigation & Actions Bar */}
                <div className="flex items-center justify-between gap-4 border-b border-border/50 pb-2.5">
                    <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={onBack}
                        className="h-8 px-2.5 text-xs font-medium text-muted-foreground hover:text-foreground hover:bg-muted"
                    >
                        <ArrowLeft className="size-4 mr-1.5" />
                        Back to templates
                    </Button>
                </div>

                <div className="rounded-xl border border-border/70 bg-card p-6 shadow-xs space-y-4">
                    <div className="flex items-center gap-4">
                        <Skeleton className="size-14 rounded-xl" />
                        <div className="space-y-2 flex-1">
                            <Skeleton className="h-6 w-1/3" />
                            <Skeleton className="h-4 w-1/2" />
                        </div>
                    </div>
                    <Skeleton className="h-32 w-full" />
                </div>
            </div>
        );
    }

    if (!template) {
        return (
            <div className={cn("flex flex-col gap-4 w-full", className)}>
                <div className="flex items-center justify-between gap-4 border-b border-border/50 pb-2.5">
                    <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={onBack}
                        className="h-8 px-2.5 text-xs font-medium text-muted-foreground hover:text-foreground hover:bg-muted"
                    >
                        <ArrowLeft className="size-4 mr-1.5" />
                        Back to templates
                    </Button>
                </div>

                <div className="rounded-xl border border-border/70 bg-card p-8 text-center shadow-xs">
                    <p className="text-sm font-semibold text-foreground">Template not found</p>
                    <p className="text-xs text-muted-foreground mt-1">
                        The requested template &quot;{templateName}&quot; could not be found or loaded.
                    </p>
                    <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={onBack}
                        className="mt-4"
                    >
                        Return to catalog
                    </Button>
                </div>
            </div>
        );
    }

    return (
        <div className={cn("flex flex-col gap-4 w-full", className)}>
            {/* Top Navigation & Actions Bar */}
            <div className="flex items-center justify-between gap-4 border-b border-border/50 pb-2.5">
                <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={onBack}
                    className="h-8 px-2.5 text-xs font-medium text-muted-foreground hover:text-foreground hover:bg-muted"
                >
                    <ArrowLeft className="size-4 mr-1.5" />
                    Back to templates
                </Button>

                <div className="flex items-center gap-2">
                    <span className="text-xs text-muted-foreground hidden sm:inline-block">Template:</span>
                    <span className="text-xs font-semibold text-foreground/90 bg-muted/60 px-2 py-0.5 rounded">
                        {template.name}
                    </span>
                </div>
            </div>

            {/* Template Header Card */}
            <div className="rounded-xl border border-border/70 bg-card p-4 sm:p-5 shadow-xs">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                    {/* Left: Icon, Title, Category & Tagline */}
                    <div className="flex items-start gap-4 min-w-0">
                        <div className="flex size-14 shrink-0 items-center justify-center rounded-xl border border-border/80 bg-muted/30 p-2.5 overflow-hidden shadow-2xs">
                            {!imageError && template.iconUrl ? (
                                <img
                                    src={template.iconUrl}
                                    alt={template.title}
                                    onError={() => {
                                        setImageError(true);
                                    }}
                                    className="size-full object-contain"
                                />
                            ) : (
                                <Boxes className="size-7 text-amber-600 dark:text-amber-400" />
                            )}
                        </div>

                        <div className="min-w-0">
                            <div className="flex items-center gap-2 flex-wrap">
                                <h1 className="text-2xl font-bold tracking-tight text-foreground">{template.title}</h1>
                                {primaryCategory && (
                                    <Badge
                                        variant="outline"
                                        className="text-[12px] font-medium uppercase tracking-wider px-2 py-0.5 bg-muted/50 border-border/70 text-muted-foreground"
                                    >
                                        {primaryCategory}
                                    </Badge>
                                )}
                                {template.compatible ? (
                                    <Badge
                                        variant="secondary"
                                        className="text-[11px] font-medium bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-500/20"
                                    >
                                        <CheckCircle2 className="size-3 mr-1 inline" />
                                        Compatible
                                    </Badge>
                                ) : (
                                    <Badge
                                        variant="destructive"
                                        className="text-[11px] font-medium"
                                    >
                                        <Lock className="size-3 mr-1 inline" />
                                        Requires Newer HivePaaS
                                    </Badge>
                                )}
                            </div>

                            <p className="mt-1 text-[15px] font-medium text-muted-foreground leading-relaxed">
                                {template.tagline}
                            </p>
                        </div>
                    </div>

                    {/* Right: Version Selector & Deploy Button */}
                    <div className="flex items-center gap-2.5 shrink-0 self-start sm:self-center">
                        {versions.length > 0 && (
                            <div className="flex items-center gap-1.5">
                                <Select
                                    value={selectedVersionName || defaultVer?.name}
                                    onValueChange={val => {
                                        setSelectedVersionName(val);
                                    }}
                                >
                                    <SelectTrigger className="h-9 w-auto min-w-[110px] px-3 text-xs font-medium bg-muted/40 border-border/70">
                                        <SelectValue placeholder="Version" />
                                    </SelectTrigger>
                                    <SelectContent align="end">
                                        {versions.map(v => (
                                            <SelectItem
                                                key={v.name}
                                                value={v.name}
                                                className="text-xs"
                                            >
                                                v{v.name} {v.release ? `(${v.release})` : ""}
                                                {v.default && " (latest)"}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                        )}

                        <Button
                            type="button"
                            size="default"
                            disabled={!template.compatible}
                            onClick={handleDeployClick}
                            className="h-9 px-4 text-sm font-semibold bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 text-white shadow-xs disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            <Rocket className="size-4 mr-1.5" />
                            Deploy Template
                        </Button>
                    </div>
                </div>
            </div>

            {/* SECTION 4 (BOTTOM): License & Legal Responsibility Disclaimer */}
            <div className="rounded-xl border border-amber-500/40 bg-amber-500/5 dark:bg-amber-500/10 p-4 shadow-xs">
                <div className="flex items-start gap-3.5">
                    <div className="rounded-lg bg-amber-500/15 p-2 text-amber-700 dark:text-amber-400 shrink-0">
                        <Scale className="size-5" />
                    </div>

                    <div className="flex-1 space-y-2 text-xs">
                        <div className="flex items-center gap-2 flex-wrap">
                            <h4 className="text-sm font-semibold text-foreground">
                                Important: License & Legal Responsibility
                            </h4>
                            {template.license && (
                                <Badge
                                    variant="secondary"
                                    className="bg-amber-500/20 text-amber-700 dark:text-amber-300 font-semibold text-xs px-2 py-0.5 border border-amber-500/30"
                                >
                                    {template.license}
                                </Badge>
                            )}
                        </div>

                        <p className="leading-relaxed text-foreground/90 font-medium">
                            When deploying and operating applications using this template, you are solely responsible
                            for ensuring compliance with all software license terms, copyright notices, and applicable
                            agreements (including permissions for commercial use, redistribution, and source disclosure
                            obligations under copyleft licenses).
                        </p>

                        <p className="leading-relaxed text-muted-foreground">
                            We strongly recommend reviewing the license terms and restrictions carefully before
                            deploying to production environments.
                            {template.license && (
                                <a
                                    href={`https://spdx.org/licenses/${encodeURIComponent(template.license)}.html`}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="inline-flex items-center text-amber-700 dark:text-amber-400 hover:underline font-medium ml-1.5"
                                >
                                    View {template.license} license details
                                    <ExternalLink className="size-3 ml-0.5" />
                                </a>
                            )}
                        </p>
                    </div>
                </div>
            </div>

            {/* What this template asks the host for. It sits above the license
                block because it is the one thing on this page that changes who
                may deploy the template at all. */}
            {!isLoading && templateDetail && <CapabilitiesSection template={templateDetail} />}

            {/* Main Content Layout: Left 3/4 and Right 1/4 */}
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-4 items-start">
                {/* Left Column: Sections (Description, Template Parameters, Versions Matrix) */}
                <div className="lg:col-span-9 space-y-4">
                    {/* SECTION 1: Description */}
                    <Card className="border-border/70 shadow-xs py-0 gap-0">
                        <div className="flex items-center px-4 sm:px-5 py-3 border-b border-border/40 min-h-[46px]">
                            <h3 className="text-sm font-semibold text-foreground flex items-center gap-2">
                                <Info className="size-4 text-amber-600 dark:text-amber-400" />
                                Description
                            </h3>
                        </div>
                        <CardContent className="px-4 sm:px-5 py-3.5">
                            {isLoading ? (
                                <div className="space-y-2">
                                    <Skeleton className="h-4 w-full" />
                                    <Skeleton className="h-4 w-5/6" />
                                    <Skeleton className="h-4 w-4/6" />
                                </div>
                            ) : templateDetail?.description ? (
                                <MarkdownRenderer content={templateDetail.description} />
                            ) : (
                                <p className="text-sm text-muted-foreground m-0">{template.tagline}</p>
                            )}
                        </CardContent>
                    </Card>

                    {/* SECTION 2: Template Parameters (Separate, equal-level section) */}
                    <Card className="border-border/70 shadow-xs py-0 gap-0">
                        <div className="flex items-center justify-between px-4 sm:px-5 py-3 border-b border-border/40 min-h-[46px]">
                            <h3 className="text-sm font-semibold text-foreground flex items-center gap-2">
                                <Sliders className="size-4 text-amber-600 dark:text-amber-400" />
                                Template Parameters
                            </h3>
                            {templateDetail?.parameters && (
                                <span className="text-xs text-muted-foreground">
                                    {templateDetail.parameters.length} configurable parameters
                                </span>
                            )}
                        </div>
                        <CardContent className="px-4 sm:px-5 py-3.5">
                            {isLoading ? (
                                <div className="space-y-3">
                                    <Skeleton className="h-8 w-full" />
                                    <Skeleton className="h-8 w-full" />
                                    <Skeleton className="h-8 w-full" />
                                </div>
                            ) : templateDetail?.parameters && templateDetail.parameters.length > 0 ? (
                                <ParametersTable parameters={templateDetail.parameters} />
                            ) : (
                                <p className="text-xs text-muted-foreground italic">
                                    No custom configuration parameters declared for this template.
                                </p>
                            )}
                        </CardContent>
                    </Card>

                    {/* SECTION 3: Dependent Services */}
                    {!isLoading && templateDetail?.dependencies && templateDetail.dependencies.length > 0 && (
                        <Card className="border-border/70 shadow-xs py-0 gap-0">
                            <div className="flex items-center justify-between px-4 sm:px-5 py-3 border-b border-border/40 min-h-[46px]">
                                <h3 className="text-sm font-semibold text-foreground flex items-center gap-2">
                                    <Boxes className="size-4 text-amber-600 dark:text-amber-400" />
                                    Dependent Services
                                </h3>
                                <span className="text-xs text-muted-foreground">
                                    {templateDetail.dependencies.length}{" "}
                                    {templateDetail.dependencies.length === 1
                                        ? "companion service"
                                        : "companion services"}
                                </span>
                            </div>
                            <CardContent className="px-4 sm:px-5 py-4 space-y-4">
                                <p className="text-xs text-muted-foreground leading-relaxed">
                                    This template automatically deploys companion services needed for operation.
                                </p>

                                <div className="space-y-4">
                                    {templateDetail.dependencies.map((dep: AppTemplateDependency) => (
                                        <div
                                            key={dep.name}
                                            className="rounded-xl border border-border/70 bg-muted/20 p-4 space-y-3.5"
                                        >
                                            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-border/40 pb-3">
                                                <div className="flex items-center gap-2.5">
                                                    <div className="flex size-8 shrink-0 items-center justify-center rounded-lg border border-border/70 bg-background p-1.5 shadow-2xs">
                                                        <HardDrive className="size-4 text-amber-500" />
                                                    </div>
                                                    <div>
                                                        <div className="flex items-center gap-2">
                                                            <span className="text-sm font-semibold text-foreground">
                                                                {dep.title ? dep.title : dep.name}
                                                            </span>
                                                            {projectId ? (
                                                                <AppLink.Basic
                                                                    to={ROUTE.projects.single.appTemplates.single.$route(
                                                                        projectId,
                                                                        dep.template,
                                                                    )}
                                                                    className="text-xs font-mono text-amber-600 dark:text-amber-400 hover:underline inline-flex items-center gap-1"
                                                                    title={`View ${dep.templateTitle ?? dep.template} template`}
                                                                >
                                                                    <span>({dep.template})</span>
                                                                    <ExternalLink className="size-3" />
                                                                </AppLink.Basic>
                                                            ) : (
                                                                <span className="text-xs font-mono text-muted-foreground">
                                                                    ({dep.template})
                                                                </span>
                                                            )}
                                                        </div>
                                                        {dep.templateTitle && dep.templateTitle !== dep.title && (
                                                            <span className="text-[11px] text-muted-foreground block">
                                                                Template: {dep.templateTitle}
                                                            </span>
                                                        )}
                                                    </div>
                                                </div>

                                                <div className="flex items-center gap-1.5">
                                                    <Badge
                                                        variant="outline"
                                                        className="text-[10px] font-mono px-2 py-0.5 bg-background"
                                                    >
                                                        {dep.version ? `v${dep.version}` : "default"}
                                                    </Badge>
                                                    {dep.variant && (
                                                        <Badge
                                                            variant="secondary"
                                                            className="text-[10px] px-2 py-0.5"
                                                        >
                                                            {dep.variant}
                                                        </Badge>
                                                    )}
                                                </div>
                                            </div>

                                            {dep.parameters && dep.parameters.length > 0 ? (
                                                <div className="space-y-2">
                                                    <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                                                        Service Parameters
                                                    </span>
                                                    <ParametersTable parameters={dep.parameters} />
                                                </div>
                                            ) : (
                                                <p className="text-xs text-muted-foreground italic">
                                                    No additional configuration parameters required for this service.
                                                </p>
                                            )}
                                        </div>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>
                    )}

                    {/* SECTION 4: Supported Versions & Variants */}
                    <Card className="border-border/70 shadow-xs py-0 gap-0">
                        <div className="flex items-center px-4 sm:px-5 py-3 border-b border-border/40 min-h-[46px]">
                            <h3 className="text-sm font-semibold text-foreground flex items-center gap-2">
                                <HardDrive className="size-4 text-amber-600 dark:text-amber-400" />
                                Supported Versions & Variants
                            </h3>
                        </div>
                        <CardContent className="px-4 sm:px-5 py-3.5 space-y-3.5">
                            {/* Versions List */}
                            <div className="space-y-2">
                                <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                    Available Versions
                                </span>
                                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2.5">
                                    {versions.map(v => {
                                        const isCurrent = v.name === (selectedVersionName || defaultVer?.name);
                                        return (
                                            <div
                                                key={v.name}
                                                className={cn(
                                                    "flex items-center justify-between rounded-lg border p-2.5 text-xs transition-colors",
                                                    isCurrent
                                                        ? "border-amber-500/50 bg-amber-500/5 text-foreground"
                                                        : "border-border/60 bg-muted/20 text-muted-foreground",
                                                )}
                                            >
                                                <div className="flex items-center gap-2">
                                                    <span className="font-semibold text-foreground">v{v.name}</span>
                                                    {v.release && (
                                                        <span className="text-[11px] text-muted-foreground">
                                                            ({v.release})
                                                        </span>
                                                    )}
                                                </div>
                                                <div className="flex items-center gap-1">
                                                    {v.default && (
                                                        <Badge
                                                            variant="secondary"
                                                            className="text-[10px] px-1.5 py-0 bg-amber-500/20 text-amber-700 dark:text-amber-300"
                                                        >
                                                            latest
                                                        </Badge>
                                                    )}
                                                    {v.deprecated && (
                                                        <Badge
                                                            variant="destructive"
                                                            className="text-[10px] px-1.5 py-0"
                                                        >
                                                            deprecated
                                                        </Badge>
                                                    )}
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            </div>

                            {/* Variants List (if available in detail) */}
                            {variants.length > 0 && (
                                <div className="space-y-2 pt-2 border-t border-border/40">
                                    <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                        Variants
                                    </span>
                                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
                                        {variants.map(variant => (
                                            <div
                                                key={variant.name}
                                                className="rounded-lg border border-border/60 bg-muted/20 p-2.5 text-xs space-y-1"
                                            >
                                                <div className="flex items-center justify-between">
                                                    <span className="font-semibold text-foreground capitalize">
                                                        {variant.title || variant.name}
                                                    </span>
                                                    {variant.default && (
                                                        <Badge
                                                            variant="secondary"
                                                            className="text-[10px] px-1 py-0"
                                                        >
                                                            default
                                                        </Badge>
                                                    )}
                                                </div>
                                                {variant.description && (
                                                    <p className="text-[11px] text-muted-foreground leading-relaxed">
                                                        {variant.description}
                                                    </p>
                                                )}
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </div>

                {/* Right Column: Side Panel (Metadata, Official Links, Tags) */}
                <div className="lg:col-span-3 space-y-4">
                    {/* Official Links */}
                    {templateDetail?.links && (
                        <Card className="border-border/70 shadow-xs py-0 gap-0">
                            <div className="flex items-center px-4 py-2.5 border-b border-border/40 min-h-[40px]">
                                <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                                    Official Links
                                </h3>
                            </div>
                            <CardContent className="px-4 py-2.5 space-y-1.5 text-xs">
                                {templateDetail.links.website && (
                                    <a
                                        href={templateDetail.links.website}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="flex items-center justify-between gap-2 p-2 rounded-lg hover:bg-muted/60 text-muted-foreground hover:text-foreground transition-colors group"
                                    >
                                        <div className="flex items-center gap-2 min-w-0">
                                            <Globe className="size-3.5 text-amber-600 dark:text-amber-400 shrink-0" />
                                            <span className="truncate">Official Website</span>
                                        </div>
                                        <ExternalLink className="size-3 text-muted-foreground group-hover:text-foreground shrink-0" />
                                    </a>
                                )}

                                {templateDetail.links.documentation && (
                                    <a
                                        href={templateDetail.links.documentation}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="flex items-center justify-between gap-2 p-2 rounded-lg hover:bg-muted/60 text-muted-foreground hover:text-foreground transition-colors group"
                                    >
                                        <div className="flex items-center gap-2 min-w-0">
                                            <BookOpen className="size-3.5 text-amber-600 dark:text-amber-400 shrink-0" />
                                            <span className="truncate">Documentation</span>
                                        </div>
                                        <ExternalLink className="size-3 text-muted-foreground group-hover:text-foreground shrink-0" />
                                    </a>
                                )}

                                {templateDetail.links.source && (
                                    <a
                                        href={templateDetail.links.source}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="flex items-center justify-between gap-2 p-2 rounded-lg hover:bg-muted/60 text-muted-foreground hover:text-foreground transition-colors group"
                                    >
                                        <div className="flex items-center gap-2 min-w-0">
                                            <Code2 className="size-3.5 text-amber-600 dark:text-amber-400 shrink-0" />
                                            <span className="truncate">Source Code</span>
                                        </div>
                                        <ExternalLink className="size-3 text-muted-foreground group-hover:text-foreground shrink-0" />
                                    </a>
                                )}
                            </CardContent>
                        </Card>
                    )}

                    {/* Template Metadata */}
                    <Card className="border-border/70 shadow-xs py-0 gap-0">
                        <div className="flex items-center px-4 py-2.5 border-b border-border/40 min-h-[40px]">
                            <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                                Template Metadata
                            </h3>
                        </div>
                        <CardContent className="px-4 py-2.5 space-y-2 text-xs">
                            <div className="flex items-center justify-between gap-2">
                                <span className="text-muted-foreground shrink-0">Identifier:</span>
                                <span
                                    className="font-mono font-medium text-foreground truncate"
                                    title={template.name}
                                >
                                    {template.name}
                                </span>
                            </div>

                            {templateDetail?.source && (
                                <div className="flex items-center justify-between gap-2">
                                    <span className="text-muted-foreground shrink-0">Catalog Source:</span>
                                    <span
                                        className="text-foreground truncate max-w-[120px]"
                                        title={templateDetail.source}
                                    >
                                        {templateDetail.source}
                                    </span>
                                </div>
                            )}

                            {templateDetail?.revision && (
                                <div className="flex items-center justify-between gap-2">
                                    <span className="text-muted-foreground shrink-0">Revision:</span>
                                    <span
                                        className="font-mono text-[11px] text-muted-foreground truncate max-w-[120px]"
                                        title={templateDetail.revision}
                                    >
                                        {templateDetail.revision.slice(0, 10)}
                                    </span>
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    {/* Tags List */}
                    {template.tags.length > 0 && (
                        <Card className="border-border/70 shadow-xs py-0 gap-0">
                            <div className="flex items-center px-4 py-2.5 border-b border-border/40 min-h-[40px]">
                                <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-1.5">
                                    <Tag className="size-3.5 text-muted-foreground" />
                                    Tags
                                </h3>
                            </div>
                            <CardContent className="px-4 py-2.5">
                                <div className="flex flex-wrap gap-1.5">
                                    {template.tags.map(rawTag => {
                                        const cleanTag = rawTag.replace(/^#/, "");
                                        return (
                                            <button
                                                key={rawTag}
                                                type="button"
                                                onClick={() => {
                                                    onSelectTag?.(cleanTag);
                                                }}
                                                className="inline-flex items-center rounded-md bg-muted/60 px-2 py-0.5 text-[12px] font-medium text-muted-foreground hover:bg-amber-500/15 hover:text-amber-700 dark:hover:text-amber-400 transition-colors cursor-pointer"
                                            >
                                                {cleanTag}
                                            </button>
                                        );
                                    })}
                                </div>
                            </CardContent>
                        </Card>
                    )}
                </div>
            </div>
        </div>
    );
}

/**
 * What deploying this template grants the apps it creates, and the permission
 * needed to grant it. It says nothing at all for the templates that ask for
 * nothing, which is nearly all of them.
 */
function CapabilitiesSection({ template }: { template: AppTemplateDetail }) {
    const granted = useMemo(() => grantedByTemplate(template), [template]);
    const { canWrite: canGrant } = useConditionalModule({ id: MODULE_IDS.Cluster });

    if (granted.length === 0) {
        return null;
    }
    return (
        <div className="rounded-xl border border-orange-500/40 bg-orange-500/5 dark:bg-orange-500/10 p-4 shadow-xs">
            <div className="flex items-start gap-3.5">
                <div className="rounded-lg bg-orange-500/15 p-2 text-orange-700 dark:text-orange-400 shrink-0">
                    <ShieldAlert className="size-5" />
                </div>

                <div className="flex-1 space-y-2 text-xs">
                    <h4 className="text-sm font-semibold text-foreground">Elevated container privileges</h4>

                    {granted.map(one => (
                        <div
                            key={one.app}
                            className="flex flex-wrap items-center gap-1.5"
                        >
                            <span className="text-muted-foreground">{one.app} gets</span>
                            {describeCapabilities(one.capabilities).map(item => (
                                <Badge
                                    key={item}
                                    variant="outline"
                                    className="h-5 px-1.5 font-mono text-[10px] border-orange-500/40 text-orange-700 dark:text-orange-400"
                                >
                                    {item}
                                </Badge>
                            ))}
                        </div>
                    ))}

                    <p className="leading-relaxed text-foreground/90 font-medium">
                        Capabilities give a container access the host would otherwise keep from it, and can introduce
                        severe security risks. Deploy this only if you understand what it is being given.
                    </p>

                    <p className="leading-relaxed text-muted-foreground">
                        {canGrant ? (
                            <>
                                Granting this needs <span className="font-medium text-orange-500">Write</span>{" "}
                                permission on the <span className="font-medium text-orange-500">Cluster</span> module,
                                which you have.
                            </>
                        ) : (
                            <span className="font-medium text-destructive">
                                You need Write permission on the Cluster module to deploy this template.
                            </span>
                        )}
                    </p>
                </div>
            </div>
        </div>
    );
}

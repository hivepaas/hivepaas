import React, { useState } from "react";

import { cn } from "@/lib/utils";
import { Boxes, Lock, Rocket, Scale } from "lucide-react";
import { useParams } from "react-router";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import type { AppTemplateSummary, AppTemplateVersionSummary } from "../api";
import { useDeployTemplateDialogState } from "../dialogs";

interface AppTemplateCardProps {
    template: AppTemplateSummary;
    onSelect: (template: AppTemplateSummary) => void;
    onSelectTag?: (tag: string) => void;
    selectedTag?: string;
    className?: string;
}

export function AppTemplateCard({ template, onSelect, onSelectTag, selectedTag, className }: AppTemplateCardProps) {
    const [imageError, setImageError] = useState(false);
    const { id: projectId } = useParams<{ id: string }>();
    const { open: openDeployDialog } = useDeployTemplateDialogState();

    // Find default version or first available version
    const defaultVer = template.versions.find(v => v.default) ?? template.versions[0];
    const [selectedVersionName, setSelectedVersionName] = useState<string>(defaultVer?.name ?? "");

    const selectedVersion: AppTemplateVersionSummary | undefined =
        template.versions.find(v => v.name === selectedVersionName) ?? defaultVer;

    const primaryCategory = template.categories[0]?.split("/").pop() ?? template.categories[0];

    const handleDeployClick = (e: React.MouseEvent) => {
        e.stopPropagation();
        if (!template.compatible || !projectId) return;
        openDeployDialog(projectId, {
            templateName: template.name,
            initialVersion: selectedVersionName,
        });
    };

    return (
        <div
            role="button"
            tabIndex={0}
            onClick={() => {
                onSelect(template);
            }}
            onKeyDown={e => {
                if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    onSelect(template);
                }
            }}
            className={cn(
                "group relative flex flex-col justify-between rounded-xl border border-border/70 bg-card p-4 text-card-foreground shadow-xs transition-all duration-200",
                "hover:border-amber-500/50 hover:shadow-md hover:bg-card/90 cursor-pointer text-left",
                !template.compatible && "opacity-80 border-dashed",
                className,
            )}
        >
            <div>
                {/* Top Row: Icon, Title & Primary Category */}
                <div className="flex items-start gap-3.5">
                    {/* Icon with fallback */}
                    <div className="flex size-12 shrink-0 items-center justify-center rounded-lg border border-border/80 bg-muted/40 p-2 overflow-hidden shadow-2xs group-hover:border-amber-500/30 group-hover:bg-amber-500/5 transition-colors">
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
                            <Boxes className="size-6 text-amber-600 dark:text-amber-400" />
                        )}
                    </div>

                    <div className="flex-1 min-w-0">
                        <div className="flex items-center justify-between gap-1.5">
                            <h3 className="truncate font-semibold text-base tracking-tight text-foreground group-hover:text-amber-600 dark:group-hover:text-amber-400 transition-colors">
                                {template.title}
                            </h3>
                            {primaryCategory && (
                                <Badge
                                    variant="outline"
                                    className="shrink-0 text-[12px] font-medium uppercase tracking-wider px-2 py-0.5 bg-muted/50 border-border/60 text-muted-foreground"
                                >
                                    {primaryCategory}
                                </Badge>
                            )}
                        </div>

                        {/* Tagline */}
                        <p className="mt-1 line-clamp-2 text-[14px] leading-relaxed text-muted-foreground">
                            {template.tagline}
                        </p>
                    </div>
                </div>

                {/* Compatibility Warning if not compatible */}
                {!template.compatible && (
                    <div className="mt-3 flex items-center gap-1.5 rounded-md bg-amber-500/10 px-2 py-1 text-[12px] text-amber-700 dark:text-amber-400 border border-amber-500/20">
                        <Lock className="size-3.5 shrink-0" />
                        <span className="truncate">Requires Newer HivePaaS</span>
                    </div>
                )}

                {/* Tags */}
                {template.tags.length > 0 && (
                    <div className="mt-3.5 flex flex-wrap gap-1">
                        {template.tags.slice(0, 4).map(rawTag => {
                            const cleanTag = rawTag.replace(/^#/, "");
                            const isTagSelected = selectedTag === cleanTag;
                            return (
                                <button
                                    key={rawTag}
                                    type="button"
                                    onClick={e => {
                                        e.stopPropagation();
                                        onSelectTag?.(cleanTag);
                                    }}
                                    className={cn(
                                        "inline-flex items-center rounded-md px-2 py-0.5 text-[12px] font-medium transition-colors cursor-pointer",
                                        isTagSelected
                                            ? "bg-amber-500/20 text-amber-700 dark:text-amber-300 font-semibold border border-amber-500/40"
                                            : "bg-muted/60 text-muted-foreground hover:bg-amber-500/15 hover:text-amber-700 dark:hover:text-amber-400",
                                    )}
                                >
                                    {cleanTag}
                                </button>
                            );
                        })}
                        {template.tags.length > 4 && (
                            <span className="inline-flex items-center rounded-md bg-muted/40 px-1.5 py-0.5 text-[11px] text-muted-foreground">
                                +{template.tags.length - 4}
                            </span>
                        )}
                    </div>
                )}
            </div>

            {/* Bottom Meta & Action Buttons */}
            <div className="mt-4 pt-3 border-t border-border/50 flex flex-col gap-2.5">
                {/* Versions & Compatibility indicator */}
                <div className="flex items-center justify-between gap-2 text-[12px] text-muted-foreground">
                    <div className="flex items-center gap-1.5">
                        <span className="text-[12px] text-muted-foreground/80 font-medium">Ver:</span>
                        {template.versions.length > 1 ? (
                            <Select
                                value={selectedVersionName}
                                onValueChange={val => {
                                    setSelectedVersionName(val);
                                }}
                            >
                                <SelectTrigger
                                    className="h-6.5 w-auto min-w-[70px] px-2 py-0 text-[12px] bg-muted/50 border-border/60 hover:bg-muted font-medium"
                                    onClick={e => {
                                        e.stopPropagation();
                                    }}
                                >
                                    <SelectValue placeholder="Version" />
                                </SelectTrigger>
                                <SelectContent align="start">
                                    {template.versions.map(v => (
                                        <SelectItem
                                            key={v.name}
                                            value={v.name}
                                            className="text-[12px]"
                                        >
                                            {v.name} {v.release ? `(${v.release})` : ""}
                                            {v.default && " (latest)"}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        ) : (
                            <Badge
                                variant="secondary"
                                className="px-1.5 py-0 text-[12px] font-normal"
                            >
                                {selectedVersion ? `${selectedVersion.name} (${selectedVersion.release})` : "latest"}
                            </Badge>
                        )}
                    </div>

                    {template.license ? (
                        <div
                            className="inline-flex items-center gap-1.5 rounded-md bg-muted/60 px-2 py-0.5 text-[12px] font-medium text-muted-foreground"
                            title={`License: ${template.license}`}
                        >
                            <Scale className="size-3 text-muted-foreground/80" />
                            <span className="truncate max-w-[110px]">{template.license}</span>
                        </div>
                    ) : (
                        <div
                            className="inline-flex items-center gap-1.5 rounded-md bg-muted/40 px-2 py-0.5 text-[12px] font-medium text-muted-foreground/60"
                            title="License not specified"
                        >
                            <Scale className="size-3 text-muted-foreground/60" />
                            <span>—</span>
                        </div>
                    )}
                </div>

                {/* Actions */}
                <div className="flex items-center gap-2">
                    <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={e => {
                            e.stopPropagation();
                            onSelect(template);
                        }}
                        className="flex-1 h-8.5 text-[14px] font-medium border-border/80 hover:bg-muted/70"
                    >
                        Details
                    </Button>
                    <Button
                        type="button"
                        size="sm"
                        disabled={!template.compatible}
                        onClick={handleDeployClick}
                        className="flex-1 h-8.5 text-[14px] font-medium bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 text-white shadow-xs disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        <Rocket className="size-3.5 mr-1" />
                        Deploy
                    </Button>
                </div>
            </div>
        </div>
    );
}

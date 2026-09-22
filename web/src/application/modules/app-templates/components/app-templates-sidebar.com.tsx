import React, { useEffect, useState } from "react";

import { cn } from "@/lib/utils";
import {
    Activity,
    BarChart3,
    ChevronDown,
    ChevronRight,
    Cpu,
    Database,
    ExternalLink,
    Globe,
    HardDrive,
    Layers,
    LayoutGrid,
    type LucideIcon,
    Server,
    Shield,
    Terminal,
    Workflow,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";

import type { AppTemplateCategory } from "../api";

interface AppTemplatesSidebarProps {
    categories: AppTemplateCategory[];
    selectedCategory?: string;
    onSelectCategory: (categoryId?: string) => void;
    totalCount?: number;
    className?: string;
}

const CATEGORY_ICON_MAP: Record<string, LucideIcon> = {
    "databases": Database,
    "message-queues": Workflow,
    "queues": Workflow,
    "web": Globe,
    "web-servers": Server,
    "analytics": BarChart3,
    "storage": HardDrive,
    "dev-tools": Terminal,
    "ai": Cpu,
    "ai-ml": Cpu,
    "monitoring": Activity,
    "security": Shield,
};

function getCategoryIcon(id: string): LucideIcon {
    const key = id.toLowerCase().split("/")[0] ?? "";
    return CATEGORY_ICON_MAP[key] ?? Layers;
}

export function AppTemplatesSidebar({
    categories,
    selectedCategory,
    onSelectCategory,
    totalCount,
    className,
}: AppTemplatesSidebarProps) {
    // Single expanded category ID: at most one category expanded at any time
    const [expandedCategoryId, setExpandedCategoryId] = useState<string | undefined>(() => {
        return selectedCategory ? selectedCategory.split("/")[0] : undefined;
    });

    // Auto-sync expanded category when selectedCategory changes externally
    useEffect(() => {
        if (selectedCategory) {
            const parentId = selectedCategory.split("/")[0];
            setExpandedCategoryId(parentId);
        }
    }, [selectedCategory]);

    const handleCategoryClick = (categoryId: string, hasChildren: boolean) => {
        onSelectCategory(categoryId);
        if (hasChildren) {
            // Expand this category immediately
            setExpandedCategoryId(categoryId);
        } else {
            setExpandedCategoryId(undefined);
        }
    };

    const handleToggleExpand = (e: React.MouseEvent, categoryId: string) => {
        e.stopPropagation();
        setExpandedCategoryId(prev => (prev === categoryId ? undefined : categoryId));
    };

    const isAllSelected = !selectedCategory;

    // Calculate total count by summing categories if BE doesn't provide a total count
    const calculatedCategoriesCount = React.useMemo(() => {
        if (categories.length === 0) return 0;
        return categories.reduce((sum, cat) => {
            const catCount = cat.count > 0 ? cat.count : (cat.children?.reduce((cSum, c) => cSum + c.count, 0) ?? 0);
            return sum + catCount;
        }, 0);
    }, [categories]);

    const effectiveTotalCount =
        totalCount !== undefined && totalCount > 0
            ? totalCount
            : calculatedCategoriesCount > 0
              ? calculatedCategoriesCount
              : undefined;

    return (
        <aside
            className={cn(
                "w-full md:w-[240px] shrink-0 flex flex-col gap-1 rounded-xl border border-border/60 bg-card p-3 shadow-xs",
                "md:sticky md:top-[56px] md:max-h-[calc(100vh-70px)] md:overflow-y-auto z-10",
                className,
            )}
        >
            <div className="px-2.5 py-2">
                <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground/80">
                    Categories
                </span>
            </div>

            {/* All Templates button */}
            <button
                type="button"
                onClick={() => {
                    onSelectCategory(undefined);
                    setExpandedCategoryId(undefined);
                }}
                className={cn(
                    "flex w-full items-center gap-2.5 rounded-lg px-2.5 py-1.5 text-sm font-medium transition-colors text-left",
                    isAllSelected
                        ? "bg-amber-500/15 text-amber-700 dark:text-amber-400 font-semibold"
                        : "text-muted-foreground hover:bg-muted/70 hover:text-foreground",
                )}
            >
                <LayoutGrid
                    className={cn(
                        "size-4 shrink-0",
                        isAllSelected ? "text-amber-600 dark:text-amber-400" : "text-muted-foreground",
                    )}
                />
                <span className="truncate">All Templates</span>
                {effectiveTotalCount !== undefined && effectiveTotalCount > 0 && (
                    <Badge
                        variant="outline"
                        className={cn(
                            "px-1.5 py-0 text-[11px] font-semibold tabular-nums shrink-0 h-4.5 rounded-md border",
                            isAllSelected
                                ? "bg-amber-500/20 text-amber-800 dark:text-amber-300 border-amber-500/40"
                                : "bg-muted/70 text-muted-foreground border-border/60",
                        )}
                    >
                        {effectiveTotalCount}
                    </Badge>
                )}
            </button>

            <div className="my-1 border-t border-border/40" />

            {/* Category tree */}
            <div className="flex flex-col gap-0.5">
                {categories.map(category => {
                    const hasChildren = Boolean(category.children && category.children.length > 0);
                    const isSelected = selectedCategory === category.id;
                    const Icon = getCategoryIcon(category.id);

                    // Check if any child is selected
                    const isChildSelected = Boolean(
                        category.children?.some(
                            c => selectedCategory === `${category.id}/${c.id}` || selectedCategory === c.id,
                        ),
                    );
                    const isExpanded = expandedCategoryId === category.id;

                    return (
                        <div
                            key={category.id}
                            className="flex flex-col"
                        >
                            <div
                                className={cn(
                                    "group flex items-center justify-between rounded-lg px-2.5 py-1.5 text-sm font-medium transition-colors",
                                    isSelected
                                        ? "bg-amber-500/15 text-amber-700 dark:text-amber-400 font-semibold"
                                        : isChildSelected
                                          ? "text-amber-600 dark:text-amber-400 bg-amber-500/5 font-medium"
                                          : "text-muted-foreground hover:bg-muted/70 hover:text-foreground",
                                )}
                            >
                                <button
                                    type="button"
                                    onClick={() => {
                                        handleCategoryClick(category.id, hasChildren);
                                    }}
                                    className="flex flex-1 items-center gap-2 min-w-0 text-left"
                                >
                                    <Icon
                                        className={cn(
                                            "size-4 shrink-0",
                                            isSelected || isChildSelected
                                                ? "text-amber-600 dark:text-amber-400"
                                                : "text-muted-foreground group-hover:text-foreground",
                                        )}
                                    />
                                    <span className="truncate">{category.title}</span>
                                    <Badge
                                        variant="outline"
                                        className={cn(
                                            "px-1.5 py-0 text-[11px] font-semibold tabular-nums shrink-0 h-4.5 rounded-md border",
                                            isSelected
                                                ? "bg-amber-500/20 text-amber-800 dark:text-amber-300 border-amber-500/40"
                                                : isChildSelected
                                                  ? "bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30"
                                                  : "bg-muted/70 text-muted-foreground border-border/60",
                                            category.count === 0 && "opacity-50",
                                        )}
                                    >
                                        {category.count}
                                    </Badge>
                                </button>

                                {hasChildren && (
                                    <button
                                        type="button"
                                        onClick={e => {
                                            handleToggleExpand(e, category.id);
                                        }}
                                        className="p-1 ml-1 text-muted-foreground hover:text-foreground rounded-md shrink-0 transition-colors"
                                        aria-label={isExpanded ? "Collapse category" : "Expand category"}
                                    >
                                        {isExpanded ? (
                                            <ChevronDown className="size-3.5" />
                                        ) : (
                                            <ChevronRight className="size-3.5" />
                                        )}
                                    </button>
                                )}
                            </div>

                            {/* Subcategories */}
                            {hasChildren && isExpanded && (
                                <div className="ml-5 mt-0.5 flex flex-col gap-0.5 border-l border-border/50 pl-2">
                                    {category.children?.map(child => {
                                        const fullSubId = `${category.id}/${child.id}`;
                                        const isSubSelected =
                                            selectedCategory === fullSubId || selectedCategory === child.id;
                                        return (
                                            <button
                                                key={child.id}
                                                type="button"
                                                onClick={() => {
                                                    onSelectCategory(fullSubId);
                                                }}
                                                className={cn(
                                                    "flex w-full items-center gap-2 rounded-md px-2 py-1 text-xs font-medium transition-colors text-left",
                                                    isSubSelected
                                                        ? "bg-amber-500/15 text-amber-700 dark:text-amber-400 font-semibold"
                                                        : "text-muted-foreground hover:bg-muted/70 hover:text-foreground",
                                                )}
                                            >
                                                <span className="truncate">{child.title}</span>
                                                <Badge
                                                    variant="outline"
                                                    className={cn(
                                                        "px-1.5 py-0 text-[10px] font-semibold tabular-nums shrink-0 h-4 rounded-md border",
                                                        isSubSelected
                                                            ? "bg-amber-500/20 text-amber-800 dark:text-amber-300 border-amber-500/40"
                                                            : "bg-muted/70 text-muted-foreground border-border/60",
                                                        child.count === 0 && "opacity-50",
                                                    )}
                                                >
                                                    {child.count}
                                                </Badge>
                                            </button>
                                        );
                                    })}
                                </div>
                            )}
                        </div>
                    );
                })}
            </div>

            <div className="my-1 border-t border-border/40" />

            <a
                href="https://github.com/hivepaas/app-templates/issues"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 rounded-lg px-2.5 py-1.5 text-xs font-medium text-link hover:bg-muted/70 hover:underline transition-colors group"
                title="Report issue on GitHub (opens in new tab)"
            >
                <GitHubIcon className="size-3.5 shrink-0" />
                <span className="truncate">Report issue on GitHub</span>
                <ExternalLink className="size-3 shrink-0 ml-auto opacity-70 group-hover:opacity-100" />
            </a>
        </aside>
    );
}

function GitHubIcon({ className }: { className?: string }) {
    return (
        <svg
            role="img"
            viewBox="0 0 24 24"
            fill="currentColor"
            xmlns="http://www.w3.org/2000/svg"
            className={className}
        >
            <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
        </svg>
    );
}

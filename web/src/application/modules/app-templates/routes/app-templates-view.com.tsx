import React, { useEffect, useMemo, useState } from "react";

import { ROUTE } from "@/application/shared/constants";
import { Loader2, Search, Tag, X } from "lucide-react";
import { useNavigate, useParams } from "react-router";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";

import type { AppTemplateSummary } from "../api";
import { AppTemplateCard, AppTemplatesDetailsView, AppTemplatesSidebar } from "../components";
import { useGetAppTemplateCatalog, useListAppTemplatesInfinite } from "../data";

export function AppTemplatesView() {
    const { id, templateName } = useParams<{ id: string; templateName?: string }>();
    const navigate = useNavigate();

    // 1. Fetch catalog metadata
    const { data: catalog } = useGetAppTemplateCatalog();

    // Active filters
    const [selectedCategory, setSelectedCategory] = useState<string | undefined>(undefined);
    const [searchQuery, setSearchQuery] = useState<string>("");
    const [submittedSearch, setSubmittedSearch] = useState<string>("");
    const [selectedTag, setSelectedTag] = useState<string | undefined>(undefined);

    // Cached template summary when clicked from card for instant rendering
    const [selectedTemplateSummary, setSelectedTemplateSummary] = useState<AppTemplateSummary | undefined>(undefined);

    // Handlers for search submit (Enter) and clear
    const handleSearchSubmit = () => {
        setSubmittedSearch(searchQuery.trim());
    };

    const handleSearchKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === "Enter") {
            e.preventDefault();
            handleSearchSubmit();
        }
    };

    const handleClearSearch = () => {
        setSearchQuery("");
        setSubmittedSearch("");
    };

    // 2. Fetch templates with infinite pagination (50 items per page)
    const {
        data: templatesData,
        isLoading: isTemplatesLoading,
        isFetchingNextPage,
        hasNextPage,
        fetchNextPage,
    } = useListAppTemplatesInfinite({
        category: selectedCategory,
        search: submittedSearch,
        tag: selectedTag,
    });

    // Flatten all pages into a single template list
    const templates = useMemo(() => {
        if (!templatesData) return [];
        return templatesData.pages.flatMap(page => page.data);
    }, [templatesData]);

    const totalTemplates = templatesData?.pages[0]?.meta.total ?? 0;

    // Cache the total count of all templates from the first unfiltered query
    const [allTemplatesTotal, setAllTemplatesTotal] = useState<number | undefined>(undefined);

    useEffect(() => {
        if (
            !selectedCategory &&
            !submittedSearch &&
            !selectedTag &&
            templatesData?.pages[0]?.meta.total !== undefined &&
            templatesData.pages[0].meta.total > 0
        ) {
            setAllTemplatesTotal(templatesData.pages[0].meta.total);
        }
    }, [selectedCategory, submittedSearch, selectedTag, templatesData]);

    // Calculate sum of categories from catalog as fallback when BE does not return a total
    const categoriesSum = useMemo(() => {
        if (!catalog || catalog.categories.length === 0) return undefined;
        const sum = catalog.categories.reduce((acc, cat) => {
            const catCount = cat.count > 0 ? cat.count : (cat.children?.reduce((cSum, c) => cSum + c.count, 0) ?? 0);
            return acc + catCount;
        }, 0);
        return sum > 0 ? sum : undefined;
    }, [catalog]);

    const allTemplatesCount =
        allTemplatesTotal && allTemplatesTotal > 0
            ? allTemplatesTotal
            : !selectedCategory && !submittedSearch && !selectedTag && totalTemplates > 0
              ? totalTemplates
              : categoriesSum;

    // Resolve template summary if already loaded in templates list
    const currentTemplateSummary = useMemo(() => {
        if (!templateName) return undefined;
        return selectedTemplateSummary ?? templates.find(t => t.name === templateName);
    }, [templateName, selectedTemplateSummary, templates]);

    // Navigation handlers
    const handleSelectTemplate = (template: AppTemplateSummary) => {
        setSelectedTemplateSummary(template);
        if (id) {
            void navigate(ROUTE.projects.single.appTemplates.single.$route(id, template.name));
        }
    };

    const handleBackToCatalog = () => {
        setSelectedTemplateSummary(undefined);
        if (id) {
            void navigate(ROUTE.projects.single.appTemplates.$route(id));
        }
    };

    const handleSelectTagFromDetails = (tag: string) => {
        setSelectedTag(tag);
        setSelectedCategory(undefined);
        if (id) {
            void navigate(ROUTE.projects.single.appTemplates.$route(id));
        }
    };

    const handleTagSelect = (tag: string) => {
        if (selectedTag === tag) {
            // Toggling off the active tag
            setSelectedTag(undefined);
        } else {
            // Selecting tag: Option 1 -> Reset category so BE API filters globally by tag
            setSelectedTag(tag);
            setSelectedCategory(undefined);
        }
    };

    // Handler when changing category from sidebar
    const handleCategorySelect = (categoryId?: string) => {
        setSelectedCategory(categoryId);
        setSelectedTag(undefined);
        if (id && templateName) {
            void navigate(ROUTE.projects.single.appTemplates.$route(id));
        }
    };

    return (
        <div className="w-full max-w-[1600px] mx-auto space-y-4">
            {/* Top Overview Banner - Sticky on scroll */}
            <div className="sticky top-[49px] md:top-0 z-20 bg-canvas/95 backdrop-blur-md py-3 border-b border-border/60">
                <p className="text-[16px] font-semibold text-foreground tracking-tight">
                    Discover, configure, and launch production-ready applications in seconds.
                </p>
            </div>

            {/* Main Layout: Left Sidebar + Right Content Area */}
            <div className="flex flex-col md:flex-row gap-6 items-start">
                {/* Left Side Menu - Always kept visible */}
                <AppTemplatesSidebar
                    categories={catalog?.categories ?? []}
                    selectedCategory={selectedCategory}
                    onSelectCategory={handleCategorySelect}
                    totalCount={allTemplatesCount}
                />

                {/* Right Content Area */}
                <div className="flex-1 min-w-0 w-full">
                    {templateName ? (
                        /* Details View: Takes over the right content area, left side menu preserved */
                        <AppTemplatesDetailsView
                            templateName={templateName}
                            templateSummary={currentTemplateSummary}
                            onBack={handleBackToCatalog}
                            onSelectTag={handleSelectTagFromDetails}
                        />
                    ) : (
                        /* Catalog List View */
                        <div className="flex flex-col gap-4">
                            {/* Sticky Toolbar: Search input + subcategories (NO Custom Template button, NO Sort By) */}
                            <div className="sticky top-[95px] md:top-[56px] z-10 bg-canvas/95 backdrop-blur-md py-2 border-b border-border/40 space-y-2">
                                <div className="flex items-center gap-3">
                                    <div className="relative flex-1">
                                        <button
                                            type="button"
                                            onClick={handleSearchSubmit}
                                            className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground cursor-pointer transition-colors"
                                            title="Search (Enter)"
                                            aria-label="Submit search"
                                        >
                                            <Search className="size-4" />
                                        </button>
                                        <Input
                                            value={searchQuery}
                                            onChange={e => {
                                                setSearchQuery(e.target.value);
                                            }}
                                            onKeyDown={handleSearchKeyDown}
                                            placeholder="Search templates by name, tagline, or tags (press Enter)..."
                                            className="pl-9 pr-8 h-9 text-sm bg-card border-border/70 focus-visible:ring-amber-500/30"
                                        />
                                        {searchQuery && (
                                            <button
                                                type="button"
                                                onClick={handleClearSearch}
                                                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground cursor-pointer transition-colors"
                                                title="Clear search"
                                                aria-label="Clear search"
                                            >
                                                <X className="size-3.5" />
                                            </button>
                                        )}
                                    </div>
                                </div>

                                {/* Active Tag Filter pill under search input */}
                                {selectedTag && (
                                    <div className="flex items-center gap-2 pt-0.5">
                                        <span className="text-[13px] text-muted-foreground font-medium">
                                            Filtered by tag:
                                        </span>
                                        <div className="inline-flex items-center gap-1.5 rounded-md bg-amber-500/15 border border-amber-500/30 px-2.5 py-0.5 text-[13px] font-medium text-amber-700 dark:text-amber-300">
                                            <Tag className="size-3 text-amber-600 dark:text-amber-400" />
                                            <span>{selectedTag}</span>
                                            <button
                                                type="button"
                                                onClick={() => {
                                                    setSelectedTag(undefined);
                                                }}
                                                className="ml-0.5 rounded p-0.5 hover:bg-amber-500/20 text-amber-700 dark:text-amber-300 transition-colors cursor-pointer"
                                                title="Clear tag"
                                            >
                                                <X className="size-3" />
                                                <span className="sr-only">Clear tag</span>
                                            </button>
                                        </div>
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="sm"
                                            onClick={() => {
                                                setSelectedTag(undefined);
                                            }}
                                            className="h-6 px-2 text-[12px] text-muted-foreground hover:text-foreground hover:bg-muted/60"
                                        >
                                            Clear
                                        </Button>
                                    </div>
                                )}
                            </div>

                            {/* Cards Grid / Loading / Empty States */}
                            {isTemplatesLoading ? (
                                <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-4">
                                    {Array.from({ length: 6 }).map((_, i) => (
                                        <div
                                            key={`sk-${i}`}
                                            className="rounded-xl border border-border/60 bg-card p-4 space-y-3"
                                        >
                                            <div className="flex items-start gap-3">
                                                <Skeleton className="size-12 rounded-lg" />
                                                <div className="flex-1 space-y-2">
                                                    <Skeleton className="h-4 w-3/4" />
                                                    <Skeleton className="h-3 w-full" />
                                                </div>
                                            </div>
                                            <Skeleton className="h-4 w-1/2" />
                                            <div className="pt-3 border-t border-border/40 flex justify-between gap-2">
                                                <Skeleton className="h-8 flex-1 rounded-md" />
                                                <Skeleton className="h-8 flex-1 rounded-md" />
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            ) : templates.length === 0 ? (
                                <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border/70 bg-card/50 p-12 text-center min-h-[300px]">
                                    <Search className="size-10 text-muted-foreground/50 mb-3" />
                                    <h3 className="text-base font-semibold text-foreground">No templates found</h3>
                                    <p className="text-xs text-muted-foreground mt-1 max-w-sm">
                                        We couldn&apos;t find any templates matching your search or category filter. Try
                                        clearing filters.
                                    </p>
                                    {(Boolean(submittedSearch) ||
                                        Boolean(searchQuery) ||
                                        Boolean(selectedCategory) ||
                                        Boolean(selectedTag)) && (
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            onClick={() => {
                                                setSearchQuery("");
                                                setSubmittedSearch("");
                                                setSelectedCategory(undefined);
                                                setSelectedTag(undefined);
                                            }}
                                            className="mt-4 text-xs"
                                        >
                                            Reset Filters
                                        </Button>
                                    )}
                                </div>
                            ) : (
                                <>
                                    {/* Template Cards Grid */}
                                    <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-4">
                                        {templates.map(template => (
                                            <AppTemplateCard
                                                key={template.name}
                                                template={template}
                                                onSelect={handleSelectTemplate}
                                                onSelectTag={handleTagSelect}
                                                selectedTag={selectedTag}
                                            />
                                        ))}
                                    </div>

                                    {/* Footer / Pagination */}
                                    <div className="flex flex-col sm:flex-row items-center justify-between gap-3 pt-4 border-t border-border/50 text-xs text-muted-foreground">
                                        <span>
                                            Showing {templates.length} of {totalTemplates} templates
                                        </span>

                                        {hasNextPage && (
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => {
                                                    void fetchNextPage();
                                                }}
                                                disabled={isFetchingNextPage}
                                                className="text-xs px-4 border-border/80 hover:bg-muted/70 min-w-[140px]"
                                            >
                                                {isFetchingNextPage ? (
                                                    <>
                                                        <Loader2 className="size-3.5 mr-2 animate-spin" />
                                                        Loading more...
                                                    </>
                                                ) : (
                                                    "Load more templates"
                                                )}
                                            </Button>
                                        )}
                                    </div>
                                </>
                            )}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

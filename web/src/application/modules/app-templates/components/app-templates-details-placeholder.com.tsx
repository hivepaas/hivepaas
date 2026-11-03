import React from "react";

import { ArrowLeft } from "lucide-react";

import { Button } from "@/components/ui/button";

import type { AppTemplateSummary } from "../api";

interface AppTemplatesDetailsPlaceholderProps {
    template: AppTemplateSummary;
    onBack: () => void;
}

export function AppTemplatesDetailsPlaceholder({ template, onBack }: AppTemplatesDetailsPlaceholderProps) {
    return (
        <div className="flex flex-1 flex-col rounded-xl border border-border/60 bg-card p-6 min-h-[600px] shadow-xs">
            {/* Navigation Header */}
            <div className="flex items-center justify-between border-b border-border/40 pb-4 mb-6">
                <div className="flex items-center gap-3">
                    <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={onBack}
                        className="h-8 px-2 text-xs font-medium text-muted-foreground hover:text-foreground hover:bg-muted"
                    >
                        <ArrowLeft className="size-4 mr-1.5" />
                        Back to templates
                    </Button>
                    <span className="text-sm font-semibold text-foreground/80">{template.title}</span>
                </div>
            </div>

            {/* Blank Details Canvas for future phase */}
            <div className="flex-1 rounded-lg border border-dashed border-border/60 bg-muted/20 p-8 flex items-center justify-center">
                {/* Deliberately left blank placeholder as requested */}
            </div>
        </div>
    );
}

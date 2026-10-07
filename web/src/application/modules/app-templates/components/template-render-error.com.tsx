import React from "react";

import { AlertTriangle, ArrowLeft, RotateCw } from "lucide-react";

import { Button } from "@/components/ui/button";

interface TemplateRenderErrorProps {
    title: string;
    message?: string;
    /** Loads or draws the template again. */
    onRetry?: () => void;
    onBack?: () => void;
}

/**
 * What the store shows in place of a template it could not load or draw: what
 * happened, and a way on - again, or back to the catalog.
 */
export function TemplateRenderError({ title, message, onRetry, onBack }: TemplateRenderErrorProps) {
    return (
        <div
            role="alert"
            className="flex flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-border/70 bg-card/50 p-10 text-center min-h-[240px]"
        >
            <AlertTriangle className="size-8 text-amber-500/80" />
            <h3 className="text-sm font-semibold text-foreground">{title}</h3>
            {message ? <p className="max-w-md text-xs text-muted-foreground break-words">{message}</p> : null}
            <div className="mt-2 flex items-center gap-2">
                {onRetry ? (
                    <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={onRetry}
                    >
                        <RotateCw className="mr-1.5 size-3.5" />
                        Try again
                    </Button>
                ) : null}
                {onBack ? (
                    <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={onBack}
                    >
                        <ArrowLeft className="mr-1.5 size-3.5" />
                        Back to templates
                    </Button>
                ) : null}
            </div>
        </div>
    );
}

/** A card the store could not draw, standing in its place in the grid. */
export function UnrenderableTemplateCard({ name }: { name: string }) {
    return (
        <div
            role="alert"
            className="flex flex-col items-center justify-center gap-1 rounded-xl border border-dashed border-border/70 bg-card/50 p-4 text-center min-h-[160px]"
        >
            <AlertTriangle className="size-5 text-amber-500/80" />
            <p className="text-xs font-medium text-foreground">{name}</p>
            <p className="text-[11px] text-muted-foreground">This template cannot be shown.</p>
        </div>
    );
}

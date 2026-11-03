import { cn } from "@/lib/utils";

import { Button, Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui";

import type { LogsViewerToolbarIconButtonProps } from "../../logs-viewer.types";

export function LogsViewerToolbarIconButton({
    label,
    isActive = false,
    children,
    onClick,
}: LogsViewerToolbarIconButtonProps) {
    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <Button
                    type="button"
                    variant="ghost"
                    size="icon-sm"
                    className={cn(
                        "text-muted-foreground hover:text-foreground",
                        isActive &&
                            "text-amber-500 dark:text-amber-400 bg-amber-500/[0.08] dark:bg-amber-400/[0.1] hover:bg-amber-500/15 dark:hover:bg-amber-400/20 hover:text-amber-500 dark:hover:text-amber-400",
                    )}
                    aria-label={label}
                    title={label}
                    aria-pressed={isActive}
                    onClick={onClick}
                >
                    {children}
                </Button>
            </TooltipTrigger>
            <TooltipContent side="top">{label}</TooltipContent>
        </Tooltip>
    );
}

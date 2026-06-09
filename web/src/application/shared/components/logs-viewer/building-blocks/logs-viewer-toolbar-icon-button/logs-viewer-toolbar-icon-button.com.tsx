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
                    className={cn("text-muted-foreground hover:text-foreground", isActive && "text-primary")}
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

import { cn } from "@lib/utils";
import { LayoutDashboard, List } from "lucide-react";

import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui";

import type { ArgViewModeSwitchProps } from "./arg-view-mode-switch.helpers";

export function ArgViewModeSwitch({ value, onChange, disabled = false, className }: ArgViewModeSwitchProps) {
    return (
        <div
            className={cn(
                "inline-flex items-center rounded-lg border border-border/70 bg-muted/30 p-0.5 shadow-2xs",
                disabled && "opacity-60 pointer-events-none",
                className,
            )}
            role="radiogroup"
            aria-label="Arguments view mode"
        >
            <Tooltip>
                <TooltipTrigger asChild>
                    <button
                        type="button"
                        role="radio"
                        aria-checked={value === "list"}
                        aria-label="List view"
                        disabled={disabled}
                        onClick={() => {
                            onChange("list");
                        }}
                        className={cn(
                            "flex size-7 items-center justify-center rounded-md text-xs font-medium transition-all",
                            value === "list"
                                ? "bg-primary text-primary-foreground shadow-xs"
                                : "text-muted-foreground hover:bg-muted hover:text-foreground",
                        )}
                    >
                        <List className="size-4" />
                    </button>
                </TooltipTrigger>
                <TooltipContent>List view (1 per line)</TooltipContent>
            </Tooltip>

            <Tooltip>
                <TooltipTrigger asChild>
                    <button
                        type="button"
                        role="radio"
                        aria-checked={value === "grid"}
                        aria-label="Grid view"
                        disabled={disabled}
                        onClick={() => {
                            onChange("grid");
                        }}
                        className={cn(
                            "flex size-7 items-center justify-center rounded-md text-xs font-medium transition-all",
                            value === "grid"
                                ? "bg-primary text-primary-foreground shadow-xs"
                                : "text-muted-foreground hover:bg-muted hover:text-foreground",
                        )}
                    >
                        <LayoutDashboard className="size-4" />
                    </button>
                </TooltipTrigger>
                <TooltipContent>Grid view (compact wrap)</TooltipContent>
            </Tooltip>
        </div>
    );
}

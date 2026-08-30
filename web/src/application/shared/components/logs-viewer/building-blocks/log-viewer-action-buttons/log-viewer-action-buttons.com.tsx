import { cn } from "@lib/utils";
import { ALargeSmall, ChevronsLeftRight, ChevronsUpDown, Maximize2, Minimize2 } from "lucide-react";

import { Button, Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui";

import { TerminalThemePicker } from "../terminal-theme-picker";

export interface LogViewerActionButtonsProps {
    isFullscreen?: boolean;
    isFullView?: boolean;
    isFullHeight?: boolean;
    fontSize?: number;
    themeId?: string;
    onToggleFullscreen?: () => void;
    onToggleFullView?: () => void;
    onToggleFullHeight?: () => void;
    onCycleFontSize?: () => void;
    onSelectTheme?: (themeId: string) => void;
}

export function LogViewerActionButtons({
    isFullscreen,
    isFullView,
    isFullHeight,
    fontSize,
    themeId,
    onToggleFullscreen,
    onToggleFullView,
    onToggleFullHeight,
    onCycleFontSize,
    onSelectTheme,
}: LogViewerActionButtonsProps) {
    return (
        <div className="flex items-center gap-1">
            {onSelectTheme && (
                <TerminalThemePicker
                    currentThemeId={themeId}
                    onSelectTheme={onSelectTheme}
                />
            )}

            {onCycleFontSize && fontSize && (
                <Tooltip>
                    <TooltipTrigger asChild>
                        <Button
                            type="button"
                            variant="ghost"
                            size="icon-sm"
                            aria-label={`Text size: ${fontSize}px`}
                            onClick={event => {
                                event.stopPropagation();
                                onCycleFontSize();
                            }}
                        >
                            <ALargeSmall className="size-4" />
                        </Button>
                    </TooltipTrigger>
                    <TooltipContent>{`Text size: ${fontSize}px`}</TooltipContent>
                </Tooltip>
            )}

            {onToggleFullHeight && (
                <Tooltip>
                    <TooltipTrigger asChild>
                        <Button
                            type="button"
                            variant="ghost"
                            size="icon-sm"
                            aria-label={isFullHeight ? "Exit full height" : "Full height"}
                            className={cn(
                                isFullHeight &&
                                    "text-amber-500 dark:text-amber-400 bg-amber-500/[0.08] dark:bg-amber-400/[0.1] hover:bg-amber-500/15 dark:hover:bg-amber-400/20 hover:text-amber-500 dark:hover:text-amber-400",
                            )}
                            onClick={event => {
                                event.stopPropagation();
                                onToggleFullHeight();
                            }}
                        >
                            <ChevronsUpDown className="size-4" />
                        </Button>
                    </TooltipTrigger>
                    <TooltipContent>{isFullHeight ? "Exit full height" : "Full height"}</TooltipContent>
                </Tooltip>
            )}

            {onToggleFullView && (
                <Tooltip>
                    <TooltipTrigger asChild>
                        <Button
                            type="button"
                            variant="ghost"
                            size="icon-sm"
                            aria-label={isFullView ? "Exit full view" : "Full view"}
                            className={cn(
                                "hidden md:inline-flex",
                                isFullView &&
                                    "text-amber-500 dark:text-amber-400 bg-amber-500/[0.08] dark:bg-amber-400/[0.1] hover:bg-amber-500/15 dark:hover:bg-amber-400/20 hover:text-amber-500 dark:hover:text-amber-400",
                            )}
                            onClick={event => {
                                event.stopPropagation();
                                onToggleFullView();
                            }}
                        >
                            <ChevronsLeftRight className="size-4" />
                        </Button>
                    </TooltipTrigger>
                    <TooltipContent>{isFullView ? "Exit full view" : "Full view"}</TooltipContent>
                </Tooltip>
            )}

            {onToggleFullscreen && (
                <Tooltip>
                    <TooltipTrigger asChild>
                        <Button
                            type="button"
                            variant="ghost"
                            size="icon-sm"
                            aria-label={isFullscreen ? "Exit fullscreen" : "Fullscreen logs"}
                            className={cn(
                                isFullscreen &&
                                    "text-amber-500 dark:text-amber-400 bg-amber-500/[0.08] dark:bg-amber-400/[0.1] hover:bg-amber-500/15 dark:hover:bg-amber-400/20 hover:text-amber-500 dark:hover:text-amber-400",
                            )}
                            onClick={event => {
                                event.stopPropagation();
                                onToggleFullscreen();
                            }}
                        >
                            {isFullscreen ? <Minimize2 className="size-4" /> : <Maximize2 className="size-4" />}
                        </Button>
                    </TooltipTrigger>
                    <TooltipContent>{isFullscreen ? "Exit fullscreen" : "Fullscreen logs"}</TooltipContent>
                </Tooltip>
            )}
        </div>
    );
}

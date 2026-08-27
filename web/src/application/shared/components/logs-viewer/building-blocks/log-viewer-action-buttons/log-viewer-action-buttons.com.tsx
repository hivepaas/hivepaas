import { Maximize2, Minimize2 } from "lucide-react";

import { TextZoomIcon } from "@assets/icons";

import { Button, Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui";

export interface LogViewerActionButtonsProps {
    isFullscreen?: boolean;
    fontSize?: number;
    onToggleFullscreen?: () => void;
    onCycleFontSize?: () => void;
}

export function LogViewerActionButtons({
    isFullscreen,
    fontSize,
    onToggleFullscreen,
    onCycleFontSize,
}: LogViewerActionButtonsProps) {
    return (
        <div className="flex items-center gap-1">
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
                            <TextZoomIcon className="size-4" />
                        </Button>
                    </TooltipTrigger>
                    <TooltipContent>{`Text size: ${fontSize}px`}</TooltipContent>
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

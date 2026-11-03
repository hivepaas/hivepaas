import { Check, Palette, Sparkles } from "lucide-react";

import {
    Button,
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
    Tooltip,
    TooltipContent,
    TooltipTrigger,
} from "@/components/ui";

import { TERMINAL_THEMES, type TerminalThemeDefinition, getTerminalTheme } from "../../logs-viewer.themes";

export interface TerminalThemePickerProps {
    currentThemeId?: string;
    onSelectTheme?: (themeId: string) => void;
}

export function TerminalThemePicker({ currentThemeId = "default", onSelectTheme }: TerminalThemePickerProps) {
    const allThemes = Object.values(TERMINAL_THEMES);
    const gradientThemes = allThemes.filter(t => t.category === "gradient");
    const classicThemes = allThemes.filter(t => t.category === "classic");
    const activeTheme = getTerminalTheme(currentThemeId);

    const renderThemeItem = (t: TerminalThemeDefinition) => {
        const isSelected = t.id === currentThemeId;
        const isGradient = t.category === "gradient";

        return (
            <DropdownMenuItem
                key={t.id}
                onClick={() => onSelectTheme?.(t.id)}
                className="flex items-center justify-between cursor-pointer py-1.5 px-2 text-xs rounded-sm transition-colors"
            >
                <div className="flex items-center gap-2.5 min-w-0">
                    <span
                        className="size-3.5 rounded-full border border-border/70 shadow-xs flex-shrink-0"
                        style={{
                            background: isGradient
                                ? t.background
                                : `linear-gradient(135deg, ${t.accentColor} 50%, ${t.backgroundColor} 50%)`,
                        }}
                    />
                    <div className="flex flex-col min-w-0">
                        <span className="font-medium text-foreground truncate">{t.label}</span>
                    </div>
                </div>
                {isSelected && <Check className="size-3.5 text-primary ml-auto flex-shrink-0" />}
            </DropdownMenuItem>
        );
    };

    return (
        <DropdownMenu>
            <Tooltip>
                <TooltipTrigger asChild>
                    <DropdownMenuTrigger asChild>
                        <Button
                            type="button"
                            variant="ghost"
                            size="icon-sm"
                            aria-label={`Theme: ${activeTheme.label}`}
                            onClick={event => {
                                event.stopPropagation();
                            }}
                        >
                            <Palette className="size-4" />
                        </Button>
                    </DropdownMenuTrigger>
                </TooltipTrigger>
                <TooltipContent>{`Theme: ${activeTheme.label}`}</TooltipContent>
            </Tooltip>

            <DropdownMenuContent
                align="end"
                className="w-60 p-1.5 max-h-[420px] overflow-y-auto"
                onClick={event => {
                    event.stopPropagation();
                }}
            >
                {gradientThemes.length > 0 && (
                    <>
                        <DropdownMenuLabel className="text-[11px] font-semibold tracking-wider text-muted-foreground uppercase px-2 py-1 flex items-center gap-1.5">
                            <Sparkles className="size-3 text-amber-500" />
                            Gradient Themes
                        </DropdownMenuLabel>
                        {gradientThemes.map(renderThemeItem)}
                        <DropdownMenuSeparator className="my-1" />
                    </>
                )}

                <DropdownMenuLabel className="text-[11px] font-semibold tracking-wider text-muted-foreground uppercase px-2 py-1">
                    Classic Themes
                </DropdownMenuLabel>
                {classicThemes.map(renderThemeItem)}
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

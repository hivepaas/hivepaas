import { type KeyboardEventHandler, type ReactNode, useRef, useState } from "react";

import { cn } from "@/lib/utils";
import { CaseSensitive, Regex, Search, X } from "lucide-react";

import type { SearchMode } from "./collapsible-search-input.constants";

/**
 * CollapsibleSearchInput is a search field that sits as its own icon until it
 * is used.
 *
 * The input is never unmounted. Collapsing shrinks it to the width of the icon
 * drawn on top of it, so the icon a person clicks is the field itself and focus
 * lands where they aimed - no click that opens a box they then have to click
 * again. Collapsed, it is the same 32px box as the toolbar's icon buttons and
 * belongs in the same group as them, or the row's spacing reads as uneven.
 *
 * A field holding a term stays open whether or not it has focus. These fields
 * are doing something while they hold text - filtering what the server returns,
 * or holding a match count and its controls - and work in progress has to stay
 * visible.
 */
export function CollapsibleSearchInput({
    value,
    placeholder,
    label,
    expandedClassName = "w-40 sm:w-56",
    trailing,
    mode,
    isInvalid = false,
    onValueChange,
    onModeChange,
    onClear,
    onKeyDown,
}: CollapsibleSearchInputProps) {
    const containerRef = useRef<HTMLDivElement>(null);
    const [hasFocus, setHasFocus] = useState(false);
    const isExpanded = hasFocus || value !== "";
    const hasModes = mode !== undefined && onModeChange !== undefined;
    const hasRail = Boolean(trailing) || hasModes || (onClear !== undefined && value !== "");

    return (
        <div
            ref={containerRef}
            className={cn(
                "group relative flex min-w-0 max-w-full items-center transition-[width] duration-200 ease-out",
                isExpanded ? expandedClassName : "w-8",
            )}
            onFocus={() => {
                setHasFocus(true);
            }}
            onBlur={() => {
                // Not relatedTarget: clearing the field removes the very button
                // that was clicked, and a blur from a removed element reports
                // nowhere. Ask where focus actually landed instead, once the
                // browser has finished moving it.
                requestAnimationFrame(() => {
                    setHasFocus(containerRef.current?.contains(document.activeElement) ?? false);
                });
            }}
        >
            <Search
                className={cn(
                    "pointer-events-none absolute transition-[left] duration-200 ease-out",
                    isExpanded
                        ? "left-2.5 size-3.5 text-muted-foreground"
                        : "left-1/2 size-3.5 -translate-x-1/2 text-muted-foreground group-hover:text-foreground sm:size-4",
                )}
            />
            <input
                type="text"
                aria-label={label}
                placeholder={isExpanded ? placeholder : undefined}
                value={value}
                title={isExpanded ? undefined : label}
                className={cn(
                    "w-full rounded-md border py-1 text-xs text-foreground placeholder:text-muted-foreground focus:outline-none sm:text-sm",
                    isExpanded
                        ? cn(
                              "h-8 bg-background/50 pl-8 focus:bg-background focus:ring-1 sm:h-9",
                              isInvalid ? "border-destructive focus:ring-destructive" : "border-input focus:ring-ring",
                              railPadding({ trailing, hasModes, hasClear: onClear !== undefined && value !== "" }),
                          )
                        : // No border of its own and the same box an icon button
                          // is - just the hover.
                          "size-8 cursor-pointer border-transparent bg-transparent pl-8 pr-0 hover:bg-muted",
                )}
                onChange={event => {
                    onValueChange(event.target.value);
                }}
                onKeyDown={onKeyDown}
            />
            {isExpanded && hasRail && (
                <div className="absolute right-1.5 flex items-center gap-0.5 text-muted-foreground">
                    {trailing}
                    {hasModes && (
                        <>
                            <ModeToggle
                                label="Match case"
                                isOn={mode.isCaseSensitive}
                                onToggle={() => {
                                    onModeChange({ ...mode, isCaseSensitive: !mode.isCaseSensitive });
                                }}
                            >
                                <CaseSensitive className="size-3.5 sm:size-4" />
                            </ModeToggle>
                            <ModeToggle
                                label="Use regular expression"
                                isOn={mode.isRegex}
                                onToggle={() => {
                                    onModeChange({ ...mode, isRegex: !mode.isRegex });
                                }}
                            >
                                <Regex className="size-3.5 sm:size-4" />
                            </ModeToggle>
                        </>
                    )}
                    {onClear !== undefined && value !== "" && (
                        <button
                            type="button"
                            aria-label={`Clear ${label.toLowerCase()}`}
                            className="rounded p-1 hover:bg-muted hover:text-foreground"
                            onClick={onClear}
                        >
                            <X className="size-3 sm:size-3.5" />
                        </button>
                    )}
                </div>
            )}
        </div>
    );
}

/**
 * ModeToggle is one of the switches that change what the text means, drawn the
 * way an editor's find box draws them: pressed rather than checked.
 */
function ModeToggle({ label, isOn, children, onToggle }: ModeToggleProps) {
    return (
        <button
            type="button"
            aria-label={label}
            aria-pressed={isOn}
            title={label}
            className={cn(
                "rounded p-1 hover:bg-muted hover:text-foreground",
                isOn && "bg-primary/15 text-primary hover:bg-primary/20 hover:text-primary",
            )}
            onClick={onToggle}
        >
            {children}
        </button>
    );
}

/** Room on the right for whatever the rail holds, and none when it holds nothing. */
function railPadding({ trailing, hasModes, hasClear }: RailContents): string {
    const slots = (trailing ? 3 : 0) + (hasModes ? 2 : 0) + (hasClear ? 1 : 0);

    if (slots === 0) {
        return "pr-3";
    }
    if (slots <= 2) {
        return "pr-14";
    }
    return slots <= 4 ? "pr-20" : "pr-32";
}

interface RailContents {
    trailing: ReactNode;
    hasModes: boolean;
    hasClear: boolean;
}

interface ModeToggleProps {
    label: string;
    isOn: boolean;
    children: ReactNode;
    onToggle: () => void;
}

interface CollapsibleSearchInputProps {
    value: string;
    placeholder: string;
    /** What the field is, for anyone who only ever sees the collapsed icon. */
    label: string;
    /** Width once open. Defaults to the toolbar's own search width. */
    expandedClassName?: string;
    /** Extra controls for the right-hand rail, drawn before the mode toggles. */
    trailing?: ReactNode;
    /** Given with onModeChange, the rail carries match-case and regex toggles. */
    mode?: SearchMode;
    /** Marks the field as holding something the search cannot use. */
    isInvalid?: boolean;
    onValueChange: (value: string) => void;
    onModeChange?: (mode: SearchMode) => void;
    /** Given, a clear button appears in the rail whenever the field holds text. */
    onClear?: () => void;
    onKeyDown?: KeyboardEventHandler<HTMLInputElement>;
}

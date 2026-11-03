import { cn } from "@lib/utils";
import { type LucideIcon } from "lucide-react";

export interface OptionCard<TValue extends string> {
    value: TValue;
    label: string;
    description: string;
    icon: LucideIcon;
    /** One option that cannot be picked right now, while the others still can. */
    disabled?: boolean;
}

interface Props<TValue extends string> {
    options: OptionCard<TValue>[];
    value: TValue | undefined;
    onChange: (value: TValue) => void;
    /** Grid columns and width, e.g. "grid-cols-1 sm:grid-cols-2 max-w-[520px]". */
    className?: string;
    readOnly?: boolean;
}

/**
 * A row of cards to pick one of a few options from.
 *
 * It is for the choice a form is built around - the one that decides which
 * fields come next - where a card with a line of explanation is worth the room
 * it takes. An ordinary choice belongs in a select or a tab list.
 */
export function OptionCardGroup<TValue extends string>({
    options,
    value,
    onChange,
    className,
    readOnly = false,
}: Props<TValue>) {
    return (
        <div className={cn("grid w-full gap-3", className)}>
            {options.map(option => {
                const isSelected = value === option.value;
                const isLocked = readOnly || option.disabled === true;
                const Icon = option.icon;

                return (
                    <div
                        key={option.value}
                        role="button"
                        tabIndex={isLocked ? -1 : 0}
                        aria-pressed={isSelected}
                        aria-disabled={isLocked || undefined}
                        onClick={() => {
                            if (!isLocked) {
                                onChange(option.value);
                            }
                        }}
                        onKeyDown={e => {
                            if (!isLocked && (e.key === "Enter" || e.key === " ")) {
                                e.preventDefault();
                                onChange(option.value);
                            }
                        }}
                        className={cn(
                            "flex flex-col items-start gap-1.5 p-3 rounded-lg border text-left transition-all select-none",
                            isLocked && "pointer-events-none cursor-not-allowed",
                            !isLocked && "cursor-pointer",
                            // A whole group nobody may touch reads as one dimmed block; a single
                            // card among live ones has to look refused next to them.
                            readOnly && "opacity-70",
                            option.disabled === true && "opacity-50",
                            isSelected
                                ? "border-primary bg-primary/5 ring-1 ring-primary text-foreground"
                                : "border-border hover:border-primary/50 hover:bg-accent/40 text-muted-foreground",
                        )}
                    >
                        <div className="flex items-center gap-2">
                            <Icon
                                className={cn("size-4 shrink-0", isSelected ? "text-primary" : "text-muted-foreground")}
                            />
                            <span
                                className={cn(
                                    "font-medium text-sm",
                                    isSelected ? "text-foreground" : "text-muted-foreground",
                                )}
                            >
                                {option.label}
                            </span>
                        </div>
                        <p className="text-xs text-muted-foreground line-clamp-2 leading-relaxed">
                            {option.description}
                        </p>
                    </div>
                );
            })}
        </div>
    );
}

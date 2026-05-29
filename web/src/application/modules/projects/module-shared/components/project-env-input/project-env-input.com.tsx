import * as React from "react";

import { cn } from "@/lib/utils";
import { Input } from "@components/ui";
import { Plus, X } from "lucide-react";
import { type ProjectEnvEntity } from "~/projects/domain";

const DEFAULT_ENV_COLOR = "#64748b";
const MAX_PROJECT_ENVS = 10;

export interface ProjectEnvInputProps {
    envs: ProjectEnvEntity[];
    onCreate: (env: ProjectEnvEntity) => void;
    onDelete: (envName: string) => void;
    onUpdateColor: (envName: string, color: string) => void;
    placeholder?: string;
    className?: string;
    disabled?: boolean;
}

export function ProjectEnvInput({
    envs,
    onCreate,
    onDelete,
    onUpdateColor,
    placeholder = "Enter environment",
    className,
    disabled = false,
}: ProjectEnvInputProps) {
    const [isInputVisible, setIsInputVisible] = React.useState(false);
    const [inputValue, setInputValue] = React.useState("");
    const [selectedColor, setSelectedColor] = React.useState(DEFAULT_ENV_COLOR);
    const [draftColors, setDraftColors] = React.useState<Record<string, string>>({});
    const inputRef = React.useRef<HTMLInputElement>(null);
    const colorInputIdPrefix = React.useId();
    const newColorInputId = React.useId();
    const hasReachedMax = envs.length >= MAX_PROJECT_ENVS;

    React.useEffect(() => {
        if (isInputVisible && inputRef.current) {
            inputRef.current.focus();
        }
    }, [isInputVisible]);

    React.useEffect(() => {
        if (disabled) {
            setIsInputVisible(false);
            setInputValue("");
            setSelectedColor(DEFAULT_ENV_COLOR);
        }
    }, [disabled]);

    React.useEffect(() => {
        setDraftColors(previous => {
            const next: Record<string, string> = {};
            envs.forEach(env => {
                const draftColor = previous[env.name];
                if (draftColor !== undefined && draftColor !== env.color) {
                    next[env.name] = draftColor;
                }
            });
            return next;
        });
    }, [envs]);

    function handleCreateEnv() {
        if (disabled) {
            return;
        }

        const trimmedValue = inputValue.trim();
        if (!hasReachedMax && trimmedValue && !envs.some(env => env.name === trimmedValue)) {
            onCreate({ name: trimmedValue, color: selectedColor });
            setInputValue("");
            setSelectedColor(DEFAULT_ENV_COLOR);
            setIsInputVisible(false);
        }
    }

    function handleKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
        if (event.key === "Enter") {
            event.preventDefault();
            handleCreateEnv();
        } else if (event.key === "Escape") {
            setIsInputVisible(false);
            setInputValue("");
            setSelectedColor(DEFAULT_ENV_COLOR);
        }
    }

    function handleBlur() {
        if (inputValue.trim()) {
            handleCreateEnv();
        } else {
            setIsInputVisible(false);
            setInputValue("");
            setSelectedColor(DEFAULT_ENV_COLOR);
        }
    }

    function handleDraftColor(envName: string, color: string) {
        if (disabled) {
            return;
        }

        setDraftColors(previous => ({
            ...previous,
            [envName]: color,
        }));
    }

    function handleCommitColor(env: ProjectEnvEntity) {
        if (disabled) {
            return;
        }

        const draftColor = draftColors[env.name];
        if (!draftColor || draftColor === env.color) {
            return;
        }

        onUpdateColor(env.name, draftColor);
    }

    return (
        <div className={cn("flex flex-wrap items-center gap-2", className)}>
            {envs.map((env, index) => {
                const colorInputId = `${colorInputIdPrefix}-${index}`;
                const displayColor = draftColors[env.name] ?? env.color;
                return (
                    <div
                        key={env.name}
                        className="inline-flex h-8 items-center gap-1.5 rounded-md px-2.5 py-1 text-white shadow-xs"
                        style={{ backgroundColor: displayColor }}
                    >
                        <span className="text-sm font-medium">{env.name}</span>
                        <label
                            htmlFor={colorInputId}
                            className={cn(
                                "inline-flex size-5 cursor-pointer items-center justify-center rounded-sm border-2 border-white/80 transition-colors hover:border-white focus-within:ring-2 focus-within:ring-white/80",
                                disabled && "cursor-not-allowed opacity-70",
                            )}
                            style={{ backgroundColor: displayColor }}
                            aria-label={`Change ${env.name} color`}
                        >
                            <input
                                id={colorInputId}
                                type="color"
                                value={displayColor}
                                onChange={event => {
                                    handleDraftColor(env.name, event.target.value);
                                }}
                                onBlur={() => {
                                    handleCommitColor(env);
                                }}
                                className="sr-only"
                                disabled={disabled}
                            />
                        </label>
                        <button
                            type="button"
                            onClick={() => {
                                if (disabled) {
                                    return;
                                }

                                onDelete(env.name);
                            }}
                            disabled={disabled}
                            className="rounded-sm p-0.5 transition-colors hover:bg-white/20 focus:outline-none focus:ring-2 focus:ring-white/80 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50"
                            aria-label={`Remove ${env.name}`}
                        >
                            <X className="size-3.5 text-white/90" />
                        </button>
                    </div>
                );
            })}

            {isInputVisible ? (
                <div className="flex items-center gap-2">
                    <Input
                        ref={inputRef}
                        type="text"
                        value={inputValue}
                        onChange={event => {
                            setInputValue(event.target.value);
                        }}
                        onKeyDown={handleKeyDown}
                        onBlur={handleBlur}
                        placeholder={placeholder}
                        className="h-8 w-36"
                        disabled={disabled}
                    />
                    <label
                        htmlFor={newColorInputId}
                        className={cn(
                            "inline-flex size-8 cursor-pointer items-center justify-center rounded-md border border-border transition-colors focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-2",
                            disabled && "cursor-not-allowed opacity-70",
                        )}
                        style={{ backgroundColor: selectedColor }}
                        aria-label="New environment color"
                    >
                        <input
                            id={newColorInputId}
                            type="color"
                            value={selectedColor}
                            onChange={event => {
                                setSelectedColor(event.target.value);
                            }}
                            className="sr-only"
                            disabled={disabled}
                        />
                    </label>
                </div>
            ) : hasReachedMax || disabled ? (
                <button
                    type="button"
                    disabled
                    className="inline-flex size-8 cursor-not-allowed items-center justify-center rounded-md border border-border bg-secondary text-secondary-foreground opacity-50"
                    title={disabled ? "Environment editing is disabled" : `Maximum ${MAX_PROJECT_ENVS} environments`}
                    aria-label={disabled ? "Environment editing is disabled" : `Maximum ${MAX_PROJECT_ENVS} environments`}
                >
                    <Plus className="size-4" />
                </button>
            ) : (
                <button
                    type="button"
                    onClick={() => {
                        setIsInputVisible(true);
                    }}
                    className="inline-flex size-8 items-center justify-center rounded-md border border-border bg-secondary text-secondary-foreground transition-colors hover:bg-secondary/80 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                    aria-label="Add environment"
                >
                    <Plus className="size-4" />
                </button>
            )}
            <span className="text-xs text-muted-foreground">
                {envs.length}/{MAX_PROJECT_ENVS}
            </span>
        </div>
    );
}

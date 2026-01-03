import * as React from "react";

import { cn } from "@/lib/utils";
import { Plus, X } from "lucide-react";

import { Input } from "@/components/ui/input";

export interface TagInputProps {
    tags: string[];
    onCreate: (tag: string) => void;
    onDelete: (tag: string) => void;
    placeholder?: string;
    className?: string;
}

export function TagInput({ tags, onCreate, onDelete, placeholder = "Enter tag", className }: TagInputProps) {
    const [isInputVisible, setIsInputVisible] = React.useState(false);
    const [inputValue, setInputValue] = React.useState("");
    const inputRef = React.useRef<HTMLInputElement>(null);

    React.useEffect(() => {
        if (isInputVisible && inputRef.current) {
            inputRef.current.focus();
        }
    }, [isInputVisible]);

    function handleCreateTag() {
        const trimmedValue = inputValue.trim();
        if (trimmedValue && !tags.includes(trimmedValue)) {
            onCreate(trimmedValue);
            setInputValue("");
            setIsInputVisible(false);
        }
    }

    function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
        if (e.key === "Enter") {
            e.preventDefault();
            handleCreateTag();
        } else if (e.key === "Escape") {
            setIsInputVisible(false);
            setInputValue("");
        }
    }

    function handleBlur() {
        if (inputValue.trim()) {
            handleCreateTag();
        } else {
            setIsInputVisible(false);
            setInputValue("");
        }
    }

    return (
        <div className={cn("flex flex-wrap items-center gap-2", className)}>
            {tags.map(tag => (
                <div
                    key={tag}
                    className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-secondary text-secondary-foreground border border-border"
                >
                    <span className="text-sm font-normal">{tag}</span>
                    <button
                        type="button"
                        onClick={() => {
                            onDelete(tag);
                        }}
                        className="ml-0.5 rounded-sm hover:bg-secondary-foreground/20 p-0.5 transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                        aria-label={`Remove ${tag}`}
                    >
                        <X className="h-3 w-3 text-secondary-foreground/70" />
                    </button>
                </div>
            ))}

            {isInputVisible ? (
                <Input
                    ref={inputRef}
                    type="text"
                    value={inputValue}
                    onChange={e => {
                        setInputValue(e.target.value);
                    }}
                    onKeyDown={handleKeyDown}
                    onBlur={handleBlur}
                    placeholder={placeholder}
                    className="h-8 w-24 min-w-[100px]"
                />
            ) : (
                <button
                    type="button"
                    onClick={() => {
                        setIsInputVisible(true);
                    }}
                    className="inline-flex items-center justify-center rounded-md bg-secondary text-secondary-foreground h-8 w-8 hover:bg-secondary/80 transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 border border-border"
                    aria-label="Add tag"
                >
                    <Plus className="h-4 w-4" />
                </button>
            )}
        </div>
    );
}

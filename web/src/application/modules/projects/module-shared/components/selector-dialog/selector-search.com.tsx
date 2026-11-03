import { useEffect, useRef } from "react";

import { Search } from "lucide-react";

import { Input } from "@/components/ui";

interface Props {
    value: string;
    onChange: (value: string) => void;
    autoFocus?: boolean;
    placeholder?: string;
}

export function SelectorSearch({ value, onChange, autoFocus = true, placeholder = "Search" }: Props) {
    const inputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        if (autoFocus) {
            inputRef.current?.focus();
        }
    }, [autoFocus]);

    return (
        <div className="relative w-full max-w-[320px]">
            <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center justify-center pl-3 text-muted-foreground">
                <Search className="size-4" />
                <span className="sr-only">Search</span>
            </div>
            <Input
                ref={inputRef}
                value={value}
                type="search"
                placeholder={placeholder}
                className="px-9 [&::-webkit-search-cancel-button]:appearance-none [&::-webkit-search-decoration]:appearance-none [&::-webkit-search-results-button]:appearance-none [&::-webkit-search-results-decoration]:appearance-none"
                onChange={event => {
                    onChange(event.target.value);
                }}
            />
        </div>
    );
}

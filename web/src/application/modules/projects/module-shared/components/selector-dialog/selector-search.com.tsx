import { Search } from "lucide-react";

import { Input } from "@/components/ui";

interface Props {
    value: string;
    onChange: (value: string) => void;
}

export function SelectorSearch({ value, onChange }: Props) {
    return (
        <div className="relative w-full max-w-[320px]">
            <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center justify-center pl-3 text-muted-foreground">
                <Search className="size-4" />
                <span className="sr-only">Search</span>
            </div>
            <Input
                value={value}
                type="search"
                placeholder="Search"
                className="px-9 [&::-webkit-search-cancel-button]:appearance-none [&::-webkit-search-decoration]:appearance-none [&::-webkit-search-results-button]:appearance-none [&::-webkit-search-results-decoration]:appearance-none"
                onChange={event => {
                    onChange(event.target.value);
                }}
            />
        </div>
    );
}

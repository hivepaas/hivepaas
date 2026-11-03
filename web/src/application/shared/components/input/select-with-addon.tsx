import React from "react";

import { cn } from "@lib/utils";

import { Select, SelectContent, SelectTrigger, SelectValue } from "@/components/ui/select";

function View({
    addonLeft,
    addonRight,
    classNameContainer,
    triggerClassName,
    contentClassName,
    placeholder,
    children,
    ...selectProps
}: SelectWithAddonProps) {
    return (
        <div className={cn("w-full space-y-2", classNameContainer)}>
            <div className="flex min-w-0 items-stretch rounded-md shadow-xs">
                {addonLeft && (
                    <span className="border-input bg-background inline-flex shrink-0 items-center rounded-l-md border px-3 text-sm">
                        {addonLeft}
                    </span>
                )}

                <Select {...selectProps}>
                    <SelectTrigger
                        className={cn(
                            "-mx-px h-9 min-w-0 flex-1 py-1 w-full rounded-none shadow-none",
                            addonLeft ? "rounded-l-none rounded-r-md" : "",
                            addonRight ? "rounded-r-none rounded-l-md" : "",
                            triggerClassName,
                        )}
                    >
                        <SelectValue placeholder={placeholder} />
                    </SelectTrigger>
                    <SelectContent className={contentClassName}>{children}</SelectContent>
                </Select>

                {addonRight && (
                    <span className="border-input bg-background inline-flex shrink-0 items-center rounded-r-md border px-3 text-sm">
                        {addonRight}
                    </span>
                )}
            </div>
        </div>
    );
}

type SelectWithAddonProps = React.ComponentProps<typeof Select> & {
    addonLeft?: string;
    addonRight?: string;
    classNameContainer?: string;
    triggerClassName?: string;
    contentClassName?: string;
    placeholder?: string;
    children: React.ReactNode;
};

export const SelectWithAddon = React.memo(View);

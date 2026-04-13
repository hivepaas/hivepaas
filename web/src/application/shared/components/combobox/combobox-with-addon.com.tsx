"use client";

import React from "react";

import { cn } from "@lib/utils";

import { Combobox, type ComboboxProps } from "./combobox.com";

function View<T extends Record<string, unknown> = Record<string, unknown>>({
    addonLeft,
    addonRight,
    classNameContainer,
    comboboxClassName,
    ...comboboxProps
}: ComboboxWithAddonProps<T>) {
    return (
        <div className={cn("w-full space-y-2", classNameContainer)}>
            <div className="flex min-w-0 rounded-md shadow-xs">
                {addonLeft && (
                    <span className="border-input bg-background inline-flex items-center rounded-l-md border px-3 text-sm">
                        {addonLeft}
                    </span>
                )}

                <Combobox
                    {...comboboxProps}
                    className={cn(
                        "-mx-px min-w-0 flex-1",
                        addonLeft ? "[&_button[role='combobox']]:rounded-l-none" : "",
                        addonRight ? "[&_button[role='combobox']]:rounded-r-none" : "",
                        comboboxClassName,
                    )}
                />

                {addonRight && (
                    <span className="border-input bg-background inline-flex items-center rounded-r-md border px-3 text-sm">
                        {addonRight}
                    </span>
                )}
            </div>
        </div>
    );
}

export type ComboboxWithAddonProps<T extends Record<string, unknown> = Record<string, unknown>> = ComboboxProps<T> & {
    addonLeft?: string;
    addonRight?: string;
    classNameContainer?: string;
    comboboxClassName?: string;
};

export const ComboboxWithAddon = React.memo(View) as typeof View;

import React from "react";

import { cn } from "@lib/utils";

import { EditableCombobox, type EditableComboboxProps } from "./editable-combobox.com";

type EditComboboxWithAddOnProps = EditableComboboxProps & {
    addonLeft?: React.ReactNode;
    addonRight?: React.ReactNode;
    classNameContainer?: string;
};

function View({ addonLeft, addonRight, classNameContainer, ...comboboxProps }: EditComboboxWithAddOnProps) {
    return (
        <div className={cn("w-full space-y-2", classNameContainer)}>
            <div className="flex min-w-0 rounded-md shadow-xs">
                {addonLeft && (
                    <span className="border-input bg-background inline-flex items-center rounded-l-md border px-3 text-sm">
                        {addonLeft}
                    </span>
                )}
                <EditableCombobox
                    {...comboboxProps}
                    className={cn("-mx-px min-w-0 flex-1 gap-0", comboboxProps.className)}
                    inputClassName={cn(
                        "border-input rounded-md border shadow-none focus-visible:ring-0",
                        addonLeft ? "rounded-l-none" : "",
                        addonRight ? "rounded-r-none" : "",
                        comboboxProps.inputClassName,
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

export const EditComboboxWithAddOn = React.memo(View);

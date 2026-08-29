import React, { useId } from "react";

import { InputNumber, type NumberInputProps } from "@components/ui/input-number";
import { cn } from "@lib/utils";

function View({ addonLeft, addonRight, classNameContainer, ...inputProps }: InputNumberWithAddonProps) {
    const id = useId();

    return (
        <div className={cn("w-full space-y-2 min-w-0", classNameContainer)}>
            <div className="flex rounded-md shadow-xs min-w-0">
                {addonLeft && (
                    <span className="border-input bg-background inline-flex items-center rounded-l-md border px-2.5 sm:px-3 text-xs sm:text-sm shrink-0">
                        {addonLeft}
                    </span>
                )}
                <InputNumber
                    id={id}
                    {...inputProps}
                    className="min-w-0"
                    classNameInput={cn(
                        "-mx-px rounded-none shadow-none min-w-0",
                        addonLeft ? "rounded-l-none rounded-r-md" : "",
                        addonRight ? "rounded-r-none rounded-l-md" : "",
                        inputProps.className,
                    )}
                />
                {addonRight && (
                    <span className="border-input bg-background inline-flex items-center rounded-r-md border px-2.5 sm:px-3 text-xs sm:text-sm shrink-0">
                        {addonRight}
                    </span>
                )}
            </div>
        </div>
    );
}

type InputNumberWithAddonProps = NumberInputProps & {
    addonLeft?: string;
    addonRight?: string;
    classNameContainer?: string;
};

export const InputNumberWithAddon = React.memo(View);

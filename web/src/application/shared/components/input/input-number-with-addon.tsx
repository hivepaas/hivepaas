import React, { useId } from "react";

import { InputNumber, type NumberInputProps } from "@components/ui/input-number";
import { cn } from "@lib/utils";

function View({ addonLeft, addonRight, classNameContainer, ...inputProps }: InputNumberWithAddonProps) {
    const id = useId();

    return (
        <div className={cn("w-full space-y-2", classNameContainer)}>
            <div className="flex rounded-md shadow-xs">
                {addonLeft && (
                    <span className="border-input bg-background inline-flex items-center rounded-l-md border px-3 text-sm">
                        {addonLeft}
                    </span>
                )}
                <InputNumber
                    id={id}
                    {...inputProps}
                    classNameInput={cn(
                        "-mx-px rounded-none shadow-none",
                        addonLeft ? "rounded-l-none rounded-r-md" : "",
                        addonRight ? "rounded-r-none rounded-l-md" : "",
                        inputProps.className,
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

type InputNumberWithAddonProps = NumberInputProps & {
    addonLeft?: string;
    addonRight?: string;
    classNameContainer?: string;
};

export const InputNumberWithAddon = React.memo(View);

import React, { useId } from "react";

import { cn } from "@lib/utils";

import { Input, type InputProps } from "@/components/ui/input";

function View({ addonLeft, addonRight, classNameContainer, ...inputProps }: InputWithAddOnProps) {
    const id = useId();

    return (
        <div className={cn("w-full space-y-2", classNameContainer)}>
            <div className="flex rounded-md shadow-xs">
                {addonLeft && (
                    <span className="border-input bg-background inline-flex items-center rounded-l-md border px-3 text-sm">
                        {addonLeft}
                    </span>
                )}
                <Input
                    id={id}
                    type="text"
                    {...inputProps}
                    className={cn(
                        "-mx-px rounded-none shadow-none",
                        addonLeft ? "rounded-l-none rounded-r-md" : "",
                        addonRight ? "rounded-r-none rounded-l-md" : "",
                        inputProps.className,
                    )}
                />
            </div>
        </div>
    );
}

type InputWithAddOnProps = InputProps & {
    addonLeft?: string;
    addonRight?: string;
    classNameContainer?: string;
};

export const InputWithAddOn = React.memo(View);

import React, { useId } from "react";

import { Input, type InputProps } from "@/components/ui/input";

function View({ addonLeft, addonRight, ...inputProps }: InputWithAddOnProps) {
    const id = useId();

    return (
        <div className="w-full max-w-xs space-y-2">
            <div className="flex rounded-md shadow-xs">
                {addonLeft && (
                    <span className="border-input bg-background inline-flex items-center rounded-l-md border px-3 text-sm">
                        {addonLeft}
                    </span>
                )}
                <Input
                    id={id}
                    type="text"
                    className={`-mx-px rounded-none shadow-none ${addonLeft ? "rounded-l-none rounded-r-md" : ""} ${addonRight ? "rounded-r-none rounded-l-md" : ""}`}
                    {...inputProps}
                />
            </div>
        </div>
    );
}

type InputWithAddOnProps = InputProps & {
    addonLeft?: string;
    addonRight?: string;
};

export const InputWithAddOn = React.memo(View);

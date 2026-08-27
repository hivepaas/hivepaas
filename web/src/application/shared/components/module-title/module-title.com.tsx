import React, { type PropsWithChildren } from "react";

import { moduleHeaderBox } from "@lib/styles";
import { cn } from "@lib/utils";

export function ModuleTitle({ title, children }: Props) {
    return (
        <div className="flex flex-col gap-2 sm:gap-4">
            <div className={cn(moduleHeaderBox)}>
                <div className="flex items-center justify-between py-0.5 sm:py-1.5">
                    <h1 className="text-base sm:text-lg font-bold text-foreground">{title}</h1>
                </div>
            </div>
            {children}
        </div>
    );
}

type Props = PropsWithChildren<{
    title: string;
}>;

import React, { type PropsWithChildren } from "react";

import { moduleHeaderBox } from "@lib/styles";
import { cn } from "@lib/utils";

export function ModuleTitle({ title, children }: Props) {
    return (
        <div className="flex flex-col gap-2 sm:gap-4">
            <div className={cn(moduleHeaderBox)}>
                <div className="flex items-center justify-between py-0.5 sm:py-1.5">
                    <h1 className="text-base sm:text-lg font-bold text-foreground flex items-center gap-1.5 tracking-tight">
                        <span className="text-amber-500/50 dark:text-amber-400/50 font-mono font-normal select-none">
                            /
                        </span>
                        <span>{title}</span>
                        <span className="text-amber-500/50 dark:text-amber-400/50 font-mono font-normal select-none">
                            /
                        </span>
                    </h1>
                </div>
            </div>
            {children}
        </div>
    );
}

type Props = PropsWithChildren<{
    title: string;
}>;

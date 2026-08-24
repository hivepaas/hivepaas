import React, { type PropsWithChildren } from "react";

import { cn } from "@lib/utils";

function View({ label, headerClassName, className, children }: Props) {
    return (
        <div className={cn("flex flex-col gap-4", className)}>
            <div
                className={cn(
                    "sticky top-0 z-10 rounded-lg bg-accent px-3 py-2 font-medium shadow-xs",
                    headerClassName,
                )}
            >
                {label}
            </div>
            <div className="px-4">{children}</div>
        </div>
    );
}

type Props = PropsWithChildren<{
    label: React.ReactNode;
    headerClassName?: string;
    className?: string;
}>;

export const ContentBlock = React.memo(View);

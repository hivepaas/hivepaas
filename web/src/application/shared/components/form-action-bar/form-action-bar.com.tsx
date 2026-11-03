import { type ComponentProps } from "react";

import { cn } from "@lib/utils";

import { Separator } from "@/components/ui/separator";

export function FormActionBar({
    children,
    className,
    contentClassName,
    sticky = true,
    ...props
}: ComponentProps<"div"> & Props) {
    return (
        <div
            {...props}
            data-form-action-bar=""
            className={cn(
                sticky ? "sticky bottom-0 z-10 mt-6 shrink-0 bg-background" : "mt-6 shrink-0 px-0 pb-6",
                className,
            )}
        >
            {sticky && <Separator className="opacity-50" />}
            <div className={cn("flex items-center justify-end gap-3", sticky && "pt-4 pb-4", contentClassName)}>
                {children}
            </div>
        </div>
    );
}

interface Props {
    contentClassName?: string;
    sticky?: boolean;
}

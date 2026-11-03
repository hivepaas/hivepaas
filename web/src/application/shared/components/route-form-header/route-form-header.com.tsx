import type { ReactNode } from "react";

import { Separator } from "@/components/ui/separator";

export function RouteFormHeader({ title, actions }: Props) {
    return (
        <div className="flex flex-col gap-3 mb-6">
            <div className="flex flex-wrap items-center justify-between gap-3 px-0">
                <div className="flex min-w-0 flex-col gap-3">
                    <h1 className="text-lg font-semibold text-foreground">{title}</h1>
                </div>
                {actions ? <div className="flex items-center gap-2">{actions}</div> : null}
            </div>
            <Separator className="opacity-50" />
        </div>
    );
}

interface Props {
    title: ReactNode;
    actions?: ReactNode;
}

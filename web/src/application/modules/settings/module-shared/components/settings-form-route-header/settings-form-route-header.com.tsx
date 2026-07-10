import type { ReactNode } from "react";

export function SettingsFormRouteHeader({ title }: Props) {
    return (
        <div className="flex flex-wrap items-start justify-between gap-3 border-b pb-3 px-0 mb-4">
            <div className="flex min-w-0 flex-col gap-3">
                <h1 className="text-lg font-semibold text-foreground">{title}</h1>
            </div>
        </div>
    );
}

interface Props {
    title: ReactNode;
}

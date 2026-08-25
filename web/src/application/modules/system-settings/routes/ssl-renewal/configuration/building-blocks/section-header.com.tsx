import type { PropsWithChildren } from "react";

export function SectionHeader({ children }: PropsWithChildren) {
    return (
        <div className="sticky top-0 z-10 rounded-lg bg-accent px-3 py-2 text-sm font-medium shadow-xs">{children}</div>
    );
}

import type { PropsWithChildren } from "react";

export function SectionTitle({ children }: PropsWithChildren) {
    return <div className="rounded-lg bg-accent px-3 py-2 text-sm font-medium shadow-xs">{children}</div>;
}

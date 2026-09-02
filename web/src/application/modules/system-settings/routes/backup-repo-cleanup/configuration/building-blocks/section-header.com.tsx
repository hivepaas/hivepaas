import type { PropsWithChildren } from "react";

export function SectionHeader({ children }: PropsWithChildren) {
    return <div className="text-foreground text-sm font-semibold border-b pb-2">{children}</div>;
}

import type { PropsWithChildren } from "react";

/**
 * The heading above a group of settings rows.
 *
 * One component rather than a copy per page: seven settings forms had declared
 * this same markup locally.
 */
export function SectionHeader({ children }: PropsWithChildren) {
    return (
        <div className="sticky top-0 z-10 rounded-lg bg-accent px-3 py-2 text-sm font-medium shadow-xs">{children}</div>
    );
}

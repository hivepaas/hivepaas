import React, { type PropsWithChildren } from "react";

function View({ children }: PropsWithChildren) {
    return <div className="rounded-lg bg-accent px-3 py-2 text-sm font-medium shadow-xs">{children}</div>;
}

export const SectionHeader = React.memo(View);

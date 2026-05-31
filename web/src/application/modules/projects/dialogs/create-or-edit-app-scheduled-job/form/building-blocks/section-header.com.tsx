import React, { type PropsWithChildren } from "react";

function View({ children }: PropsWithChildren) {
    return <div className="rounded-lg bg-muted px-4 py-3 text-sm font-semibold text-foreground">{children}</div>;
}

export const SectionHeader = React.memo(View);

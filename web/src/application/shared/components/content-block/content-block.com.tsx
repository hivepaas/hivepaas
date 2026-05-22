import React, { type PropsWithChildren } from "react";

function View({ label, children }: Props) {
    return (
        <div className="flex flex-col gap-4">
            <div className="font-medium bg-accent py-2 px-3 rounded-lg">{label}</div>
            <div className="px-4">{children}</div>
        </div>
    );
}

type Props = PropsWithChildren<{
    label: React.ReactNode;
}>;

export const ContentBlock = React.memo(View);

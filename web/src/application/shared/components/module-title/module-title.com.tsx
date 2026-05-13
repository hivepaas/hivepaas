import React, { type PropsWithChildren } from "react";

export function ModuleTitle({ title, children }: Props) {
    return (
        <div className="flex flex-col gap-4">
            <div className="bg-background rounded-lg p-6">
                <div className="flex flex-col gap-2">
                    <h1 className="text-lg font-semibold text-foreground">{title}</h1>
                </div>
            </div>
            {children}
        </div>
    );
}

type Props = PropsWithChildren<{
    title: string;
}>;

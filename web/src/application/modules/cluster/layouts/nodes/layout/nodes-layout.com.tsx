import { type PropsWithChildren } from "react";

import { NodesHeader } from "../header";

export function NodesLayout({ children }: PropsWithChildren) {
    return (
        <div className="flex flex-col gap-5">
            <NodesHeader />
            {children}
        </div>
    );
}

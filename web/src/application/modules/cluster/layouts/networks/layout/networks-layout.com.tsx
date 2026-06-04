import type { PropsWithChildren } from "react";

import { NetworksHeader } from "../header";

export function NetworksLayout({ children }: PropsWithChildren) {
    return (
        <div className="flex flex-col gap-5">
            <NetworksHeader />
            {children}
        </div>
    );
}

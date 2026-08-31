import type { PropsWithChildren } from "react";

import { NetworksHeader } from "../header";

export function NetworksLayout({ children }: PropsWithChildren) {
    return (
        <div className="flex flex-col gap-2 sm:gap-4">
            <NetworksHeader />
            {children}
        </div>
    );
}

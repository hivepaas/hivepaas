import type { PropsWithChildren } from "react";

import { VolumesHeader } from "../header";

export function VolumesLayout({ children }: PropsWithChildren) {
    return (
        <div className="flex flex-col gap-5">
            <VolumesHeader />
            {children}
        </div>
    );
}

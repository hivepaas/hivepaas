import { type PropsWithChildren } from "react";

import { MainHeader } from "../header";

export function MainLayout({ children }: PropsWithChildren) {
    return (
        <div className="flex flex-col gap-2 sm:gap-4">
            <MainHeader />
            {children}
        </div>
    );
}

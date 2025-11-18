import { type PropsWithChildren } from "react";

import { MainHeader } from "../header";

export function MainLayout({ children }: PropsWithChildren) {
    return (
        <div className="flex flex-col gap-5">
            <MainHeader />
            {children}
        </div>
    );
}

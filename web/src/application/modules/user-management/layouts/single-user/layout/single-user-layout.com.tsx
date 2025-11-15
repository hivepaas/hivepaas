import { type PropsWithChildren } from "react";

import { SingleUserHeader } from "../header/single-user-header.com";

export function SingleUserLayout({ children }: PropsWithChildren) {
    return (
        <div className="flex flex-col gap-5">
            <SingleUserHeader />
            {children}
        </div>
    );
}

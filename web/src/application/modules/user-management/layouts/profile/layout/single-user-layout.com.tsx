import { type PropsWithChildren } from "react";

import { ProfileHeader } from "../header";

export function ProfileLayout({ children }: PropsWithChildren) {
    return (
        <div className="flex flex-col gap-5">
            <ProfileHeader />
            {children}
        </div>
    );
}

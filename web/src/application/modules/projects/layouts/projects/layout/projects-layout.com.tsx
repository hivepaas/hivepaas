import { type PropsWithChildren } from "react";

import { ProjectsHeader } from "../header";

export function ProjectsLayout({ children }: PropsWithChildren) {
    return (
        <div className="flex flex-col gap-2 sm:gap-4">
            <ProjectsHeader />
            {children}
        </div>
    );
}

import type React from "react";

import { Lock } from "lucide-react";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

interface Props {
    projectId: string;
    env: string;
    appId: string;
    /** What the engine owns on this screen, as a sentence. */
    children: React.ReactNode;
}

/**
 * Says why a setting cannot be changed: the managed database's engine owns it.
 * The way to change it is to stop managing the database.
 */
export function ManagedLockNotice({ projectId, env, appId, children }: Props) {
    return (
        <div
            role="note"
            className="flex items-start gap-2 rounded-md border border-violet-500/30 bg-violet-500/5 px-3 py-2 text-xs text-violet-800 dark:text-violet-300"
        >
            <Lock className="mt-0.5 size-3.5 shrink-0" />
            <span>
                {children} The managed database&apos;s engine relies on them: to change them, stop managing the database
                on its{" "}
                <AppLink.Basic
                    to={ROUTE.projects.single.apps.single.database.$route(projectId, env, appId)}
                    className="underline"
                >
                    Database tab
                </AppLink.Basic>
                .
            </span>
        </div>
    );
}

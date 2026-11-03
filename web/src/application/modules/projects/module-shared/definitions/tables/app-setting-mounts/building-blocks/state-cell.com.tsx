import React from "react";

import { Badge, statusClassName } from "@components/ui/badge";
import { cn } from "@lib/utils";
import { type AppSettingMount, SETTING_MOUNT_REASONS, type SettingMountReason } from "~/projects/domain";

function reasonText(reason: string): string {
    // A reason this screen does not know yet is shown as the backend says it.
    return reason in SETTING_MOUNT_REASONS ? SETTING_MOUNT_REASONS[reason as SettingMountReason] : reason;
}

function View({ settingMount }: Props) {
    const { state } = settingMount;

    if (!state) {
        return <Badge variant="outline">Unknown</Badge>;
    }

    if (state.reason) {
        return (
            <div className="flex flex-col gap-1">
                <Badge className={cn("w-fit", statusClassName("attention"))}>Not mounted</Badge>
                <span className="text-xs text-muted-foreground">{reasonText(state.reason)}</span>
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-1">
            <Badge className={cn("w-fit", statusClassName("success"))}>Mounted</Badge>
            {state.mounted.map(path => (
                <code
                    key={path}
                    className="break-all text-xs text-muted-foreground"
                >
                    {path}
                </code>
            ))}
        </div>
    );
}

interface Props {
    settingMount: AppSettingMount;
}

export const StateCell = React.memo(View);

import React from "react";

import { Badge } from "@components/ui/badge";
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
                <Badge className="w-fit bg-amber-500 text-white">Not mounted</Badge>
                <span className="text-xs text-muted-foreground">{reasonText(state.reason)}</span>
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-1">
            <Badge className="w-fit bg-green-600 text-white">Mounted</Badge>
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

import { type PropsWithChildren, memo } from "react";

import { Play, Settings } from "lucide-react";
import { SystemSettingsSidebarLayout, type SystemSettingsTabSection } from "~/system-settings/module-shared";

import { ROUTE } from "@application/shared/constants";

const sections: SystemSettingsTabSection[] = [
    {
        title: "Configuration",
        items: [
            {
                label: "Configuration",
                route: ROUTE.systemSettings.backupRepoCleanup.configuration.$route,
                icon: Settings,
            },
            {
                label: "Actions",
                route: ROUTE.systemSettings.backupRepoCleanup.actions.$route,
                icon: Play,
            },
        ],
    },
];

function View({ children }: PropsWithChildren) {
    return <SystemSettingsSidebarLayout sections={sections}>{children}</SystemSettingsSidebarLayout>;
}

export const BackupRepoCleanupLayout = memo(View);

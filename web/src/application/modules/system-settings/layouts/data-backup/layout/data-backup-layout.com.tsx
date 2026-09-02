import { type PropsWithChildren, memo } from "react";

import { HardDrive, Play, Settings } from "lucide-react";
import { SystemSettingsSidebarLayout, type SystemSettingsTabSection } from "~/system-settings/module-shared";

import { ROUTE } from "@application/shared/constants";

const sections: SystemSettingsTabSection[] = [
    {
        title: "Configuration",
        items: [
            {
                label: "Configuration",
                route: ROUTE.systemSettings.dataBackup.configuration.$route,
                icon: Settings,
            },
            {
                label: "Backup Files",
                route: ROUTE.systemSettings.dataBackup.backupFiles.$route,
                icon: HardDrive,
            },
            {
                label: "Actions",
                route: ROUTE.systemSettings.dataBackup.actions.$route,
                icon: Play,
            },
        ],
    },
];

function View({ children }: PropsWithChildren) {
    return <SystemSettingsSidebarLayout sections={sections}>{children}</SystemSettingsSidebarLayout>;
}

export const DataBackupLayout = memo(View);

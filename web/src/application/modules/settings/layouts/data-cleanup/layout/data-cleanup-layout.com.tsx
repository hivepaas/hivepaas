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
                route: ROUTE.appSettings.dataCleanup.configuration.$route,
                icon: Settings,
            },
            {
                label: "Actions",
                route: ROUTE.appSettings.dataCleanup.actions.$route,
                icon: Play,
            },
        ],
    },
];

function View({ children }: PropsWithChildren) {
    return <SystemSettingsSidebarLayout sections={sections}>{children}</SystemSettingsSidebarLayout>;
}

export const DataCleanupLayout = memo(View);

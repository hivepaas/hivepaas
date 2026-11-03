import { type PropsWithChildren, memo } from "react";

import { Activity, Settings } from "lucide-react";
import { SystemSettingsSidebarLayout, type SystemSettingsTabSection } from "~/system-settings/module-shared";

import { ROUTE } from "@application/shared/constants";

const sections: SystemSettingsTabSection[] = [
    {
        title: "Configuration",
        items: [
            {
                label: "Configuration",
                route: ROUTE.systemSettings.logging.configuration.$route,
                icon: Settings,
            },
            {
                label: "Routes and Calls",
                route: ROUTE.systemSettings.logging.performance.$route,
                icon: Activity,
            },
        ],
    },
];

function View({ children }: PropsWithChildren) {
    return <SystemSettingsSidebarLayout sections={sections}>{children}</SystemSettingsSidebarLayout>;
}

export const LoggingLayout = memo(View);

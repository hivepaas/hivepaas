import { type PropsWithChildren, memo } from "react";

import { Settings, SlidersHorizontal } from "lucide-react";
import { SystemSettingsSidebarLayout, type SystemSettingsTabSection } from "~/system-settings/module-shared";

import { ROUTE } from "@application/shared/constants";

const sections: SystemSettingsTabSection[] = [
    {
        title: "Configuration",
        items: [
            {
                label: "General",
                route: ROUTE.systemSettings.traefik.general.$route,
                icon: Settings,
            },
            {
                label: "Config Options",
                route: ROUTE.systemSettings.traefik.configOptions.$route,
                icon: SlidersHorizontal,
            },
        ],
    },
];

function View({ children }: PropsWithChildren) {
    return <SystemSettingsSidebarLayout sections={sections}>{children}</SystemSettingsSidebarLayout>;
}

export const TraefikLayout = memo(View);

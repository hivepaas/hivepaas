import { type PropsWithChildren, memo } from "react";

import { Globe, Play, ScrollText, Settings, ShieldCheck } from "lucide-react";
import { SystemSettingsSidebarLayout, type SystemSettingsTabSection } from "~/system-settings/module-shared";

import { ROUTE } from "@application/shared/constants";

import { useResumeSettingsTrial } from "./use-resume-settings-trial";

const sections: SystemSettingsTabSection[] = [
    {
        title: "Configuration",
        items: [
            {
                label: "General",
                route: ROUTE.systemSettings.hivepaas.general.$route,
                icon: Settings,
            },
            {
                label: "Routing Settings",
                route: ROUTE.systemSettings.hivepaas.routingSettings.$route,
                icon: Globe,
            },
            {
                label: "Security",
                route: ROUTE.systemSettings.hivepaas.security.$route,
                icon: ShieldCheck,
            },
            {
                label: "Logging",
                route: ROUTE.systemSettings.hivepaas.logging.$route,
                icon: ScrollText,
            },
            {
                label: "Actions",
                route: ROUTE.systemSettings.hivepaas.actions.$route,
                icon: Play,
            },
        ],
    },
];

function View({ children }: PropsWithChildren) {
    useResumeSettingsTrial();

    return <SystemSettingsSidebarLayout sections={sections}>{children}</SystemSettingsSidebarLayout>;
}

export const HivePaaSLayout = memo(View);

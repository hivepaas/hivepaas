import { type PropsWithChildren, memo } from "react";

import { BadgeCheck } from "lucide-react";
import { SystemSettingsSidebarLayout, type SystemSettingsTabSection } from "~/system-settings/module-shared";

import { ROUTE } from "@application/shared/constants";

// One tab, as Registry has: the page has room to grow into, and its address
// stays the same when it does.
const sections: SystemSettingsTabSection[] = [
    {
        title: "License",
        items: [
            {
                label: "Overview",
                route: ROUTE.systemSettings.license.overview.$route,
                icon: BadgeCheck,
            },
        ],
    },
];

function View({ children }: PropsWithChildren) {
    return <SystemSettingsSidebarLayout sections={sections}>{children}</SystemSettingsSidebarLayout>;
}

export const LicenseLayout = memo(View);

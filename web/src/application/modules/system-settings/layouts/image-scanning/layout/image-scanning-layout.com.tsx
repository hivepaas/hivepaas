import { type PropsWithChildren, memo } from "react";

import { Boxes, ListChecks, Play, Settings } from "lucide-react";
import { SystemSettingsSidebarLayout, type SystemSettingsTabSection } from "~/system-settings/module-shared";

import { ROUTE } from "@application/shared/constants";

const sections: SystemSettingsTabSection[] = [
    {
        title: "Image Scanning",
        items: [
            {
                label: "Configuration",
                route: ROUTE.systemSettings.imageScanning.configuration.$route,
                icon: Settings,
            },
            {
                label: "Apps",
                route: ROUTE.systemSettings.imageScanning.apps.$route,
                icon: ListChecks,
            },
            {
                label: "HivePaaS",
                route: ROUTE.systemSettings.imageScanning.hivepaas.$route,
                icon: Boxes,
            },
            {
                label: "Actions",
                route: ROUTE.systemSettings.imageScanning.actions.$route,
                icon: Play,
            },
        ],
    },
];

function View({ children }: PropsWithChildren) {
    return <SystemSettingsSidebarLayout sections={sections}>{children}</SystemSettingsSidebarLayout>;
}

export const ImageScanningLayout = memo(View);

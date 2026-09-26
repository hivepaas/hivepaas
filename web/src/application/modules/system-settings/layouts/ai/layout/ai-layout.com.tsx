import { type PropsWithChildren, memo } from "react";

import { Plug } from "lucide-react";
import { SystemSettingsSidebarLayout, type SystemSettingsTabSection } from "~/system-settings/module-shared";

import { ROUTE } from "@application/shared/constants";

// One tab for now; the AI provider an assistant built into the dashboard would
// use joins it later.
const sections: SystemSettingsTabSection[] = [
    {
        title: "AI",
        items: [
            {
                label: "MCP server",
                route: ROUTE.systemSettings.ai.mcp.$route,
                icon: Plug,
            },
        ],
    },
];

function View({ children }: PropsWithChildren) {
    return <SystemSettingsSidebarLayout sections={sections}>{children}</SystemSettingsSidebarLayout>;
}

export const AiLayout = memo(View);

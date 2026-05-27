import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsGithubAppTable } from "~/settings/module-shared/components";

export function SourcesGithubAppsRoute() {
    return (
        <div className={cn(listBox, "min-h-64")}>
            <SettingsGithubAppTable />
        </div>
    );
}

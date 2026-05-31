import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsRepoWebhookTable } from "~/settings/module-shared/components";

export function SourcesWebhooksRoute() {
    return (
        <div className={cn(listBox, "min-h-64")}>
            <SettingsRepoWebhookTable />
        </div>
    );
}

import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { RepoWebhookFormRoute } from "~/settings/module-shared/components/repo-webhook-form-route";

export function SourcesWebhookCreateRoute() {
    return (
        <div className={cn(listBox, "min-h-64")}>
            <RepoWebhookFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}

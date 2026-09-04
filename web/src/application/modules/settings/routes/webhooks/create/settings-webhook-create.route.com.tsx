import { formBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { RepoWebhookFormRoute } from "~/settings/module-shared/components/repo-webhook-form-route";

export function SettingsWebhookCreateRoute() {
    return (
        <div className={cn(formBox, "min-h-64")}>
            <RepoWebhookFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}

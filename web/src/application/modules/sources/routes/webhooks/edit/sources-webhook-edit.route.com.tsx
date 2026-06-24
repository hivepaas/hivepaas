import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { RepoWebhookFormRoute } from "~/settings/module-shared/components/repo-webhook-form-route";

export function SourcesWebhookEditRoute() {
    const { repoWebhookId = "" } = useParams();

    return (
        <div className={cn(listBox, "min-h-64")}>
            <RepoWebhookFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                repoWebhookId={repoWebhookId}
            />
        </div>
    );
}

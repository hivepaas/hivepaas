import { useParams } from "react-router";
import { RepoWebhookFormRoute } from "~/settings/module-shared/components/repo-webhook-form-route";

export function ProjectWebhookEditRoute() {
    const { id: projectId = "", repoWebhookId = "" } = useParams();

    return (
        <RepoWebhookFormRoute
            mode="edit"
            scope={{ type: "project", projectId }}
            repoWebhookId={repoWebhookId}
        />
    );
}

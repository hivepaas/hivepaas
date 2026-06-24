import { useParams } from "react-router";
import { RepoWebhookFormRoute } from "~/settings/module-shared/components/repo-webhook-form-route";

export function ProjectWebhookCreateRoute() {
    const { id: projectId = "" } = useParams();

    return (
        <RepoWebhookFormRoute
            mode="create"
            scope={{ type: "project", projectId }}
        />
    );
}

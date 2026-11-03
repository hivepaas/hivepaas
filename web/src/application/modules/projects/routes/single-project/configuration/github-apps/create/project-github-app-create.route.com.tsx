import { useParams } from "react-router";
import { GithubAppFormRoute } from "~/settings/module-shared/components/github-app-form-route";

export function ProjectGithubAppCreateRoute() {
    const { id: projectId = "" } = useParams();

    return (
        <GithubAppFormRoute
            mode="create"
            scope={{ type: "project", projectId }}
        />
    );
}

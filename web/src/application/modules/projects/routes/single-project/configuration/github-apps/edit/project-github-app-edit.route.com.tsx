import { useParams } from "react-router";
import { GithubAppFormRoute } from "~/settings/module-shared/components/github-app-form-route";

export function ProjectGithubAppEditRoute() {
    const { githubAppId = "", id: projectId = "" } = useParams();

    return (
        <GithubAppFormRoute
            mode="edit"
            scope={{ type: "project", projectId }}
            githubAppId={githubAppId}
        />
    );
}

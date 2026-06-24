import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { GithubAppFormRoute } from "~/settings/module-shared/components/github-app-form-route";

export function SourcesGithubAppEditRoute() {
    const { githubAppId = "" } = useParams();

    return (
        <div className={cn(listBox, "min-h-64")}>
            <GithubAppFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                githubAppId={githubAppId}
            />
        </div>
    );
}

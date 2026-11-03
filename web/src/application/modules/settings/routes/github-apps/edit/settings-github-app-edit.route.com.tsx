import { formBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { GithubAppFormRoute } from "~/settings/module-shared/components/github-app-form-route";

export function SettingsGithubAppEditRoute() {
    const { githubAppId = "" } = useParams();

    return (
        <div className={cn(formBox, "min-h-64")}>
            <GithubAppFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                githubAppId={githubAppId}
            />
        </div>
    );
}

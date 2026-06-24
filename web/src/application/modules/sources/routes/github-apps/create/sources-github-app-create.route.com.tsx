import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { GithubAppFormRoute } from "~/settings/module-shared/components/github-app-form-route";

export function SourcesGithubAppCreateRoute() {
    return (
        <div className={cn(listBox, "min-h-64")}>
            <GithubAppFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}

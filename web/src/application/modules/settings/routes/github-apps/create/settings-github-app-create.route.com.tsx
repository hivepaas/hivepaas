import { formBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { GithubAppFormRoute } from "~/settings/module-shared/components/github-app-form-route";

export function SettingsGithubAppCreateRoute() {
    return (
        <div className={cn(formBox, "min-h-64")}>
            <GithubAppFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}

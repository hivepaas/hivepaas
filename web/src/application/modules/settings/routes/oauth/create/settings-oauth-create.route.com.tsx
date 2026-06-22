import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { OAuthFormRoute } from "~/settings/module-shared/components/oauth-form-route";

export function SettingsOAuthCreateRoute() {
    return (
        <div className={cn(listBox, "p-0")}>
            <OAuthFormRoute mode="create" />
        </div>
    );
}

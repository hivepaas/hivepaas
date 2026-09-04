import { formBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { OAuthFormRoute } from "~/settings/module-shared/components/oauth-form-route";

export function SettingsOAuthCreateRoute() {
    return (
        <div className={cn(formBox)}>
            <OAuthFormRoute mode="create" />
        </div>
    );
}

import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { OAuthFormRoute } from "~/settings/module-shared/components/oauth-form-route";

export function SettingsOAuthEditRoute() {
    const { oauthId = "" } = useParams();

    return (
        <div className={cn(listBox, "p-0")}>
            <OAuthFormRoute
                mode="edit"
                oauthId={oauthId}
            />
        </div>
    );
}

import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { AccessTokenFormRoute } from "~/settings/module-shared/components/access-token-form-route";

export function SettingsAccessTokenEditRoute() {
    const { accessTokenId = "" } = useParams();

    return (
        <div className={cn(listBox, "p-0")}>
            <AccessTokenFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                accessTokenId={accessTokenId}
            />
        </div>
    );
}

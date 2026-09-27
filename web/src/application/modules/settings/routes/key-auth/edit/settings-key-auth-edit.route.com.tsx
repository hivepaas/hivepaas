import { formBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { KeyAuthFormRoute } from "~/settings/module-shared/components/key-auth-form-route";

export function SettingsKeyAuthEditRoute() {
    const { keyAuthId = "" } = useParams();

    return (
        <div className={cn(formBox)}>
            <KeyAuthFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                keyAuthId={keyAuthId}
            />
        </div>
    );
}

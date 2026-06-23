import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { BasicAuthFormRoute } from "~/settings/module-shared/components/basic-auth-form-route";

export function SettingsBasicAuthEditRoute() {
    const { basicAuthId = "" } = useParams();

    return (
        <div className={cn(listBox)}>
            <BasicAuthFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                basicAuthId={basicAuthId}
            />
        </div>
    );
}

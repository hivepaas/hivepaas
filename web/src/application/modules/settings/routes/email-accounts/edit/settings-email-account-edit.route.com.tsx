import { formBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { EmailAccountFormRoute } from "~/settings/module-shared/components/email-account-form-route";

export function SettingsEmailAccountEditRoute() {
    const { emailAccountId = "" } = useParams();

    return (
        <div className={cn(formBox)}>
            <EmailAccountFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                emailAccountId={emailAccountId}
            />
        </div>
    );
}

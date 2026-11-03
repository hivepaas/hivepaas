import { formBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { EmailAccountFormRoute } from "~/settings/module-shared/components/email-account-form-route";

export function SettingsEmailAccountCreateRoute() {
    return (
        <div className={cn(formBox)}>
            <EmailAccountFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}

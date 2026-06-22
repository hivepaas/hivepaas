import { useParams } from "react-router";
import { EmailAccountFormRoute } from "~/settings/module-shared/components/email-account-form-route";

export function ProjectEmailAccountEditRoute() {
    const { emailAccountId = "", id: projectId = "" } = useParams();

    return (
        <EmailAccountFormRoute
            mode="edit"
            scope={{ type: "project", projectId }}
            emailAccountId={emailAccountId}
        />
    );
}

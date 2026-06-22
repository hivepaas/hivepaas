import { useParams } from "react-router";
import { EmailAccountFormRoute } from "~/settings/module-shared/components/email-account-form-route";

export function ProjectEmailAccountCreateRoute() {
    const { id: projectId = "" } = useParams();

    return (
        <EmailAccountFormRoute
            mode="create"
            scope={{ type: "project", projectId }}
        />
    );
}

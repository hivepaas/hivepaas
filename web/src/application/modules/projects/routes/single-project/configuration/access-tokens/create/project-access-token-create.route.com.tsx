import { useParams } from "react-router";
import { AccessTokenFormRoute } from "~/settings/module-shared/components/access-token-form-route";

export function ProjectAccessTokenCreateRoute() {
    const { id: projectId = "" } = useParams();

    return (
        <AccessTokenFormRoute
            mode="create"
            scope={{ type: "project", projectId }}
        />
    );
}

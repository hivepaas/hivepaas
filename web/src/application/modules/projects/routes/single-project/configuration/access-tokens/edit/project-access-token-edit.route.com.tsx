import { useParams } from "react-router";
import { AccessTokenFormRoute } from "~/settings/module-shared/components/access-token-form-route";

export function ProjectAccessTokenEditRoute() {
    const { accessTokenId = "", id: projectId = "" } = useParams();

    return (
        <AccessTokenFormRoute
            mode="edit"
            scope={{ type: "project", projectId }}
            accessTokenId={accessTokenId}
        />
    );
}

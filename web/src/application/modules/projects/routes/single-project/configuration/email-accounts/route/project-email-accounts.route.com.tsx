import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectEmailAccountsTable } from "~/settings/module-shared/components";

import { ProjectProviderSettingsScopeHeader } from "@application/modules/projects/module-shared/components";

export function ProjectEmailAccountsRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return (
        <div className="flex flex-col gap-4">
            <ProjectProviderSettingsScopeHeader projectId={projectId} />
            <ProjectEmailAccountsTable projectId={projectId} />
        </div>
    );
}

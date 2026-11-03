import type { PropsWithChildren } from "react";

import { MODULE_IDS } from "@application/shared/constants";
import { PageNoAccess } from "@application/shared/pages";
import { useConditionalModule, useConditionalProjectCollections } from "@application/shared/permissions";

export function ConditionalProjectsAccess({ children }: PropsWithChildren) {
    const { canRead } = useConditionalModule({ id: MODULE_IDS.Project });
    const { list: projectPermissions } = useConditionalProjectCollections();
    const hasReadableProjectAccess = projectPermissions.some(project => project.actions.read);

    if (!canRead && !hasReadableProjectAccess) {
        return <PageNoAccess />;
    }

    return children;
}

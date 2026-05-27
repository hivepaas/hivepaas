import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditGithubAppDialog } from "~/settings/dialogs/create-or-edit-github-app";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { GithubAppTableScope } from "../../github-app-table.types";

function View({ scope, id, inherited }: Props) {
    const createOrEditDialog = useCreateOrEditGithubAppDialog();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                if (isInheritedProjectSetting(scope, inherited)) {
                    createOrEditDialog.actions.openEdit(scope, id, {
                        props: {
                            readOnlyInherited: true,
                            entityTitle: "Github App",
                        },
                    });
                    return;
                }

                createOrEditDialog.actions.openEdit(scope, id);
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit github app</span>
        </Button>
    );
}

interface Props {
    scope: GithubAppTableScope;
    id: string;
    inherited?: boolean;
}

export const GithubAppEditCell = memo(View);

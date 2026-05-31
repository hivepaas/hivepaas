import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditRepoWebhookDialog } from "~/settings/dialogs/create-or-edit-repo-webhook";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { RepoWebhookTableScope } from "../../repo-webhook-table.types";

function View({ scope, id, inherited }: Props) {
    const createOrEditDialog = useCreateOrEditRepoWebhookDialog();

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
                            entityTitle: "Webhook",
                        },
                    });
                    return;
                }

                createOrEditDialog.actions.openEdit(scope, id);
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit webhook</span>
        </Button>
    );
}

interface Props {
    scope: RepoWebhookTableScope;
    id: string;
    inherited?: boolean;
}

export const RepoWebhookEditCell = memo(View);

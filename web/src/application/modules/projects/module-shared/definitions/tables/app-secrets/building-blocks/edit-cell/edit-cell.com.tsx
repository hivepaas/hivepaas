import React from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditAppSecretDialog } from "~/projects/dialogs/create-or-edit-app-secret/hooks";
import type { AppSecret } from "~/projects/domain";

function View({ projectId, appId, secret }: Props) {
    const { actions: secretDialogActions } = useCreateOrEditAppSecretDialog();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                secretDialogActions.openEdit(projectId, appId, secret);
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit app secret</span>
        </Button>
    );
}

interface Props {
    projectId: string;
    appId: string;
    secret: AppSecret;
}

export const EditCell = React.memo(View);

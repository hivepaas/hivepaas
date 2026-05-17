import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditAccessTokenDialog } from "~/settings/dialogs/create-or-edit-access-token";

import type { AccessTokenTableScope } from "../access-token-table.types";

function View({ scope, id }: Props) {
    const createOrEditDialog = useCreateOrEditAccessTokenDialog();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                createOrEditDialog.actions.openEdit(scope, id);
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit access token</span>
        </Button>
    );
}

interface Props {
    scope: AccessTokenTableScope;
    id: string;
}

export const AccessTokenEditCell = memo(View);

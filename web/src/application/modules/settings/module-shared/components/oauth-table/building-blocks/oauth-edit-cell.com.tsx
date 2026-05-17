import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditOAuthDialog } from "~/settings/dialogs/create-or-edit-oauth";

function View({ id }: Props) {
    const createOrEditDialog = useCreateOrEditOAuthDialog();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                createOrEditDialog.actions.openEdit(id);
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit OAuth</span>
        </Button>
    );
}

interface Props {
    id: string;
}

export const OAuthEditCell = memo(View);

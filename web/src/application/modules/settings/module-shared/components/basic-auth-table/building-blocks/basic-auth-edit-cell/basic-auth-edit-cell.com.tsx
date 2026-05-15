import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditBasicAuthDialog } from "~/settings/dialogs/create-or-edit-basic-auth";

import type { BasicAuthTableScope } from "../../basic-auth-table.types";

function View({ scope, id }: Props) {
    const createOrEditDialog = useCreateOrEditBasicAuthDialog();

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
            <span className="sr-only">Edit basic auth</span>
        </Button>
    );
}

interface Props {
    scope: BasicAuthTableScope;
    id: string;
}

export const BasicAuthEditCell = memo(View);

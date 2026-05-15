import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditEmailAccountDialog } from "~/settings/dialogs/create-or-edit-email-account";

import type { EmailAccountTableScope } from "../../email-account-table.types";

function View({ scope, id }: Props) {
    const createOrEditDialog = useCreateOrEditEmailAccountDialog();

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
            <span className="sr-only">Edit email account</span>
        </Button>
    );
}

interface Props {
    scope: EmailAccountTableScope;
    id: string;
}

export const EmailAccountEditCell = memo(View);

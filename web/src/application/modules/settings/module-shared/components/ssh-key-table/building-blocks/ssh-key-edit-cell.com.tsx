import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditSSHKeyDialog } from "~/settings/dialogs/create-or-edit-ssh-key";

import type { SSHKeyTableScope } from "../ssh-key-table.types";

function View({ scope, id }: Props) {
    const createOrEditDialog = useCreateOrEditSSHKeyDialog();

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
            <span className="sr-only">Edit SSH key</span>
        </Button>
    );
}

interface Props {
    scope: SSHKeyTableScope;
    id: string;
}

export const SSHKeyEditCell = memo(View);

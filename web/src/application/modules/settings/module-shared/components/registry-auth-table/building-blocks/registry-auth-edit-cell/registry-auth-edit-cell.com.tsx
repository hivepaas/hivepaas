import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditRegistryAuthDialog } from "~/settings/dialogs/create-or-edit-registry-auth";

import type { RegistryAuthTableScope } from "../../registry-auth-table.types";

function View({ scope, id }: Props) {
    const createOrEditDialog = useCreateOrEditRegistryAuthDialog();

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
            <span className="sr-only">Edit registry auth</span>
        </Button>
    );
}

interface Props {
    scope: RegistryAuthTableScope;
    id: string;
}

export const RegistryAuthEditCell = memo(View);

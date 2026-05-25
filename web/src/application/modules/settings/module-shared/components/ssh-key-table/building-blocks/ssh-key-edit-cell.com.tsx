import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditSSHKeyDialog } from "~/settings/dialogs/create-or-edit-ssh-key";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { SSHKeyTableScope } from "../ssh-key-table.types";

function View({ scope, id, inherited }: Props) {
    const createOrEditDialog = useCreateOrEditSSHKeyDialog();

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
                            entityTitle: SETTINGS_ENTITY_TITLES.sshKey,
                        },
                    });
                    return;
                }

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
    inherited?: boolean;
}

export const SSHKeyEditCell = memo(View);

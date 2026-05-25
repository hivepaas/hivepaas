import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditAccessTokenDialog } from "~/settings/dialogs/create-or-edit-access-token";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { AccessTokenTableScope } from "../access-token-table.types";

function View({ scope, id, inherited }: Props) {
    const createOrEditDialog = useCreateOrEditAccessTokenDialog();

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
                            entityTitle: SETTINGS_ENTITY_TITLES.accessToken,
                        },
                    });
                    return;
                }

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
    inherited?: boolean;
}

export const AccessTokenEditCell = memo(View);

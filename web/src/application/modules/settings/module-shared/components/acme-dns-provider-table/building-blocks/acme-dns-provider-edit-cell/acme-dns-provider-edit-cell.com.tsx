import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditAcmeDnsProviderDialog } from "~/settings/dialogs/create-or-edit-acme-dns-provider";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { AcmeDnsProviderTableScope } from "../../acme-dns-provider-table.types";

function View({ scope, id, inherited }: Props) {
    const createOrEditDialog = useCreateOrEditAcmeDnsProviderDialog();

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
                            entityTitle: SETTINGS_ENTITY_TITLES.acmeDnsProvider,
                        },
                    });
                    return;
                }

                createOrEditDialog.actions.openEdit(scope, id);
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit ACME DNS provider</span>
        </Button>
    );
}

interface Props {
    scope: AcmeDnsProviderTableScope;
    id: string;
    inherited?: boolean;
}

export const AcmeDnsProviderEditCell = memo(View);

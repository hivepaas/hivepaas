import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditSslCertDialog } from "~/settings/dialogs/create-or-edit-ssl-cert";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { SslCertTableScope } from "../../ssl-cert-table.types";

function View({ scope, id, inherited }: Props) {
    const createOrEditDialog = useCreateOrEditSslCertDialog();

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
                            entityTitle: SETTINGS_ENTITY_TITLES.sslCert,
                        },
                    });
                    return;
                }

                createOrEditDialog.actions.openEdit(scope, id);
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit SSL certificate</span>
        </Button>
    );
}

interface Props {
    scope: SslCertTableScope;
    id: string;
    inherited?: boolean;
}

export const SslCertEditCell = memo(View);

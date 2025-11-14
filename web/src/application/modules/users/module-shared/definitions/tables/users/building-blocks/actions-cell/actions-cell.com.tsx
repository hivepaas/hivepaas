import React from "react";

import { EyeIcon, Trash2Icon } from "lucide-react";

import { AppLink } from "@application/shared/components/navigation";
import { PopConfirm } from "@application/shared/components/pop-confirm";
import { ROUTE } from "@application/shared/constants";

function View({ id }: Props) {
    return (
        <div className="flex items-center justify-center gap-4">
            <AppLink.Modules
                className="flex items-center justify-center text-link hover:opacity-50 transition-opacity duration-200"
                to={ROUTE.modules.userManagement.users.details.$route(id)}
            >
                <EyeIcon className="size-5" />
            </AppLink.Modules>

            <PopConfirm
                title="Remove user"
                variant="destructive"
                confirmText="Remove"
                cancelText="Cancel"
                description="Are you sure you want to remove this user?"
                onConfirm={() => {
                    console.log("Remove user:", id);
                }}
            >
                <Trash2Icon className="size-5 text-destructive hover:opacity-50 transition-opacity duration-200" />
            </PopConfirm>
        </div>
    );
}

interface Props {
    id: string;
}

export const ActionsCell = React.memo(View);

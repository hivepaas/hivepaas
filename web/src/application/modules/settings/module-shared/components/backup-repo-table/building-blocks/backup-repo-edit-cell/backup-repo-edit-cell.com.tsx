import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";

import { useAppNavigate } from "@application/shared/hooks/router";

import type { BackupRepoTableScope } from "../../backup-repo-table.types";

import { getBackupRepoEditRoute } from "./backup-repo-routes.utils";

function View({ scope, id }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(getBackupRepoEditRoute(scope, id));
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit Backup Repo</span>
        </Button>
    );
}

interface Props {
    scope: BackupRepoTableScope;
    id: string;
}

export const BackupRepoEditCell = memo(View);

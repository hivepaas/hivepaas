import { memo, useState } from "react";

import type { SettingGithubApp } from "~/settings/domain";

import { GithubAppSetupIconDialog } from "./github-app-setup-icon.dialog.com";

interface Props {
    githubApp: SettingGithubApp;
}

function View({ githubApp }: Props) {
    const [open, setOpen] = useState(false);

    return (
        <>
            <button
                type="button"
                className="text-sm font-medium text-primary hover:underline cursor-pointer"
                onClick={() => {
                    setOpen(true);
                }}
            >
                Setup
            </button>
            <GithubAppSetupIconDialog
                open={open}
                onOpenChange={setOpen}
                githubApp={githubApp}
            />
        </>
    );
}

export const GithubAppSetupIconCell = memo(View);

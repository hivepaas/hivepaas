import { type PropsWithChildren, useEffect } from "react";

import { useF2aSetupDialogState, useMfaSetupRequiredDialogState } from "@application/shared/dialogs";

import { useAuthContext } from "@application/authentication/context";

import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";

import { ModuleSidebar } from "../module-sidebar";

export function ModuleLayout({ children }: PropsWithChildren) {
    const { data } = useAuthContext();

    useEffect(() => {
        if (!("mfaSetupRequired" in data) || !data.mfaSetupRequired) {
            return;
        }

        const f2aSetupState = useF2aSetupDialogState.getState();
        if (f2aSetupState.state.mode !== "closed") {
            return;
        }

        const introState = useMfaSetupRequiredDialogState.getState();
        if (introState.state.mode !== "closed") {
            return;
        }

        useMfaSetupRequiredDialogState.getState().open();
    }, [data]);

    return (
        <SidebarProvider>
            <ModuleSidebar />
            <SidebarInset>
                <div className="flex flex-1 flex-col gap-4 p-4 bg-[#f5f5f5]">{children}</div>
            </SidebarInset>
        </SidebarProvider>
    );
}

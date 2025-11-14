import { type PropsWithChildren } from "react";

import { useUpdateEffect } from "react-use";

import { useF2aSetupDialog } from "@application/shared/dialogs";

import { useAuthContext } from "@application/authentication/context";

import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";

import { ModuleSidebar } from "../module-sidebar";

export function ModuleLayout({ children }: PropsWithChildren) {
    const f2aSetupDialog = useF2aSetupDialog({
        onClose: () => {
            f2aSetupDialog.actions.close();
        },
    });
    const { data } = useAuthContext();

    useUpdateEffect(() => {
        if ("mfaSetupRequired" in data && data.mfaSetupRequired) {
            f2aSetupDialog.actions.open();
        }
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

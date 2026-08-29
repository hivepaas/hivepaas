import { type PropsWithChildren, useEffect } from "react";

import { LogoIcon } from "@/assets/icons";

import { useF2aSetupDialogState, useMfaSetupRequiredDialogState } from "@application/shared/dialogs";

import { useAuthContext } from "@application/authentication/context";

import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";

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
                {/* Mobile Header: only visible on mobile screens (below md breakpoint) */}
                <header className="flex md:hidden items-center justify-between px-4 py-2.5 border-b bg-background sticky top-0 z-30 shadow-xs">
                    <div className="flex items-center gap-2.5">
                        <SidebarTrigger className="h-8 w-8 text-foreground" />
                        <div className="flex items-center gap-2">
                            <LogoIcon className="x-logo h-7 w-7 text-foreground" />
                            <span className="font-semibold text-sm tracking-tight text-foreground">HivePaaS</span>
                        </div>
                    </div>
                </header>

                <div className="flex flex-1 flex-col gap-2 sm:gap-4 p-2 sm:p-4 bg-[#f5f5f5]">{children}</div>
            </SidebarInset>
        </SidebarProvider>
    );
}

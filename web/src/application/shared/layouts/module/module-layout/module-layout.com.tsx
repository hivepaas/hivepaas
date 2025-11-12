import { type PropsWithChildren } from "react";

import { useUpdateEffect } from "react-use";

import { useF2aSetupDialog } from "@application/shared/dialogs";

import { useAuthContext } from "@application/authentication/context";

import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";

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
                <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
                    <div className="flex items-center gap-2 px-4">
                        <SidebarTrigger className="-ml-1" />
                        <Separator
                            orientation="vertical"
                            className="mr-2 data-[orientation=vertical]:h-4"
                        />
                        <Breadcrumb>
                            <BreadcrumbList>
                                <BreadcrumbItem className="hidden md:block">
                                    <BreadcrumbLink href="#">Building Your Application</BreadcrumbLink>
                                </BreadcrumbItem>
                                <BreadcrumbSeparator className="hidden md:block" />
                                <BreadcrumbItem>
                                    <BreadcrumbPage>Data Fetching</BreadcrumbPage>
                                </BreadcrumbItem>
                            </BreadcrumbList>
                        </Breadcrumb>
                    </div>
                </header>
                <div className="flex flex-1 flex-col gap-4 p-4 pt-0">{children}</div>
            </SidebarInset>
        </SidebarProvider>
    );
}

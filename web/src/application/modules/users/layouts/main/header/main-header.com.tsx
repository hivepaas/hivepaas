import { TabNavigation } from "@application/shared/components/navigation";
import { ROUTE } from "@application/shared/constants";

export function MainHeader() {
    const links = [
        {
            route: ROUTE.userManagement.users.$route,
            label: "Users",
        },
        {
            route: "#",
            label: "Configuration",
        },
    ];
    return (
        <div className="bg-background pt-4 px-5 rounded-lg">
            <div className="py-3 border-b border-border">
                <h1 className="text-2xl font-bold">User management</h1>
            </div>
            <TabNavigation links={links} />
        </div>
    );
}

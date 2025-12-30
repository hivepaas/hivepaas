import { moduleHeaderBox } from "@lib/styles";
import { cn } from "@lib/utils";

export function MainHeader() {
    return (
        <div className={cn(moduleHeaderBox)}>
            <div className="py-2">
                <h1 className="text-lg font-bold">User management</h1>
            </div>
            {/* <TabNavigation links={links} /> */}
        </div>
    );
}

import { moduleHeaderBox } from "@lib/styles";
import { cn } from "@lib/utils";

export function MainHeader() {
    return (
        <div className={cn(moduleHeaderBox)}>
            <div className="py-0.5 sm:py-1.5">
                <h1 className="text-base sm:text-lg font-bold text-foreground flex items-center gap-1.5 tracking-tight">
                    <span className="text-amber-500/70 dark:text-amber-400/70 font-mono font-normal select-none">
                        /
                    </span>
                    <span>User management</span>
                    <span className="text-amber-500/70 dark:text-amber-400/70 font-mono font-normal select-none">
                        /
                    </span>
                </h1>
            </div>
            {/* <TabNavigation links={links} /> */}
        </div>
    );
}

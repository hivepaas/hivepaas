import * as React from "react";

import * as TabsPrimitive from "@radix-ui/react-tabs";
import { cn } from "@/lib/utils";

interface TabPanels {
    /** The values a TabsContent is there for. */
    values: ReadonlySet<string>;
    register: (value: string) => () => void;
}

/**
 * The panels of the tabs. Tabs often have none: they switch what a form shows,
 * or are a page's sections, each a route of its own. A trigger with no panel
 * has no aria-controls - it would name an element that is not there.
 */
const TabPanelsContext = React.createContext<TabPanels | null>(null);

function Tabs({ className, ...props }: React.ComponentProps<typeof TabsPrimitive.Root>) {
    const [values, setValues] = React.useState<ReadonlySet<string>>(new Set());
    const register = React.useCallback((value: string) => {
        setValues(previous => new Set(previous).add(value));

        return () => {
            setValues(previous => {
                const next = new Set(previous);
                next.delete(value);

                return next;
            });
        };
    }, []);
    const panels = React.useMemo(() => ({ values, register }), [values, register]);

    return (
        <TabPanelsContext.Provider value={panels}>
            <TabsPrimitive.Root
                data-slot="tabs"
                className={cn("flex flex-col gap-2", className)}
                {...props}
            />
        </TabPanelsContext.Provider>
    );
}

function TabsList({ className, ...props }: React.ComponentProps<typeof TabsPrimitive.List>) {
    return (
        <TabsPrimitive.List
            data-slot="tabs-list"
            className={cn(
                "bg-muted text-muted-foreground inline-flex h-9 w-fit items-center justify-center rounded-lg p-[3px]",
                className,
            )}
            {...props}
        />
    );
}

function TabsTrigger({ className, ...props }: React.ComponentProps<typeof TabsPrimitive.Trigger>) {
    const panels = React.useContext(TabPanelsContext);
    const hasPanel = panels === null || panels.values.has(props.value);

    return (
        <TabsPrimitive.Trigger
            data-slot="tabs-trigger"
            className={cn(
                "data-[state=active]:bg-background data-[state=active]:text-primary focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:outline-ring dark:data-[state=active]:border-input dark:data-[state=active]:bg-input/30 text-foreground dark:text-muted-foreground inline-flex h-[calc(100%-1px)] flex-1 items-center justify-center gap-1.5 rounded-md border border-transparent px-2 py-1 text-sm font-medium whitespace-nowrap transition-[color,box-shadow] focus-visible:ring-[3px] focus-visible:outline-1 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:shadow-sm [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
                className,
            )}
            {...props}
            {...(hasPanel ? {} : { "aria-controls": undefined })}
        />
    );
}

function TabsContent({ className, ...props }: React.ComponentProps<typeof TabsPrimitive.Content>) {
    const register = React.useContext(TabPanelsContext)?.register;
    React.useLayoutEffect(() => register?.(props.value), [register, props.value]);

    return (
        <TabsPrimitive.Content
            data-slot="tabs-content"
            className={cn("flex-1 outline-none", className)}
            {...props}
        />
    );
}

export { Tabs, TabsList, TabsTrigger, TabsContent };

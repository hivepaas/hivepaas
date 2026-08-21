import { type ReactNode, useMemo, useState } from "react";

import { Button, Input, Tooltip, TooltipContent, TooltipTrigger } from "@components/ui";
import { cn } from "@lib/utils";
import { Pencil } from "lucide-react";
import { toast } from "sonner";
import { ProjectCommandTemplateCommands } from "~/projects/data/commands";
import { ProjectCommandTemplateQueries } from "~/projects/data/queries";
import type { ProjectCommandTemplate } from "~/projects/domain";

import { ROUTE } from "@application/shared/constants";

const COMMAND_TEMPLATE_LIMIT = 10_000;

export function AppTerminalCommandTemplatePanel({
    projectID,
    env,
    appID,
    isConnected,
    onInsertCommand,
}: AppTerminalCommandTemplatePanelProps) {
    const [search, setSearch] = useState("");
    const [buildingTemplateId, setBuildingTemplateId] = useState<string | null>(null);
    const commandTemplatesQuery = ProjectCommandTemplateQueries.useFindManyEnvPaginated({
        projectID,
        env,
        pagination: {
            page: 1,
            size: COMMAND_TEMPLATE_LIMIT,
        },
    });
    const { mutateAsync: buildForApp } = ProjectCommandTemplateCommands.useBuildForApp();

    const groups = useMemo(() => {
        return groupCommandTemplates(commandTemplatesQuery.data?.data ?? [], search);
    }, [commandTemplatesQuery.data?.data, search]);

    const hasTemplates = groups.length > 0;

    async function handleSelect(commandTemplate: ProjectCommandTemplate) {
        if (!isConnected) {
            toast.error("Connect terminal before inserting a command");
            return;
        }

        if (buildingTemplateId !== null) {
            return;
        }

        setBuildingTemplateId(commandTemplate.id);

        try {
            const response = await buildForApp({
                projectID,
                env,
                appID,
                id: commandTemplate.id,
            });
            onInsertCommand(stripCommandTerminator(response.data.command));
        } catch {
            // Error toast is handled by useProjectCommandTemplateApi.
        } finally {
            setBuildingTemplateId(null);
        }
    }

    function handleOpenEdit(commandTemplate: ProjectCommandTemplate) {
        window.open(
            ROUTE.projects.single.providerConfiguration.commandTemplates.edit.$route(projectID, commandTemplate.id),
            "_blank",
            "noopener,noreferrer",
        );
    }

    return (
        <aside className="flex max-h-full w-full max-w-[300px] shrink-0 flex-col rounded-md border bg-background p-3 shadow-sm">
            <Input
                value={search}
                placeholder="Search"
                className="mb-3 h-9"
                onChange={event => {
                    setSearch(event.target.value);
                }}
            />

            <div className="min-h-0 flex-1 overflow-y-auto pr-1">
                {commandTemplatesQuery.isLoading ? (
                    <PanelState>Loading command templates...</PanelState>
                ) : commandTemplatesQuery.isError ? (
                    <PanelState>Failed to load command templates.</PanelState>
                ) : !hasTemplates ? (
                    <PanelState>No command templates found.</PanelState>
                ) : (
                    <div className="flex flex-col gap-3">
                        {groups.map(group => (
                            <section
                                key={group.kind}
                                className="flex flex-col gap-1"
                            >
                                <h3 className="px-2 text-sm font-medium text-foreground">{group.kind}</h3>
                                <div className="flex flex-col gap-0.5">
                                    {group.items.map(commandTemplate => {
                                        const isBuilding = buildingTemplateId === commandTemplate.id;

                                        return (
                                            <div
                                                key={commandTemplate.id}
                                                className={cn(
                                                    "group flex min-h-9 items-center gap-1 rounded-md",
                                                    "hover:bg-accent hover:text-accent-foreground",
                                                )}
                                            >
                                                <button
                                                    type="button"
                                                    disabled={buildingTemplateId !== null}
                                                    title={commandTemplate.name}
                                                    className={cn(
                                                        "min-w-0 flex-1 rounded-md px-5 py-2 text-left text-sm",
                                                        "disabled:pointer-events-none disabled:opacity-60",
                                                    )}
                                                    onClick={() => {
                                                        void handleSelect(commandTemplate);
                                                    }}
                                                >
                                                    <span className="block truncate">
                                                        {isBuilding ? "Building..." : commandTemplate.name}
                                                    </span>
                                                </button>
                                                <Tooltip>
                                                    <TooltipTrigger asChild>
                                                        <Button
                                                            type="button"
                                                            variant="ghost"
                                                            size="icon-sm"
                                                            aria-label={`Edit ${commandTemplate.name}`}
                                                            className="mr-1 text-primary"
                                                            onClick={() => {
                                                                handleOpenEdit(commandTemplate);
                                                            }}
                                                        >
                                                            <Pencil className="size-4" />
                                                        </Button>
                                                    </TooltipTrigger>
                                                    <TooltipContent>Edit command template</TooltipContent>
                                                </Tooltip>
                                            </div>
                                        );
                                    })}
                                </div>
                            </section>
                        ))}
                    </div>
                )}
            </div>
        </aside>
    );
}

function PanelState({ children }: PanelStateProps) {
    return <div className="px-2 py-8 text-center text-sm text-muted-foreground">{children}</div>;
}

function groupCommandTemplates(commandTemplates: ProjectCommandTemplate[], search: string): CommandTemplateGroup[] {
    const normalizedSearch = search.trim().toLowerCase();

    const filtered = commandTemplates.filter(commandTemplate => {
        if (!normalizedSearch) {
            return true;
        }

        return [commandTemplate.kind, commandTemplate.name].some(value => value.toLowerCase().includes(normalizedSearch));
    });

    const sorted = [...filtered].sort((left, right) => {
        const kindSort = compareText(left.kind, right.kind);

        if (kindSort !== 0) {
            return kindSort;
        }

        return compareText(left.name, right.name);
    });

    const groups = new Map<string, ProjectCommandTemplate[]>();

    for (const commandTemplate of sorted) {
        const kind = commandTemplate.kind || "Other";
        groups.set(kind, [...(groups.get(kind) ?? []), commandTemplate]);
    }

    return Array.from(groups.entries()).map(([kind, items]) => ({
        kind,
        items,
    }));
}

function compareText(left: string, right: string): number {
    return left.localeCompare(right, undefined, {
        sensitivity: "base",
    });
}

function stripCommandTerminator(command: string): string {
    return command.replace(/[\r\n]+$/u, "");
}

interface AppTerminalCommandTemplatePanelProps {
    projectID: string;
    env: string;
    appID: string;
    isConnected: boolean;
    onInsertCommand: (command: string) => void;
}

interface PanelStateProps {
    children: ReactNode;
}

interface CommandTemplateGroup {
    kind: string;
    items: ProjectCommandTemplate[];
}

"use client";

import { useState } from "react";

import { cn } from "@/lib/utils";
import { Plus } from "lucide-react";
import { toast } from "sonner";
import { ProjectCommandTemplateCommands } from "~/projects/data/commands";
import { SettingsScopePermissionAction } from "~/settings/module-shared/components";

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

import {
    COMMAND_TEMPLATE_FROM_TEMPLATE_CATALOG,
    type CommandTemplateFromTemplateSelection,
} from "./command-template-from-template.catalog";

const PROJECT_SCOPE = { type: "project" } as const;

export function ProjectCommandTemplateFromTemplateMenu({ projectId }: Props) {
    const [open, setOpen] = useState(false);

    const { mutate: createFromTemplate, isPending } = ProjectCommandTemplateCommands.useCreateFromTemplate({
        onSuccess: () => {
            toast.success("Project Command Template created successfully");
            setOpen(false);
        },
    });

    function handleSelect(selection: CommandTemplateFromTemplateSelection) {
        createFromTemplate({
            projectID: projectId,
            payload: {
                commandType: selection.commandType,
                commandKind: selection.commandKind,
            },
        });
    }

    return (
        <SettingsScopePermissionAction
            scope={PROJECT_SCOPE}
            action="write"
        >
            {({ isDenied }) => (
                <Popover
                    open={open}
                    onOpenChange={setOpen}
                >
                    <PopoverTrigger asChild>
                        <Button
                            type="button"
                            disabled={isDenied || isPending}
                        >
                            <Plus className="size-4" />
                            New From Template
                        </Button>
                    </PopoverTrigger>
                    <PopoverContent
                        className="w-[320px] p-2"
                        align="start"
                    >
                        <Accordion
                            type="single"
                            collapsible
                            className="w-full"
                        >
                            {COMMAND_TEMPLATE_FROM_TEMPLATE_CATALOG.map(group => (
                                <AccordionItem
                                    key={group.commandType}
                                    value={group.commandType}
                                    className="border-b-0"
                                >
                                    <AccordionTrigger className="px-2 py-2 text-sm font-medium hover:no-underline">
                                        {group.heading}
                                    </AccordionTrigger>
                                    <AccordionContent className="pb-1 pt-0">
                                        <div className="flex max-h-[280px] flex-col overflow-y-auto">
                                            {group.items.map(item => (
                                                <button
                                                    key={`${item.commandType}:${item.commandKind}`}
                                                    type="button"
                                                    disabled={isPending}
                                                    className={cn(
                                                        "rounded-sm pr-4 py-1.5 pl-8 text-left text-sm",
                                                        "hover:bg-accent hover:text-accent-foreground",
                                                        "disabled:pointer-events-none disabled:opacity-50",
                                                    )}
                                                    onClick={() => {
                                                        handleSelect({
                                                            commandType: item.commandType,
                                                            commandKind: item.commandKind,
                                                        });
                                                    }}
                                                >
                                                    {item.label}
                                                </button>
                                            ))}
                                        </div>
                                    </AccordionContent>
                                </AccordionItem>
                            ))}
                        </Accordion>
                    </PopoverContent>
                </Popover>
            )}
        </SettingsScopePermissionAction>
    );
}

interface Props {
    projectId: string;
}

"use client";

import { useState } from "react";

import { Plus } from "lucide-react";
import { toast } from "sonner";
import { ProjectCommandTemplateCommands } from "~/projects/data/commands";
import { SettingsScopePermissionAction } from "~/settings/module-shared/components";

import { Button } from "@/components/ui/button";
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command";
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
                        className="w-[320px] p-0"
                        align="start"
                    >
                        <Command>
                            <CommandInput placeholder="Search template..." />
                            <CommandList className="max-h-[400px]">
                                <CommandEmpty>No template found.</CommandEmpty>
                                {COMMAND_TEMPLATE_FROM_TEMPLATE_CATALOG.map(group => (
                                    <CommandGroup
                                        key={group.commandType}
                                        heading={group.heading}
                                    >
                                        {group.items.map(item => (
                                            <CommandItem
                                                key={`${item.commandType}:${item.commandKind}`}
                                                value={`${group.heading} ${item.label} ${item.commandKind}`}
                                                disabled={isPending}
                                                className="cursor-pointer ml-2"
                                                onSelect={() => {
                                                    handleSelect({
                                                        commandType: item.commandType,
                                                        commandKind: item.commandKind,
                                                    });
                                                }}
                                            >
                                                {item.label}
                                            </CommandItem>
                                        ))}
                                    </CommandGroup>
                                ))}
                            </CommandList>
                        </Command>
                    </PopoverContent>
                </Popover>
            )}
        </SettingsScopePermissionAction>
    );
}

interface Props {
    projectId: string;
}

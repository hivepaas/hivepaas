import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical } from "lucide-react";

import { CopyIdMenuButton } from "@application/shared/components";

function View({ id }: Props) {
    const [open, setOpen] = useState(false);

    return (
        <DropdownMenu
            open={open}
            onOpenChange={setOpen}
        >
            <DropdownMenuTrigger asChild>
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                >
                    <MoreVertical className="size-4" />
                    <span className="sr-only">Actions menu</span>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <div className="flex flex-col gap-0">
                    <CopyIdMenuButton
                        id={id}
                        onCopied={() => {
                            setOpen(false);
                        }}
                    />
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    id: string;
}

export const MenuCell = React.memo(View);

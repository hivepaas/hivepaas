import { memo } from "react";

import { Button } from "@components/ui";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { ChevronDown, ExternalLink, Globe } from "lucide-react";

function View({ accessLinks }: Props) {
    if (accessLinks.length === 0) {
        return null;
    }

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button
                    type="button"
                    variant="ghost"
                    className="h-12 gap-1 rounded-none border-0 border-b-2 border-transparent px-3 hover:bg-transparent"
                >
                    <Globe className="size-4 text-blue-500" />
                    <ChevronDown className="size-3 font-semibold text-foreground" />
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
                align="end"
                className="min-w-[240px]"
            >
                {accessLinks.map(url => (
                    <DropdownMenuItem
                        key={url}
                        asChild
                    >
                        <a
                            href={url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="flex items-center gap-2"
                        >
                            <ExternalLink className="size-4 shrink-0 text-primary" />
                            <span className="truncate text-primary">{url}</span>
                        </a>
                    </DropdownMenuItem>
                ))}
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    accessLinks: string[];
}

export const AppAccessLinksDropdown = memo(View);

import { useMemo, useState } from "react";

import { Input } from "@components/ui/input";
import { Textarea } from "@components/ui/textarea";
import { SearchIcon } from "lucide-react";

import { AppLoader } from "@application/shared/components";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@/components/ui";

export type FinalEnvValueItem = {
    key: string;
    value: string;
};

type Props = {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    items: FinalEnvValueItem[];
    sectionTitle: string;
    isPending?: boolean;
};

export function FinalEnvValuesDialog({ open, onOpenChange, items, sectionTitle, isPending = false }: Props) {
    const [search, setSearch] = useState("");

    const filteredItems = useMemo(() => {
        const query = search.trim().toLowerCase();
        if (!query) {
            return items;
        }

        return items.filter(item => {
            return item.key.toLowerCase().includes(query) || item.value.toLowerCase().includes(query);
        });
    }, [items, search]);

    return (
        <Dialog
            open={open}
            onOpenChange={nextOpen => {
                if (!nextOpen) {
                    setSearch("");
                }
                onOpenChange(nextOpen);
            }}
        >
            <DialogFixedContent className="flex h-[80vh] w-[1000px] max-w-[calc(100vw-2rem)] max-h-[calc(100vh-2rem)] flex-col">
                <DialogHeader>
                    <DialogTitle>Final Env Values</DialogTitle>
                </DialogHeader>

                <DialogBody className="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden">
                    {isPending ? (
                        <div className="flex flex-1 items-center justify-center">
                            <AppLoader />
                        </div>
                    ) : (
                        <>
                            <div className="relative shrink-0">
                                <div className="text-muted-foreground pointer-events-none absolute inset-y-0 left-0 flex items-center justify-center pl-3">
                                    <SearchIcon className="size-4" />
                                    <span className="sr-only">Search</span>
                                </div>
                                <Input
                                    value={search}
                                    onChange={e => {
                                        setSearch(e.target.value);
                                    }}
                                    type="search"
                                    placeholder="Search"
                                    className="peer px-9 [&::-webkit-search-cancel-button]:appearance-none [&::-webkit-search-decoration]:appearance-none [&::-webkit-search-results-button]:appearance-none [&::-webkit-search-results-decoration]:appearance-none"
                                />
                            </div>

                            <div className="bg-accent shrink-0 rounded-lg px-3 py-2 text-sm font-medium">
                                {sectionTitle}
                            </div>

                            <div className="min-h-0 flex-1 space-y-3 overflow-y-auto pr-1">
                                {filteredItems.length > 0 ? (
                                    filteredItems.map(item => (
                                        <div
                                            key={item.key}
                                            className="flex items-start gap-3"
                                        >
                                            <Input
                                                value={item.key}
                                                readOnly
                                                disabled
                                                className="min-w-0 flex-1 bg-muted cursor-default"
                                            />
                                            <Textarea
                                                value={item.value}
                                                readOnly
                                                disabled
                                                minRows={1}
                                                maxRows={0}
                                                className="min-w-0 flex-1 resize-y bg-muted cursor-default"
                                            />
                                        </div>
                                    ))
                                ) : (
                                    <div className="text-sm text-muted-foreground py-8 text-center">
                                        {search.trim() ? "No records match your search" : "No results"}
                                    </div>
                                )}
                            </div>
                        </>
                    )}
                </DialogBody>
            </DialogFixedContent>
        </Dialog>
    );
}

import { useEffect, useState } from "react";

import { Play } from "lucide-react";
import { ProjectCommandTemplateQueries } from "~/projects/data/queries";
import type { ProjectCommandTemplate } from "~/projects/domain";
import { SelectorSearch } from "~/projects/module-shared/components/selector-dialog";
import { GIT_SELECTOR_PAGE_SIZE, appendUniqueByKey, createSelectorPagination } from "~/projects/module-shared/utils";

import {
    Button,
    Dialog,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui";

interface Props {
    open: boolean;
    projectId: string;
    onClose: () => void;
    onSelect: (template: ProjectCommandTemplate) => void;
}

export function CommandTemplatePicker({ open, projectId, onClose, onSelect }: Props) {
    const [search, setSearch] = useState("");
    const [page, setPage] = useState(1);
    const [rows, setRows] = useState<ProjectCommandTemplate[]>([]);

    const query = ProjectCommandTemplateQueries.useFindManyPaginated(
        {
            projectID: projectId,
            search,
            pagination: createSelectorPagination(page),
        },
        {
            enabled: open && Boolean(projectId),
            placeholderData: undefined,
        },
    );

    useEffect(() => {
        if (!open) {
            return;
        }

        setSearch("");
        setPage(1);
        setRows([]);
    }, [open, projectId]);

    useEffect(() => {
        setPage(1);
        setRows([]);
    }, [search]);

    useEffect(() => {
        const nextRows = query.data?.data;

        if (!nextRows) {
            return;
        }

        setRows(current => (page === 1 ? nextRows : appendUniqueByKey(current, nextRows, row => row.id)));
    }, [page, query.dataUpdatedAt, query.data?.data]);

    const canLoadMore = (query.data?.data.length ?? 0) >= GIT_SELECTOR_PAGE_SIZE;

    return (
        <Dialog
            open={open}
            onOpenChange={onClose}
        >
            <DialogFixedContent className="h-[800px] w-[800px] max-w-[calc(100vw-2rem)] max-h-[calc(100vh-2rem)]">
                <DialogHeader>
                    <DialogTitle>Command Templates</DialogTitle>
                </DialogHeader>
                <DialogBody className="flex min-h-0 flex-col gap-5">
                    <SelectorSearch
                        value={search}
                        onChange={setSearch}
                    />

                    <div className="min-h-0 overflow-auto rounded-md border">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead className="w-[120px] text-center">Action</TableHead>
                                    <TableHead>Command Template</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {query.isFetching && rows.length === 0 ? (
                                    <TableRow>
                                        <TableCell
                                            colSpan={2}
                                            className="h-24 text-center text-muted-foreground"
                                        >
                                            Loading...
                                        </TableCell>
                                    </TableRow>
                                ) : rows.length === 0 ? (
                                    <TableRow>
                                        <TableCell
                                            colSpan={2}
                                            className="h-24 text-center text-muted-foreground"
                                        >
                                            No command templates found.
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    rows.map(row => (
                                        <TableRow key={row.id}>
                                            <TableCell className="text-center">
                                                <Button
                                                    type="button"
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() => {
                                                        onSelect(row);
                                                        onClose();
                                                    }}
                                                >
                                                    <Play className="size-4" /> Select
                                                </Button>
                                            </TableCell>
                                            <TableCell>
                                                <span
                                                    className="block max-w-[640px] truncate"
                                                    title={row.name}
                                                >
                                                    {row.name}
                                                </span>
                                            </TableCell>
                                        </TableRow>
                                    ))
                                )}
                            </TableBody>
                        </Table>
                    </div>

                    {canLoadMore && (
                        <div className="shrink-0">
                            <Button
                                type="button"
                                variant="outline"
                                disabled={query.isFetching}
                                onClick={() => {
                                    setPage(current => current + 1);
                                }}
                            >
                                Load More
                            </Button>
                        </div>
                    )}
                </DialogBody>
            </DialogFixedContent>
        </Dialog>
    );
}

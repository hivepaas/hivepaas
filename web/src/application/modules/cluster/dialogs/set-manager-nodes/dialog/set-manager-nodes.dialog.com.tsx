import { useEffect, useMemo, useRef, useState } from "react";

import { cn } from "@/lib/utils";
import { dashedBorderBox } from "@lib/styles";
import { Search } from "lucide-react";
import { toast } from "sonner";
import { NodesCommands, NodesQueries } from "~/cluster/data";
import type { NodeDetails } from "~/cluster/domain";
import { NodeAvailabilityBadge, NodeRoleBadge } from "~/cluster/module-shared/components";
import { ENodeRole } from "~/cluster/module-shared/enums";
import {
    SET_MANAGER_NODES_PAGE_SIZE,
    appendUniqueByKey,
    createSelectorPagination,
} from "~/cluster/module-shared/utils";

import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import {
    Button,
    Checkbox,
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
    Input,
    Separator,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui";

import { useSetManagerNodesDialogState } from "../hooks";

const MANAGER_SEED_PAGE_SIZE = 1000;

function areSetsEqual(left: Set<string>, right: Set<string>): boolean {
    if (left.size !== right.size) {
        return false;
    }

    for (const id of left) {
        if (!right.has(id)) {
            return false;
        }
    }

    return true;
}

function getManagerIds(nodes: NodeDetails[]): Set<string> {
    return new Set(nodes.filter(node => node.role === ENodeRole.Manager).map(node => node.id));
}

export function SetManagerNodesDialog() {
    const { state, ...actions } = useSetManagerNodesDialogState();
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Cluster });

    const open = state.mode !== "closed";

    const [search, setSearch] = useState("");
    const [page, setPage] = useState(1);
    const [rows, setRows] = useState<NodeDetails[]>([]);
    const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
    const [initialManagerIds, setInitialManagerIds] = useState<Set<string>>(new Set());

    const managersSeedQuery = NodesQueries.useFindManyPaginated(
        {
            pagination: {
                page: 1,
                size: MANAGER_SEED_PAGE_SIZE,
            },
        },
        {
            enabled: open,
            placeholderData: undefined,
        },
    );

    const nodesQuery = NodesQueries.useFindManyPaginated(
        {
            search,
            pagination: createSelectorPagination(page),
        },
        {
            enabled: open,
            placeholderData: undefined,
        },
    );

    const { mutate: setManagers, isPending } = NodesCommands.useSetManagers({
        onSuccess: () => {
            toast.success("Manager nodes updated successfully");
            actions.close();
        },
    });

    const hasChanges = useMemo(() => !areSetsEqual(selectedIds, initialManagerIds), [initialManagerIds, selectedIds]);

    useEffect(() => {
        if (!open) {
            return;
        }

        setSearch("");
        setPage(1);
        setRows([]);
        setSelectedIds(new Set());
        setInitialManagerIds(new Set());
    }, [open]);

    useEffect(() => {
        if (!open || !managersSeedQuery.data) {
            return;
        }

        const managerIds = getManagerIds(managersSeedQuery.data.data);

        setSelectedIds(new Set(managerIds));
        setInitialManagerIds(new Set(managerIds));
    }, [managersSeedQuery.data, managersSeedQuery.dataUpdatedAt, open]);

    useEffect(() => {
        setPage(1);
        setRows([]);
    }, [search]);

    useEffect(() => {
        const nextRows = nodesQuery.data?.data;

        if (!nextRows) {
            return;
        }

        setRows(current => (page === 1 ? nextRows : appendUniqueByKey(current, nextRows, row => row.id)));
    }, [nodesQuery.data?.data, nodesQuery.dataUpdatedAt, page]);

    const canLoadMore = (nodesQuery.data?.data.length ?? 0) >= SET_MANAGER_NODES_PAGE_SIZE;

    function handleClose(): void {
        if (canWrite && hasChanges) {
            const userConfirmed = window.confirm("Are you sure you want to close without saving changes?");

            if (!userConfirmed) {
                return;
            }
        }

        actions.close();
    }

    function handleToggleNode(nodeId: string, checked: boolean): void {
        if (!canWrite) {
            return;
        }

        setSelectedIds(current => {
            const next = new Set(current);

            if (checked) {
                next.add(nodeId);
            } else {
                next.delete(nodeId);
            }

            return next;
        });
    }

    function handleSave(): void {
        if (!canWrite || selectedIds.size === 0) {
            return;
        }

        setManagers({
            nodes: Array.from(selectedIds).map(id => ({ id })),
        });
    }

    const searchInputRef = useRef<HTMLInputElement>(null);

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogFixedContent
                className="h-[800px] w-[1000px] max-w-[calc(100vw-1rem)] max-h-[calc(100svh-2.5rem)] sm:max-h-[85vh]"
                onOpenAutoFocus={event => {
                    event.preventDefault();
                    searchInputRef.current?.focus();
                }}
            >
                <DialogHeader>
                    <DialogTitle>Set manager nodes</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                <DialogBody className="flex min-h-0 flex-1 flex-col gap-4">
                    <div className={cn(dashedBorderBox, "text-sm")}>
                        <span className="font-semibold text-orange-500">Notice:</span> A 3-node cluster can tolerate the
                        loss of 1 manager, but a 4-node cluster can ALSO only tolerate the loss of 1 manager. So having
                        even number of managers does not make sense in terms of HA, and causes unnecessary resource
                        consumption without any reliability benefit.
                    </div>

                    <div className="relative w-full max-w-[320px]">
                        <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center justify-center pl-3 text-muted-foreground">
                            <Search className="size-4" />
                            <span className="sr-only">Search</span>
                        </div>
                        <Input
                            ref={searchInputRef}
                            value={search}
                            type="search"
                            placeholder="Search"
                            className="px-9 [&::-webkit-search-cancel-button]:appearance-none [&::-webkit-search-decoration]:appearance-none [&::-webkit-search-results-button]:appearance-none [&::-webkit-search-results-decoration]:appearance-none"
                            onChange={event => {
                                setSearch(event.target.value);
                            }}
                        />
                    </div>

                    <div className="min-h-0 flex-1 overflow-auto rounded-md border">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead className="w-[48px]" />
                                    <TableHead>Name</TableHead>
                                    <TableHead>Hostname</TableHead>
                                    <TableHead>Address</TableHead>
                                    <TableHead>Role</TableHead>
                                    <TableHead>Availability</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {nodesQuery.isFetching && rows.length === 0 ? (
                                    <TableRow>
                                        <TableCell
                                            colSpan={6}
                                            className="h-24 text-center text-muted-foreground"
                                        >
                                            Loading...
                                        </TableCell>
                                    </TableRow>
                                ) : rows.length === 0 ? (
                                    <TableRow>
                                        <TableCell
                                            colSpan={6}
                                            className="h-24 text-center text-muted-foreground"
                                        >
                                            No nodes found.
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    rows.map(row => (
                                        <TableRow key={row.id}>
                                            <TableCell>
                                                <Checkbox
                                                    checked={selectedIds.has(row.id)}
                                                    disabled={!canWrite || isPending}
                                                    onCheckedChange={checked => {
                                                        handleToggleNode(row.id, checked === true);
                                                    }}
                                                />
                                            </TableCell>
                                            <TableCell>{row.name}</TableCell>
                                            <TableCell>{row.hostname}</TableCell>
                                            <TableCell>{row.addr}</TableCell>
                                            <TableCell>
                                                <NodeRoleBadge
                                                    role={row.role}
                                                    isLeader={row.isLeader}
                                                />
                                            </TableCell>
                                            <TableCell>
                                                <NodeAvailabilityBadge availability={row.availability} />
                                            </TableCell>
                                        </TableRow>
                                    ))
                                )}
                            </TableBody>
                        </Table>
                    </div>
                </DialogBody>
                <DialogActionFooter className="flex items-center justify-between">
                    <div>
                        {canLoadMore ? (
                            <Button
                                type="button"
                                variant="outline"
                                disabled={nodesQuery.isFetching}
                                onClick={() => {
                                    setPage(current => current + 1);
                                }}
                            >
                                Load More
                            </Button>
                        ) : null}
                    </div>
                    <Button
                        type="button"
                        className="min-w-[100px]"
                        disabled={!canWrite || selectedIds.size === 0 || isPending}
                        isLoading={isPending}
                        onClick={handleSave}
                    >
                        Save
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}

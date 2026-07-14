import { useEffect, useMemo, useState } from "react";

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

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="h-[800px] w-[1000px] max-w-[calc(100vw-2rem)] max-h-[calc(100vh-2rem)]">
                <DialogHeader>
                    <DialogTitle>Set Manager Nodes</DialogTitle>
                </DialogHeader>
                <DialogBody className="flex min-h-0 flex-col gap-5">
                    <div className={cn(dashedBorderBox, "text-sm leading-6")}>
                        <span className="text-orange-500">Note:</span> Docker Swarm uses the Raft consensus algorithm.
                        It is highly recommended to configure an odd number of managers (e.g., 1, 3 or 5) to optimize
                        fault tolerance and avoid split-brain scenarios. Configuring an even number of managers (e.g.,
                        2, 4 or 6) does not increase the cluster&apos;s fault tolerance compared to the odd number below
                        it (1, 3 or 5). It only increases synchronization overhead and resource consumption without any
                        reliability benefit.
                    </div>

                    <div className="relative w-full max-w-[320px]">
                        <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center justify-center pl-3 text-muted-foreground">
                            <Search className="size-4" />
                            <span className="sr-only">Search</span>
                        </div>
                        <Input
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

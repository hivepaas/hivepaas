import { useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { Play, Search } from "lucide-react";
import type { ProjectGitBranch, ProjectGitPullRequest, ProjectGitRepo } from "~/projects/api/services";
import { ProjectGitCredentialsQueries } from "~/projects/data";
import { GIT_SELECTOR_PAGE_SIZE, type ParsedGitRepository, truncateSha } from "~/projects/module-shared/utils";

import {
    Button,
    Dialog,
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

type SelectorPagination = {
    page: number;
    size: number;
};

function createPagination(page: number): SelectorPagination {
    return {
        page,
        size: GIT_SELECTOR_PAGE_SIZE,
    };
}

function includesSearch(value: string | number | null | undefined, search: string): boolean {
    return String(value ?? "")
        .toLowerCase()
        .includes(search);
}

function appendUniqueByKey<T>(current: T[], next: T[], getKey: (item: T) => string): T[] {
    const keys = new Set(current.map(getKey));
    const merged = [...current];

    for (const item of next) {
        const key = getKey(item);

        if (keys.has(key)) {
            continue;
        }

        keys.add(key);
        merged.push(item);
    }

    return merged;
}

function SelectorSearch({ value, onChange }: { value: string; onChange: (value: string) => void }) {
    return (
        <div className="relative w-full max-w-[320px]">
            <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center justify-center pl-3 text-muted-foreground">
                <Search className="size-4" />
                <span className="sr-only">Search</span>
            </div>
            <Input
                value={value}
                type="search"
                placeholder="Search"
                className="px-9 [&::-webkit-search-cancel-button]:appearance-none [&::-webkit-search-decoration]:appearance-none [&::-webkit-search-results-button]:appearance-none [&::-webkit-search-results-decoration]:appearance-none"
                onChange={event => {
                    onChange(event.target.value);
                }}
            />
        </div>
    );
}

export function GitRepositoriesDialog({
    open,
    onOpenChange,
    projectId,
    credentialId,
    onSelect,
}: GitRepositoriesDialogProps) {
    const [search, setSearch] = useState("");
    const [page, setPage] = useState(1);
    const [rows, setRows] = useState<ProjectGitRepo[]>([]);

    const query = ProjectGitCredentialsQueries.useFindManyRepos(
        {
            projectID: projectId,
            itemID: credentialId,
            pagination: createPagination(page),
        },
        {
            enabled: open && Boolean(credentialId),
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
    }, [open, credentialId]);

    useEffect(() => {
        const nextRows = query.data?.data;

        if (!nextRows) {
            return;
        }

        setRows(current =>
            page === 1 ? nextRows : appendUniqueByKey(current, nextRows, row => row.id || row.cloneURL || row.fullName),
        );
    }, [page, query.dataUpdatedAt, query.data?.data]);

    const normalizedSearch = search.trim().toLowerCase();
    const filteredRows = useMemo(() => {
        if (!normalizedSearch) {
            return rows;
        }

        return rows.filter(row =>
            [row.name, row.fullName, row.cloneURL, row.gitURL, row.defaultBranch].some(value =>
                includesSearch(value, normalizedSearch),
            ),
        );
    }, [normalizedSearch, rows]);
    const canLoadMore = (query.data?.data.length ?? 0) >= GIT_SELECTOR_PAGE_SIZE;

    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
        >
            <DialogFixedContent className="h-[800px] w-[1000px] max-w-[calc(100vw-2rem)] max-h-[calc(100vh-2rem)]">
                <DialogHeader>
                    <DialogTitle>Repositories</DialogTitle>
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
                                    <TableHead>Repository</TableHead>
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
                                ) : filteredRows.length === 0 ? (
                                    <TableRow>
                                        <TableCell
                                            colSpan={2}
                                            className="h-24 text-center text-muted-foreground"
                                        >
                                            No repositories found.
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    filteredRows.map(row => (
                                        <TableRow key={row.id || row.cloneURL || row.fullName}>
                                            <TableCell className="text-center">
                                                <Button
                                                    type="button"
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() => {
                                                        onSelect(row);
                                                        onOpenChange(false);
                                                    }}
                                                >
                                                    <Play className="size-4" /> Select
                                                </Button>
                                            </TableCell>
                                            <TableCell>
                                                <span
                                                    className="block max-w-[760px] truncate"
                                                    title={row.cloneURL}
                                                >
                                                    {row.cloneURL}
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

export function PullRequestsDialog({
    open,
    onOpenChange,
    projectId,
    credentialId,
    repository,
    isGithubAppCredential,
    onSelect,
}: GitRefSelectorDialogProps) {
    const [search, setSearch] = useState("");
    const [page, setPage] = useState(1);
    const [rows, setRows] = useState<ProjectGitPullRequest[]>([]);
    const requestOwner = isGithubAppCredential ? undefined : repository.owner;

    const query = ProjectGitCredentialsQueries.useFindManyPullRequests(
        {
            projectID: projectId,
            itemID: credentialId,
            owner: requestOwner,
            repo: repository.repo,
            pagination: createPagination(page),
        },
        {
            enabled: open && Boolean(credentialId) && Boolean(repository.repo),
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
    }, [open, credentialId, requestOwner, repository.repo]);

    useEffect(() => {
        const nextRows = query.data?.data;

        if (!nextRows) {
            return;
        }

        setRows(current => (page === 1 ? nextRows : appendUniqueByKey(current, nextRows, row => row.ref)));
    }, [page, query.dataUpdatedAt, query.data?.data]);

    const normalizedSearch = search.trim().toLowerCase();
    const filteredRows = useMemo(() => {
        if (!normalizedSearch) {
            return rows;
        }

        return rows.filter(row =>
            [row.number, row.title, row.branch, row.sha, row.ref, row.author, row.state].some(value =>
                includesSearch(value, normalizedSearch),
            ),
        );
    }, [normalizedSearch, rows]);
    const canLoadMore = (query.data?.data.length ?? 0) >= GIT_SELECTOR_PAGE_SIZE;

    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
        >
            <DialogFixedContent className="h-[800px] w-[1200px] max-w-[calc(100vw-2rem)] max-h-[calc(100vh-2rem)]">
                <DialogHeader>
                    <DialogTitle>Pull requests</DialogTitle>
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
                                    <TableHead className="w-[110px]">Number</TableHead>
                                    <TableHead>Title</TableHead>
                                    <TableHead className="w-[260px]">Source Branch</TableHead>
                                    <TableHead className="w-[140px]">SHA</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {query.isFetching && rows.length === 0 ? (
                                    <TableRow>
                                        <TableCell
                                            colSpan={5}
                                            className="h-24 text-center text-muted-foreground"
                                        >
                                            Loading...
                                        </TableCell>
                                    </TableRow>
                                ) : filteredRows.length === 0 ? (
                                    <TableRow>
                                        <TableCell
                                            colSpan={5}
                                            className="h-24 text-center text-muted-foreground"
                                        >
                                            No pull requests found.
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    filteredRows.map(row => (
                                        <TableRow key={row.ref}>
                                            <TableCell className="text-center">
                                                <Button
                                                    type="button"
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() => {
                                                        onSelect(row.ref);
                                                        onOpenChange(false);
                                                    }}
                                                >
                                                    <Play className="size-4" /> Deploy
                                                </Button>
                                            </TableCell>
                                            <TableCell>{row.number}</TableCell>
                                            <TableCell>
                                                <span
                                                    className="block max-w-[520px] truncate"
                                                    title={row.title}
                                                >
                                                    {row.title}
                                                </span>
                                            </TableCell>
                                            <TableCell>
                                                <span
                                                    className="block max-w-[240px] truncate"
                                                    title={row.branch}
                                                >
                                                    {row.branch}
                                                </span>
                                            </TableCell>
                                            <TableCell>{truncateSha(row.sha)}</TableCell>
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

export function BranchesDialog({
    open,
    onOpenChange,
    projectId,
    credentialId,
    repository,
    isGithubAppCredential,
    onSelect,
    selectLabel = "Select",
    dialogWidthClassName = "w-[800px]",
}: GitBranchesDialogProps) {
    const [search, setSearch] = useState("");
    const [page, setPage] = useState(1);
    const [rows, setRows] = useState<ProjectGitBranch[]>([]);
    const requestOwner = isGithubAppCredential ? undefined : repository.owner;

    const query = ProjectGitCredentialsQueries.useFindManyBranches(
        {
            projectID: projectId,
            itemID: credentialId,
            owner: requestOwner,
            repo: repository.repo,
            pagination: createPagination(page),
        },
        {
            enabled: open && Boolean(credentialId) && Boolean(repository.repo),
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
    }, [open, credentialId, requestOwner, repository.repo]);

    useEffect(() => {
        const nextRows = query.data?.data;

        if (!nextRows) {
            return;
        }

        setRows(current => (page === 1 ? nextRows : appendUniqueByKey(current, nextRows, row => row.ref)));
    }, [page, query.dataUpdatedAt, query.data?.data]);

    const normalizedSearch = search.trim().toLowerCase();
    const filteredRows = useMemo(() => {
        if (!normalizedSearch) {
            return rows;
        }

        return rows.filter(row => [row.name, row.ref, row.sha].some(value => includesSearch(value, normalizedSearch)));
    }, [normalizedSearch, rows]);
    const canLoadMore = (query.data?.data.length ?? 0) >= GIT_SELECTOR_PAGE_SIZE;

    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
        >
            <DialogFixedContent
                className={cn("h-[800px] max-w-[calc(100vw-2rem)] max-h-[calc(100vh-2rem)]", dialogWidthClassName)}
            >
                <DialogHeader>
                    <DialogTitle>Branches</DialogTitle>
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
                                    <TableHead>Branch</TableHead>
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
                                ) : filteredRows.length === 0 ? (
                                    <TableRow>
                                        <TableCell
                                            colSpan={2}
                                            className="h-24 text-center text-muted-foreground"
                                        >
                                            No branches found.
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    filteredRows.map(row => (
                                        <TableRow key={row.ref}>
                                            <TableCell className="text-center">
                                                <Button
                                                    type="button"
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() => {
                                                        onSelect(row.ref);
                                                        onOpenChange(false);
                                                    }}
                                                >
                                                    <Play className="size-4" /> {selectLabel}
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

interface GitRepositoriesDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    projectId: string;
    credentialId: string;
    onSelect: (repository: ProjectGitRepo) => void;
}

interface GitRefSelectorDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    projectId: string;
    credentialId: string;
    repository: ParsedGitRepository;
    isGithubAppCredential: boolean;
    onSelect: (ref: string) => void;
}

type GitBranchesDialogProps = GitRefSelectorDialogProps & {
    selectLabel?: string;
    dialogWidthClassName?: string;
};

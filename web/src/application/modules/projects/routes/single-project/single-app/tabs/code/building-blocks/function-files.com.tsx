import { useState } from "react";

import { cn } from "@lib/utils";
import { FilePlus, Trash2 } from "lucide-react";
import { type FunctionFile } from "~/projects/domain";
import { functionFilePathProblem } from "~/projects/module-shared/utils";

import { PopConfirm } from "@application/shared/components";

import { Button, Input } from "@/components/ui";

/**
 * The files of a function's code: one to edit at a time, a file to add, one to
 * remove.
 */
export function FunctionFiles({ files, selectedPath, readOnly, onSelect, onAdd, onRemove }: Props) {
    const [newPath, setNewPath] = useState("");
    const takenPaths = files.map(file => file.path);
    const problem = newPath === "" ? null : functionFilePathProblem(newPath, takenPaths);

    function handleAdd() {
        if (newPath === "" || problem) {
            return;
        }
        onAdd(newPath);
        setNewPath("");
    }

    return (
        <div className="flex w-full flex-col gap-2 lg:w-56 lg:shrink-0">
            <ul className="flex flex-col gap-0.5">
                {files.map(file => (
                    <li
                        key={file.path}
                        className={cn(
                            "group flex items-center justify-between gap-1 rounded-md px-2 py-1 text-sm",
                            file.path === selectedPath ? "bg-accent font-medium" : "hover:bg-accent/50",
                        )}
                    >
                        <button
                            type="button"
                            className="min-w-0 flex-1 truncate text-left font-mono"
                            onClick={() => {
                                onSelect(file.path);
                            }}
                        >
                            {file.path}
                        </button>
                        {!readOnly && files.length > 1 ? (
                            <PopConfirm
                                title={`Remove ${file.path}?`}
                                description="The file is removed from the code; saving deploys the code without it."
                                confirmText="Remove"
                                variant="destructive"
                                onConfirm={() => {
                                    onRemove(file.path);
                                }}
                            >
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon-sm"
                                    className="size-6 opacity-0 group-hover:opacity-100"
                                    aria-label={`Remove ${file.path}`}
                                >
                                    <Trash2 className="size-3.5" />
                                </Button>
                            </PopConfirm>
                        ) : null}
                    </li>
                ))}
            </ul>
            {!readOnly ? (
                <form
                    className="flex flex-col gap-1"
                    onSubmit={event => {
                        event.preventDefault();
                        handleAdd();
                    }}
                >
                    <div className="flex gap-1">
                        <Input
                            value={newPath}
                            onChange={event => {
                                setNewPath(event.target.value.trim());
                            }}
                            placeholder="lib/util.js"
                            className="h-8 font-mono text-xs"
                            aria-invalid={Boolean(problem)}
                            aria-label="New file's path"
                        />
                        <Button
                            type="submit"
                            variant="outline"
                            size="icon-sm"
                            disabled={newPath === "" || Boolean(problem)}
                            aria-label="Add the file"
                        >
                            <FilePlus />
                        </Button>
                    </div>
                    {problem ? <span className="text-xs text-destructive">{problem}</span> : null}
                </form>
            ) : null}
        </div>
    );
}

interface Props {
    files: FunctionFile[];
    selectedPath: string;
    readOnly: boolean;
    onSelect: (path: string) => void;
    onAdd: (path: string) => void;
    onRemove: (path: string) => void;
}

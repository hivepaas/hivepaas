import { useRef } from "react";

import { Button, Checkbox, Label } from "@components/ui";
import { Badge } from "@components/ui/badge";
import { Textarea } from "@components/ui/textarea";
import { FileUpIcon, XIcon } from "lucide-react";
import type { ComposeFileInput, ComposeFileNeed } from "~/operations/domain";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

const AS_LABELS: Record<string, string> = {
    env_file: "env_file",
    config: "config",
    secret: "secret",
    bind: "mounted file",
    directory: "mounted directory",
    compose: "compose file",
};

/** What the page knows of a directory the compose file mounts. */
export interface ComposeDirectoryState {
    /** The folder's files under it, but those for git or the desktop. */
    available: number;
    /** The files under it given with the request. */
    given: number;
    /** Its files are given from the folder. */
    on: boolean;
    /** They fit what a request carries. */
    fits: boolean;
}

/**
 * The files the compose file reads: an env_file, a config's or a secret's
 * file, a file it mounts, a compose file it includes or extends. Each is
 * pasted or opened - or taken from the folder opened. One left missing is
 * created empty, and the app will not start right until it is filled; a
 * compose file has to be given before anything more is read.
 */
export function ComposeFiles({ needs, files, onChange, directories, onToggleDirectory, hasFolder }: Props) {
    if (needs.length === 0) {
        return null;
    }

    return (
        <InfoBlock
            titleWidth={220}
            title={
                <LabelWithInfo
                    label="Files"
                    content="The files beside the compose file that it reads. They travel with each request and are kept only as the settings they become."
                />
            }
        >
            <div className="flex w-full max-w-[900px] flex-col gap-3">
                {needs.map(need =>
                    need.as === "directory" ? (
                        <DirectoryRow
                            key={need.path}
                            need={need}
                            state={directories[need.path]}
                            hasFolder={hasFolder}
                            onToggle={on => {
                                onToggleDirectory(need.path, on);
                            }}
                        />
                    ) : (
                        <FileRow
                            key={need.path}
                            need={need}
                            value={files[need.path]}
                            onChange={value => {
                                const others = Object.entries(files).filter(([path]) => path !== need.path);
                                onChange(
                                    Object.fromEntries(value === undefined ? others : [...others, [need.path, value]]),
                                );
                            }}
                        />
                    ),
                )}
            </div>
        </InfoBlock>
    );
}

function FileRow({ need, value, onChange }: RowProps) {
    const inputRef = useRef<HTMLInputElement>(null);
    const picked = value instanceof File ? value : undefined;

    return (
        <div className="flex flex-col gap-2 rounded-md border px-3 py-2">
            <div className="flex flex-wrap items-center gap-2">
                <code className="font-mono text-xs text-foreground">{need.path}</code>
                <Badge variant="outline">{AS_LABELS[need.as] ?? need.as}</Badge>
                {need.by.length > 0 && (
                    <span className="text-xs text-muted-foreground">read by {need.by.join(", ")}</span>
                )}
                {value === undefined && (
                    <Badge
                        variant="outline"
                        className="border-amber-500/40 text-amber-700 dark:text-amber-400"
                    >
                        missing
                    </Badge>
                )}
            </div>
            {picked ? (
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    {picked.name} · {picked.size} bytes
                    <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                            onChange(undefined);
                        }}
                    >
                        <XIcon className="size-4" />
                    </Button>
                </div>
            ) : (
                <Textarea
                    value={typeof value === "string" ? value : ""}
                    onChange={event => {
                        onChange(event.target.value === "" ? undefined : event.target.value);
                    }}
                    placeholder="Paste its content, or open the file"
                    minRows={2}
                    maxRows={10}
                    spellCheck={false}
                    className="font-mono text-xs"
                />
            )}
            <div>
                <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => inputRef.current?.click()}
                >
                    <FileUpIcon className="size-4" />
                    Open file
                </Button>
                <input
                    ref={inputRef}
                    type="file"
                    className="hidden"
                    onChange={event => {
                        const file = event.target.files?.[0];
                        if (inputRef.current) {
                            inputRef.current.value = "";
                        }
                        if (file) {
                            onChange(file);
                        }
                    }}
                />
            </div>
        </div>
    );
}

/**
 * A directory the compose file mounts: the app's own on the project's volume,
 * with the folder's files under it mounted in it, read only - or empty.
 */
function DirectoryRow({ need, state, hasFolder, onToggle }: DirectoryProps) {
    const id = `compose-directory-${need.path}`;

    return (
        <div className="flex flex-col gap-2 rounded-md border px-3 py-2">
            <div className="flex flex-wrap items-center gap-2">
                <code className="font-mono text-xs text-foreground">{need.path}/</code>
                <Badge variant="outline">{AS_LABELS["directory"]}</Badge>
                {need.by.length > 0 && <span className="text-xs text-muted-foreground">by {need.by.join(", ")}</span>}
                {state && state.given > 0 && (
                    <Badge variant="outline">
                        {state.given} {state.given === 1 ? "file" : "files"} in it
                    </Badge>
                )}
            </div>
            {state && state.available > 0 ? (
                <div className="flex items-start gap-2">
                    <Checkbox
                        id={id}
                        className="mt-0.5"
                        checked={state.on}
                        onCheckedChange={value => {
                            onToggle(value === true);
                        }}
                    />
                    <div className="flex flex-col">
                        <Label
                            htmlFor={id}
                            className="text-sm font-normal"
                        >
                            Mount the folder&apos;s {state.available} {state.available === 1 ? "file" : "files"} in it,
                            read only
                        </Label>
                        <p className="text-xs text-muted-foreground">
                            {state.on && !state.fits
                                ? "They are more than a request carries - 100 files, 5 MB, 500 KB each - with the others: the directory starts empty."
                                : "The directory is the app's own on the project's volume, and what the app writes beside them is kept. Otherwise it starts empty."}
                        </p>
                    </div>
                </div>
            ) : (
                <p className="text-xs text-muted-foreground">
                    {hasFolder
                        ? "The folder has no file in it: it starts empty, the app's own on the project's volume."
                        : "It starts empty, the app's own on the project's volume. Open the compose file's folder to have its files in it."}
                </p>
            )}
        </div>
    );
}

interface Props {
    needs: ComposeFileNeed[];
    files: Record<string, ComposeFileInput>;
    onChange: (files: Record<string, ComposeFileInput>) => void;
    /** What the page knows of each directory the compose file mounts, by its path. */
    directories: Record<string, ComposeDirectoryState>;
    onToggleDirectory: (path: string, on: boolean) => void;
    hasFolder: boolean;
}

interface DirectoryProps {
    need: ComposeFileNeed;
    state: ComposeDirectoryState | undefined;
    hasFolder: boolean;
    onToggle: (on: boolean) => void;
}

interface RowProps {
    need: ComposeFileNeed;
    value: ComposeFileInput | undefined;
    onChange: (value: ComposeFileInput | undefined) => void;
}

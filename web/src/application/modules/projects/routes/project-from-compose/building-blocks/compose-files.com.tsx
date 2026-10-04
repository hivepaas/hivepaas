import { useRef } from "react";

import { Button } from "@components/ui";
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
};

/**
 * The files the compose file reads: an env_file, a config's or a secret's
 * file, a file it mounts. Each is pasted or opened; one left missing is created
 * empty, and the app will not start right until it is filled.
 */
export function ComposeFiles({ needs, files, onChange }: Props) {
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
                {needs.map(need => (
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
                ))}
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
                <span className="text-xs text-muted-foreground">read by {need.by.join(", ")}</span>
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

interface Props {
    needs: ComposeFileNeed[];
    files: Record<string, ComposeFileInput>;
    onChange: (files: Record<string, ComposeFileInput>) => void;
}

interface RowProps {
    need: ComposeFileNeed;
    value: ComposeFileInput | undefined;
    onChange: (value: ComposeFileInput | undefined) => void;
}

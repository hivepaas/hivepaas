import { useId, useState } from "react";

import { Check, FilePen } from "lucide-react";
import Prism from "prismjs";
import "prismjs/components/prism-docker";
import "prismjs/themes/prism-tomorrow.css";
import type { FieldError as ReactHookFormFieldError } from "react-hook-form";
import Editor from "react-simple-code-editor";

import { Button, FieldError } from "@/components/ui";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogContent,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";

function highlightDockerfile(code: string): string {
    const dockerGrammar = Prism.languages["docker"];

    if (!dockerGrammar) {
        return code;
    }

    return Prism.highlight(code, dockerGrammar, "docker");
}

export function DockerfileContentEditor({ value, onChange, invalid, error, readOnly = false }: Props) {
    const generatedId = useId();
    const textareaId = `dockerfile-editor-${generatedId}`;
    const [open, setOpen] = useState(false);

    return (
        <div className="flex flex-col gap-2">
            <Button
                type="button"
                variant="outline"
                onClick={() => {
                    setOpen(true);
                }}
                aria-invalid={invalid}
                disabled={readOnly}
            >
                <FilePen className="size-4" />
                Edit
            </Button>
            <FieldError errors={[error]} />

            <Dialog
                open={open}
                onOpenChange={setOpen}
            >
                <DialogContent
                    className="fixed inset-4 top-4 left-4 z-50 flex max-w-none translate-x-0 translate-y-0 flex-col gap-0 overflow-hidden rounded-lg border bg-background p-0 shadow-2xl sm:max-w-none w-[inherit]"
                    onEscapeKeyDown={event => {
                        event.stopPropagation();
                    }}
                >
                    <DialogHeader className="shrink-0 border-b px-4 py-3">
                        <DialogTitle>Dockerfile</DialogTitle>
                    </DialogHeader>
                    <DialogBody className="flex min-h-0 flex-1 flex-col overflow-hidden p-0">
                        <div className="min-h-0 flex-1 overflow-auto bg-[#1e1e1e]">
                            <Editor
                                value={value}
                                onValueChange={onChange}
                                highlight={highlightDockerfile}
                                padding={16}
                                textareaId={textareaId}
                                readOnly={readOnly}
                                style={{
                                    minHeight: "100%",
                                    fontFamily:
                                        "'Fira Code', 'Fira Mono', ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace'",
                                    fontSize: 14,
                                    lineHeight: 1.6,
                                    backgroundColor: "#1e1e1e",
                                    color: "#f8f8f2",
                                    outline: "none",
                                }}
                            />
                        </div>
                    </DialogBody>
                    <DialogActionFooter className="gap-2">
                        <Button
                            type="button"
                            onClick={() => {
                                setOpen(false);
                            }}
                        >
                            <Check className="size-4" />
                            Done
                        </Button>
                    </DialogActionFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}

interface Props {
    value: string;
    onChange: (value: string) => void;
    invalid: boolean;
    error?: ReactHookFormFieldError;
    readOnly?: boolean;
}

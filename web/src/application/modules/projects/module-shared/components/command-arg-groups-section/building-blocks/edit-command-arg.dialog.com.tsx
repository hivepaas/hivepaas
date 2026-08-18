import { useEffect, useState } from "react";

import { InfoBlock } from "@application/shared/components";

import { Button, Input } from "@/components/ui";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";

export function EditCommandArgDialog({ open, onOpenChange, name, value, onSave, readOnly = false }: Props) {
    const [draftName, setDraftName] = useState(name);
    const [draftValue, setDraftValue] = useState(value);

    useEffect(() => {
        if (open) {
            setDraftName(name);
            setDraftValue(value);
        }
    }, [open, name, value]);

    function handleSave() {
        onSave({ name: draftName, value: draftValue });
        onOpenChange(false);
    }

    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
        >
            <DialogFixedContent
                className="flex w-[700px] max-w-[calc(100%-2rem)] flex-col"
                onEscapeKeyDown={event => {
                    event.stopPropagation();
                }}
            >
                <DialogHeader>
                    <DialogTitle>Edit Command Arg</DialogTitle>
                </DialogHeader>

                <DialogBody className="flex min-h-0 flex-1 flex-col gap-6">
                    <InfoBlock
                        title="Arg"
                        titleWidth={80}
                    >
                        <Input
                            id="edit-command-arg-name"
                            value={draftName}
                            onChange={event => {
                                setDraftName(event.target.value);
                            }}
                            placeholder="--arg1"
                            disabled={readOnly}
                        />
                    </InfoBlock>

                    <InfoBlock
                        title="Value"
                        titleWidth={80}
                    >
                        <Textarea
                            id="edit-command-arg-value"
                            value={draftValue}
                            onChange={event => {
                                setDraftValue(event.target.value);
                            }}
                            placeholder="value"
                            disabled={readOnly}
                            minRows={6}
                            className="min-h-0 flex-1 resize-y whitespace-pre-wrap"
                        />
                    </InfoBlock>
                </DialogBody>

                <DialogActionFooter>
                    <Button
                        type="button"
                        className="min-w-[100px]"
                        onClick={handleSave}
                        disabled={readOnly}
                    >
                        Save
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}

interface Props {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    name: string;
    value: string;
    onSave: (values: { name: string; value: string }) => void;
    readOnly?: boolean;
}

import { useState } from "react";

import { Button, FieldError, Input } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { ChevronDown, ChevronRight, X } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import {
    type AppConfigHttpSettingsFormSchemaInput,
    type AppConfigHttpSettingsFormSchemaOutput,
} from "../schemas";

interface BasicAuthSectionProps {
    prefix: string;
    onRemove: () => void;
}

export function BasicAuthSection({ prefix, onRemove }: BasicAuthSectionProps) {
    const [open, setOpen] = useState(true);

    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const {
        field: nameField,
        fieldState: { error: nameError },
    } = useController({
        control,
        name: `${prefix}.name` as never,
    });

    useController({
        control,
        name: `${prefix}.id` as never,
    });

    return (
        <Collapsible
            open={open}
            onOpenChange={setOpen}
        >
            <div className="flex w-full items-center gap-1 rounded-md border border-dashed px-1">
                <CollapsibleTrigger asChild>
                    <button
                        type="button"
                        className="flex min-w-0 flex-1 items-center gap-2 px-2 py-2 text-sm font-medium hover:bg-accent"
                    >
                        {open ? <ChevronDown className="size-4 shrink-0" /> : <ChevronRight className="size-4 shrink-0" />}
                        Basic Auth
                    </button>
                </CollapsibleTrigger>
                <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="shrink-0 text-muted-foreground hover:text-destructive"
                    title="Remove section"
                    onClick={onRemove}
                >
                    <X className="size-4" />
                </Button>
            </div>
            <CollapsibleContent>
                <div className="flex flex-col gap-4 border-l-2 border-accent pl-4 pt-4">
                    <InfoBlock title="Credential">
                        <Input
                            value={nameField.value ?? ""}
                            placeholder="Select basic auth credential"
                            readOnly
                            className="max-w-[400px] text-muted-foreground"
 />
                        <FieldError errors={[nameError]} />
                    </InfoBlock>
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}

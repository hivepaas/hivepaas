import { useState } from "react";

import { Button, FieldError, Input } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { Textarea } from "@components/ui/textarea";
import { ChevronDown, ChevronRight, X } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

interface CompressionConfigSectionProps {
    prefix: string;
    onRemove?: () => void;
}

export function CompressionConfigSection({ prefix, onRemove }: CompressionConfigSectionProps) {
    const [open, setOpen] = useState(false);

    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const {
        field: defaultEncoding,
        fieldState: { error: defaultEncodingError, invalid: isDefaultEncodingInvalid },
    } = useController({ control, name: `${prefix}.defaultEncoding` as never });
    const {
        field: minResponseBody,
        fieldState: { error: minResponseBodyError, invalid: isMinResponseBodyInvalid },
    } = useController({ control, name: `${prefix}.minResponseBody` as never });

    const {
        field: excludedContentTypes,
        fieldState: { error: excludedContentTypesError, invalid: isExcludedContentTypesInvalid },
    } = useController({ control, name: `${prefix}.excludedContentTypes` as never });
    const {
        field: includedContentTypes,
        fieldState: { error: includedContentTypesError, invalid: isIncludedContentTypesInvalid },
    } = useController({ control, name: `${prefix}.includedContentTypes` as never });

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
                        {open ? (
                            <ChevronDown className="size-4 shrink-0" />
                        ) : (
                            <ChevronRight className="size-4 shrink-0" />
                        )}
                        Compression Configuration
                        <a
                            className="text-xs text-blue-500 hover:text-blue-600"
                            href="https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/compress/"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            (docs)
                        </a>
                    </button>
                </CollapsibleTrigger>
                {onRemove && (
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
                )}
            </div>
            <CollapsibleContent>
                <div className="flex flex-col gap-4 border-l-2 border-accent pl-4 pt-4">
                    <InfoBlock title="Default Encoding">
                        <Select
                            value={defaultEncoding.value}
                            onValueChange={defaultEncoding.onChange}
                        >
                            <SelectTrigger
                                className="max-w-[100px]"
                                aria-invalid={isDefaultEncodingInvalid}
                            >
                                <SelectValue placeholder="..." />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="br">brotli</SelectItem>
                                <SelectItem value="zstd">zstd</SelectItem>
                                <SelectItem value="gzip">gzip</SelectItem>
                            </SelectContent>
                        </Select>
                        <FieldError errors={[defaultEncodingError]} />
                    </InfoBlock>

                    <InfoBlock title="Excluded Content Types">
                        <Textarea
                            {...excludedContentTypes}
                            value={excludedContentTypes.value}
                            onChange={excludedContentTypes.onChange}
                            placeholder="text/plain"
                            rows={2}
                            className="resize-y max-w-[400px]"
                            aria-invalid={isExcludedContentTypesInvalid}
                        />
                        <FieldError errors={[excludedContentTypesError]} />
                    </InfoBlock>

                    <InfoBlock title="Included Content Types">
                        <Textarea
                            {...includedContentTypes}
                            value={includedContentTypes.value}
                            onChange={includedContentTypes.onChange}
                            placeholder="text/html"
                            rows={2}
                            className="resize-y max-w-[400px]"
                            aria-invalid={isIncludedContentTypesInvalid}
                        />
                        <FieldError errors={[includedContentTypesError]} />
                    </InfoBlock>
                    <InfoBlock title="Min Response Body Size">
                        <Input
                            value={minResponseBody.value}
                            onChange={minResponseBody.onChange}
                            className="max-w-[100px]"
                            aria-invalid={isMinResponseBodyInvalid}
                        />
                        <FieldError errors={[minResponseBodyError]} />
                    </InfoBlock>
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}

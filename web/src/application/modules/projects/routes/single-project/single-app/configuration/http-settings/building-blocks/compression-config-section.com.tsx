import { useEffect, useState } from "react";

import { Button, Checkbox, FieldError, Input } from "@components/ui";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@components/ui/collapsible";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { Textarea } from "@components/ui/textarea";
import { ChevronDown, ChevronRight, X } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS } from "../http-settings-layout.constants";
import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

interface CompressionConfigSectionProps {
    prefix: string;
    autoExpandToken?: number;
    readOnly?: boolean;
    onRemove?: () => void;
}

export function CompressionConfigSection({
    prefix,
    autoExpandToken,
    readOnly = false,
    onRemove,
}: CompressionConfigSectionProps) {
    const [open, setOpen] = useState(false);
    useEffect(() => {
        if (autoExpandToken !== undefined) {
            setOpen(true);
        }
    }, [autoExpandToken]);

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
    const { field: enabled } = useController({ control, name: `${prefix}.enabled` as never });
    const isEnabled = Boolean(enabled.value);

    return (
        <Collapsible
            open={open}
            onOpenChange={setOpen}
        >
            <div className="flex justify-between items-center font-medium bg-accent py-2 px-3 rounded-lg">
                <CollapsibleTrigger asChild>
                    <button
                        type="button"
                        className="flex min-w-0 flex-1 items-center gap-2 text-sm font-medium hover:bg-accent"
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
                        className="shrink-0 text-muted-foreground hover:text-destructive h-fit"
                        title="Remove section"
                        onClick={() => {
                            if (readOnly) {
                                return;
                            }

                            onRemove();
                        }}
                        disabled={readOnly}
                    >
                        <X className="size-4" />
                    </Button>
                )}
            </div>
            <CollapsibleContent>
                <div className="flex flex-col gap-4 border-l-2 border-accent pl-4 pt-4">
                    <InfoBlock title="Enabled">
                        <Checkbox
                            checked={isEnabled}
                            onCheckedChange={value => {
                                if (readOnly) {
                                    return;
                                }

                                enabled.onChange(value);
                            }}
                            disabled={readOnly}
                        />
                    </InfoBlock>
                    {isEnabled && (
                        <>
                            <InfoBlock title="Default Encoding">
                                <Select
                                    value={defaultEncoding.value}
                                    onValueChange={value => {
                                        if (readOnly) {
                                            return;
                                        }

                                        defaultEncoding.onChange(value);
                                    }}
                                    disabled={readOnly}
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

                            <InfoBlock title="Included Content Types">
                                <Textarea
                                    {...includedContentTypes}
                                    value={includedContentTypes.value}
                                    onChange={includedContentTypes.onChange}
                                    placeholder="auto compress all text types"
                                    rows={2}
                                    className={`resize-y ${HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS}`}
                                    aria-invalid={isIncludedContentTypesInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[includedContentTypesError]} />
                            </InfoBlock>
                            <InfoBlock title="Excluded Content Types">
                                <Textarea
                                    {...excludedContentTypes}
                                    value={excludedContentTypes.value}
                                    onChange={excludedContentTypes.onChange}
                                    placeholder="content types"
                                    rows={2}
                                    className={`resize-y ${HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS}`}
                                    aria-invalid={isExcludedContentTypesInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[excludedContentTypesError]} />
                            </InfoBlock>

                            <InfoBlock title="Min Response Body Size">
                                <Input
                                    value={minResponseBody.value}
                                    onChange={minResponseBody.onChange}
                                    className="max-w-[100px]"
                                    aria-invalid={isMinResponseBodyInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[minResponseBodyError]} />
                            </InfoBlock>
                        </>
                    )}
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}

import { Checkbox, FieldError } from "@components/ui";
import { Collapsible, CollapsibleContent } from "@components/ui/collapsible";
import { Textarea } from "@components/ui/textarea";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";

import { HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS } from "../http-settings-layout.constants";
import type { HivePaaSHttpSettingsFormInput, HivePaaSHttpSettingsFormOutput } from "../schemas";

interface ClientConfigSectionProps {
    prefix: `domains.${number}.clientConfig`;
    readOnly?: boolean;
}

export function ClientConfigSection({ prefix, readOnly = false }: ClientConfigSectionProps) {
    const { control } = useFormContext<HivePaaSHttpSettingsFormInput, unknown, HivePaaSHttpSettingsFormOutput>();

    const {
        field: allowedIPs,
        fieldState: { error: allowedIPsError, invalid: isAllowedIPsInvalid },
    } = useController({ control, name: `${prefix}.allowedIPs` });
    const { field: enabled } = useController({ control, name: `${prefix}.enabled` });
    const isEnabled = enabled.value;

    return (
        <Collapsible open>
            <div className="flex justify-between items-center font-medium bg-accent py-2 px-3 rounded-lg">
                <div className="flex min-w-0 flex-1 items-center gap-2 text-sm font-medium">
                    <span>Client Configuration</span>
                    <a
                        className="text-xs text-blue-500 hover:text-blue-600"
                        href="https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/ipallowlist/"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        (docs)
                    </a>
                </div>
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

                                enabled.onChange(value === true);
                            }}
                            disabled={readOnly}
                        />
                    </InfoBlock>
                    {isEnabled && (
                        <InfoBlock title="Allowed IPs">
                            <div className={`flex flex-col gap-2 ${HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS}`}>
                                <Textarea
                                    {...allowedIPs}
                                    onChange={allowedIPs.onChange}
                                    placeholder="1.2.3.4, 1.2.3.4/24, ipv6"
                                    rows={2}
                                    className="resize-y"
                                    aria-invalid={isAllowedIPsInvalid}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[allowedIPsError]} />
                            </div>
                        </InfoBlock>
                    )}
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}

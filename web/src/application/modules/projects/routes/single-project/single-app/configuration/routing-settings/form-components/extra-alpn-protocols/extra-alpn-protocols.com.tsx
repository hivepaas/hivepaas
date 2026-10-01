import React from "react";

import { FieldError, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS } from "../../routing-settings-layout.constants";
import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../../schemas";

/**
 * Protocols a TCP domain ending TLS accepts in the handshake beyond the ones
 * HivePaaS lists for every such domain - PostgreSQL's, SQL Server's, MQTT's,
 * DNS over TLS' and others. Only a client that offers one HivePaaS does not
 * know of needs it: without it, it is refused.
 */
function View({ domainIndex, readOnly = false }: ExtraAlpnProtocolsProps) {
    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const {
        field,
        fieldState: { error, invalid },
    } = useController({ control, name: `domains.${domainIndex}.extraAlpnProtocols` });

    return (
        <InfoBlock
            titleWidth={240}
            title={
                <LabelWithInfo
                    label="Extra ALPN Protocols"
                    content="Rarely needed. Protocols this domain accepts in the TLS handshake beyond the ones HivePaaS already lists (postgresql, tds/8.0, mqtt, imap, pop3, managesieve, xmpp-client, xmpp-server, dot, h2, http/1.1). Add one only for a client refused with 'no application protocol', such as x-amzn-mqtt-ca."
                />
            }
        >
            <Input
                {...field}
                placeholder="e.g. x-amzn-mqtt-ca"
                className={HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS}
                aria-invalid={invalid}
                disabled={readOnly}
            />
            <FieldError errors={[error]} />
        </InfoBlock>
    );
}

interface ExtraAlpnProtocolsProps {
    domainIndex: number;
    readOnly?: boolean;
}

export const ExtraAlpnProtocols = React.memo(View);

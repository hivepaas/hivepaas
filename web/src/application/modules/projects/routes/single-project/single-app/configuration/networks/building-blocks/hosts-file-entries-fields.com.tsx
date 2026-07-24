import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

import { type AppConfigNetworksFormSchemaInput } from "../schemas";

export function HostsFileEntriesFields({ readOnly = false }: Props) {
    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Hosts file entries"
                    content="Static host mappings written to container hosts file."
                />
            }
        >
            <KeyValueList<AppConfigNetworksFormSchemaInput>
                name="hostsFileEntries"
                keyField="address"
                valueField="hostnamesText"
                keyLabel="Addr"
                valueLabel="Name"
                keyPlaceholder="11.22.33.44"
                valuePlaceholder="hostname alias1 alias2"
                enableValueEditing
                disabled={readOnly}
                className="max-w-[590px]"
            />
        </InfoBlock>
    );
}

type Props = {
    readOnly?: boolean;
};

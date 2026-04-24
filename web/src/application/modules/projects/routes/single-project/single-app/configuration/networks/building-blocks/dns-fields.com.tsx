import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { SingleValueList } from "@application/shared/form";

import { type AppConfigNetworksFormSchemaInput } from "../schemas";

type DnsListName = "dnsConfig.nameservers" | "dnsConfig.search" | "dnsConfig.options";

export function DNSFields() {
    return (
        <div className="flex flex-col gap-6">
            {(
                [
                    {
                        listName: "dnsConfig.nameservers" as DnsListName,
                        label: "DNS: Nameservers",
                        placeholder: "nameserver",
                        content: "Add values for DNS nameservers.",
                    },
                    {
                        listName: "dnsConfig.search" as DnsListName,
                        label: "DNS: Search",
                        placeholder: "search",
                        content: "Add values for DNS search.",
                    },
                    {
                        listName: "dnsConfig.options" as DnsListName,
                        label: "DNS: Options",
                        placeholder: "option",
                        content: "Add values for DNS options.",
                    },
                ] as const
            ).map(({ listName, label, placeholder, content }) => (
                <InfoBlock
                    key={listName}
                    title={
                        <LabelWithInfo
                            label={label}
                            content={content}
                        />
                    }
                >
                    <SingleValueList<AppConfigNetworksFormSchemaInput>
                        name={listName}
                        placeholder={placeholder}
                        className="max-w-[590px]"
                    />
                </InfoBlock>
            ))}
        </div>
    );
}

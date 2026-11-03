import { RoutingConfigurableSections } from "./routing-configurable-sections.com";

interface DomainConfigurableSectionsProps {
    domainIndex: number;
    readOnly?: boolean;
}

export function DomainConfigurableSections({ domainIndex, readOnly = false }: DomainConfigurableSectionsProps) {
    const basePath = `domains.${domainIndex}`;
    return (
        <RoutingConfigurableSections
            basePath={basePath}
            readOnly={readOnly}
        />
    );
}

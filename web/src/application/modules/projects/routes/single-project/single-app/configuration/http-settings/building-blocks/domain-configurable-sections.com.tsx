import { HttpConfigurableSections } from "./http-configurable-sections.com";

interface DomainConfigurableSectionsProps {
    domainIndex: number;
    readOnly?: boolean;
}

export function DomainConfigurableSections({ domainIndex, readOnly = false }: DomainConfigurableSectionsProps) {
    const basePath = `domains.${domainIndex}`;
    return (
        <HttpConfigurableSections
            basePath={basePath}
            readOnly={readOnly}
        />
    );
}

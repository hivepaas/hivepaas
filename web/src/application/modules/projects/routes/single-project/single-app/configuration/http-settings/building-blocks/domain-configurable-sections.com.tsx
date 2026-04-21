import { HttpConfigurableSections } from "./http-configurable-sections.com";

interface DomainConfigurableSectionsProps {
    domainIndex: number;
}

export function DomainConfigurableSections({ domainIndex }: DomainConfigurableSectionsProps) {
    const basePath = `domains.${domainIndex}`;
    return (
        <HttpConfigurableSections
            basePath={basePath}
            scope="domain"
        />
    );
}

import React from "react";

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@components/ui/accordion";

import { ConfigVariables } from "@application/shared/form";

import { type EnvVarsFormBaseSchemaInput } from "../../schemas";

function View({ search, viewMode, isRevealed, title, name, readOnly = false }: Props) {
    return (
        <Accordion
            type="single"
            collapsible
            className="w-full"
            defaultValue="build-time-env-vars"
        >
            <AccordionItem
                value="build-time-env-vars"
                className=""
            >
                <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent">
                    {title}
                </AccordionTrigger>
                <AccordionContent className="pt-4 pb-0 pl-3">
                    <ConfigVariables<EnvVarsFormBaseSchemaInput>
                        name={name}
                        search={search}
                        viewMode={viewMode}
                        isRevealed={isRevealed}
                        readOnly={readOnly}
                    />
                </AccordionContent>
            </AccordionItem>
        </Accordion>
    );
}

type Props = {
    search: string;
    viewMode: "merge" | "individual";
    isRevealed: boolean;
    title: React.ReactNode;
    name: keyof EnvVarsFormBaseSchemaInput;
    readOnly?: boolean;
};

export const EnvVarsBaseForm = React.memo(View);

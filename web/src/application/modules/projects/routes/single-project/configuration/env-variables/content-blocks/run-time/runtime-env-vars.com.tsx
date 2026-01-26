import React from "react";

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@components/ui/accordion";

import { ConfigVariables } from "@application/shared/form";

import { type ProjectEnvVarsFormSchemaInput } from "../../schemas";

function View({ search, viewMode, isRevealed }: Props) {
    return (
        <Accordion
            type="single"
            collapsible
            className="w-full"
            defaultValue="runtime-env-vars"
        >
            <AccordionItem
                value="runtime-env-vars"
                className=""
            >
                <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent">
                    Runtime Env Vars
                </AccordionTrigger>
                <AccordionContent className="pt-4 pb-0">
                    <ConfigVariables<ProjectEnvVarsFormSchemaInput>
                        name="runtime"
                        search={search}
                        viewMode={viewMode}
                        isRevealed={isRevealed}
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
};

export const RuntimeEnvVars = React.memo(View);

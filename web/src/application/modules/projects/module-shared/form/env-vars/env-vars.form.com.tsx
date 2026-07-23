import React from "react";

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@components/ui/accordion";

import { ConfigVariables } from "@application/shared/form";

function View({ search, viewMode, isRevealed, title, name, readOnly = false, alwaysExpanded = false, notice }: Props) {
    const content = (
        <div className="flex flex-col gap-4">
            {notice}
            <ConfigVariables
                name={name as never}
                search={search}
                viewMode={viewMode}
                isRevealed={isRevealed}
                readOnly={readOnly}
            />
        </div>
    );

    if (alwaysExpanded) {
        return (
            <div className="w-full">
                <div className="bg-accent px-3 py-2 text-sm font-medium">{title}</div>
                <div className="pt-4 pb-0 pl-3">{content}</div>
            </div>
        );
    }

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
                <AccordionContent className="pt-4 pb-0 pl-3">{content}</AccordionContent>
            </AccordionItem>
        </Accordion>
    );
}

type Props = {
    search: string;
    viewMode: "merge" | "individual";
    isRevealed: boolean;
    title: React.ReactNode;
    name: "buildtime" | "runtime" | "shared";
    readOnly?: boolean;
    alwaysExpanded?: boolean;
    notice?: React.ReactNode;
};

export const EnvVarsBaseForm = React.memo(View);

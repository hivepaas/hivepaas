import React from "react";

import { Checkbox } from "@components/ui";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@components/ui/accordion";
import { Input } from "@components/ui/input";

type EnvVarRecord = {
    key: string;
    value: string;
    isLiteral: boolean;
};

type Props = {
    title: React.ReactNode;
    items: EnvVarRecord[];
    isRevealed?: boolean;
    search?: string;
};

function View({ title, items, isRevealed = false, search = "" }: Props) {
    const filteredItems = search
        ? items.filter(item => {
              const q = search.toLowerCase().trim();
              return item.key.toLowerCase().includes(q) || item.value.toLowerCase().includes(q);
          })
        : items;

    return (
        <Accordion
            type="single"
            defaultValue="inherited-env-vars"
            className="w-full"
        >
            <AccordionItem value="inherited-env-vars">
                <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent">
                    {title}
                </AccordionTrigger>
                <AccordionContent className="pt-4 pb-0">
                    {filteredItems.length > 0 ? (
                        <div className="flex flex-col gap-4">
                            <div className="space-y-3">
                                {filteredItems.map((item, index) => (
                                    <div
                                        key={index}
                                        className="flex items-center gap-3"
                                    >
                                        {/* Key */}
                                        <div className="flex-1">
                                            <Input
                                                value={item.key}
                                                readOnly
                                                disabled
                                                placeholder="Key"
                                                className="bg-muted cursor-default"
                                            />
                                        </div>

                                        {/* Value */}
                                        <div className="flex-1">
                                            <Input
                                                type={isRevealed ? "text" : "password"}
                                                value={item.value}
                                                readOnly
                                                disabled
                                                placeholder="Value"
                                                className="bg-muted cursor-default"
                                            />
                                        </div>

                                        {/* Literal badge */}
                                        <div className="flex items-center gap-2">
                                            <Checkbox
                                                checked
                                                disabled
                                            />
                                            <span className="text-sm cursor-pointer">Literal</span>
                                        </div>

                                        {/* Spacer to align with editable rows that have a delete button */}
                                        <div className="h-8 w-8" />
                                    </div>
                                ))}
                            </div>
                        </div>
                    ) : (
                        <div className="text-sm text-muted-foreground py-4 text-center">
                            {search ? "No records match your search" : "No inherited environment variables"}
                        </div>
                    )}
                </AccordionContent>
            </AccordionItem>
        </Accordion>
    );
}

export const InheritedEnvVarsAccordion = React.memo(View);

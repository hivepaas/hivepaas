import React from "react";

import { Checkbox } from "@components/ui";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@components/ui/accordion";
import { Input } from "@components/ui/input";
import { Textarea } from "@components/ui/textarea";

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
    const accordionValue = "inherited-env-vars";
    const filteredItems = search
        ? items.filter(item => {
              const q = search.toLowerCase().trim();
              return item.key.toLowerCase().includes(q) || item.value.toLowerCase().includes(q);
          })
        : items;

    return (
        <Accordion
            type="single"
            collapsible
            defaultValue={items.length > 0 ? accordionValue : undefined}
            className="w-full"
        >
            <AccordionItem value={accordionValue}>
                <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent">
                    {title}
                </AccordionTrigger>
                <AccordionContent className="pt-4 pb-0 pl-3">
                    {filteredItems.length > 0 ? (
                        <div className="flex flex-col gap-4">
                            <div className="space-y-3">
                                {filteredItems.map((item, index) => {
                                    const isMultilineValue = item.value.includes("\n");

                                    return (
                                        <div
                                            key={index}
                                            className="flex items-start gap-3"
                                        >
                                            {/* Key */}
                                            <div className="min-w-0 flex-1">
                                                <Input
                                                    value={item.key}
                                                    readOnly
                                                    disabled
                                                    placeholder="Key"
                                                    className="bg-muted cursor-default"
                                                />
                                            </div>

                                            {/* Value */}
                                            <div className="min-w-0 flex-1">
                                                {isRevealed && isMultilineValue ? (
                                                    <Textarea
                                                        value={item.value}
                                                        readOnly
                                                        disabled
                                                        placeholder="Value"
                                                        minRows={4}
                                                        maxRows={0}
                                                        className="bg-muted cursor-default resize-y"
                                                    />
                                                ) : (
                                                    <Input
                                                        type={isRevealed ? "text" : "password"}
                                                        value={item.value}
                                                        readOnly
                                                        disabled
                                                        placeholder="Value"
                                                        className="bg-muted cursor-default"
                                                    />
                                                )}
                                            </div>

                                            {/* Spacer to align with editable rows that have a multi-line toggle */}
                                            <div className="h-8 w-8" />

                                            {/* Literal badge */}
                                            <div className="flex h-9 items-center gap-2">
                                                <Checkbox
                                                    checked
                                                    disabled
                                                />
                                                <span className="text-sm cursor-pointer">Literal</span>
                                            </div>

                                            {/* Spacer to align with editable rows that have row action buttons */}
                                            <div className="h-8 w-8" />
                                        </div>
                                    );
                                })}
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

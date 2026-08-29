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
                <AccordionTrigger
                    headerClassName="sticky top-14 z-10"
                    className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent rounded-md shadow-xs"
                >
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
                                            className="flex flex-col sm:flex-row sm:items-start gap-2 sm:gap-3 p-2.5 sm:p-0 rounded-lg sm:rounded-none bg-muted/20 sm:bg-transparent border sm:border-0"
                                        >
                                            {/* Key & Value */}
                                            <div className="flex flex-1 items-start gap-2 sm:gap-3 min-w-0 w-full">
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
                                            </div>

                                            {/* Controls: Literal checkbox & spacers */}
                                            <div className="flex items-center justify-between sm:justify-end gap-2 shrink-0 w-full sm:w-auto">
                                                {/* Literal checkbox */}
                                                <div className="flex h-9 items-center gap-2">
                                                    <Checkbox
                                                        checked={item.isLiteral}
                                                        disabled
                                                    />
                                                    <span className="text-sm text-muted-foreground cursor-default select-none">
                                                        Literal
                                                    </span>
                                                </div>

                                                {/* Hidden Spacers on mobile, visible on desktop */}
                                                <div className="hidden sm:block size-9" />
                                                <div className="hidden sm:block size-9" />
                                            </div>
                                        </div>
                                    );
                                })}
                            </div>
                        </div>
                    ) : (
                        <div className="text-sm text-muted-foreground py-4 text-center">
                            {search ? "No records match your search" : "No results"}
                        </div>
                    )}
                </AccordionContent>
            </AccordionItem>
        </Accordion>
    );
}

export const InheritedEnvVarsAccordion = React.memo(View);

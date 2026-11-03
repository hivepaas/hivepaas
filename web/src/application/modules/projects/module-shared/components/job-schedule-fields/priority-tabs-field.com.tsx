import React from "react";

import { EAppScheduledJobTaskPriority } from "~/projects/module-shared/enums";

import { Tabs, TabsList, TabsTrigger } from "@/components/ui";

function View({ value, onChange, readOnly = false }: Props) {
    return (
        <Tabs
            value={value}
            onValueChange={onChange}
        >
            <TabsList>
                <TabsTrigger
                    value={EAppScheduledJobTaskPriority.Default}
                    disabled={readOnly}
                >
                    Default
                </TabsTrigger>
                <TabsTrigger
                    value={EAppScheduledJobTaskPriority.Critical}
                    disabled={readOnly}
                >
                    Critical
                </TabsTrigger>
                <TabsTrigger
                    value={EAppScheduledJobTaskPriority.Low}
                    disabled={readOnly}
                >
                    Low
                </TabsTrigger>
            </TabsList>
        </Tabs>
    );
}

interface Props {
    value: EAppScheduledJobTaskPriority;
    onChange: (value: string) => void;
    readOnly?: boolean;
}

export const PriorityTabsField = React.memo(View);

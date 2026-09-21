import { Database, Globe, HardDrive, Zap } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";
import { type EAppCategory } from "~/projects/domain";
import { type OptionCard, OptionCardGroup } from "~/projects/module-shared/components";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { type AppConfigKindSettingsFormSchemaInput, type AppConfigKindSettingsFormSchemaOutput } from "../schemas";

const CATEGORY_OPTIONS: OptionCard<EAppCategory>[] = [
    {
        value: "webapp",
        label: "Web App",
        description: "HTTP web service, API server, or microservice",
        icon: Globe,
    },
    {
        value: "database",
        label: "Database",
        description: "Relational or document database (Postgres, MySQL, Mongo)",
        icon: Database,
    },
    {
        value: "cache",
        label: "Cache Store",
        description: "In-memory key-value cache (Redis, KeyDB, Dragonfly)",
        icon: Zap,
    },
    {
        value: "storage",
        label: "Object Storage",
        description: "S3-compatible object storage or file service (MinIO)",
        icon: HardDrive,
    },
];

interface Props {
    readOnly?: boolean;
}

export function CategorySelector({ readOnly = false }: Props) {
    const { control } = useFormContext<
        AppConfigKindSettingsFormSchemaInput,
        unknown,
        AppConfigKindSettingsFormSchemaOutput
    >();

    const { field } = useController({ control, name: "category" });

    return (
        <InfoBlock
            titleWidth={220}
            title={
                <LabelWithInfo
                    label="App Category"
                    isRequired
                    content="Select the workload category for this application to configure kind-specific settings."
                />
            }
        >
            <OptionCardGroup
                options={CATEGORY_OPTIONS}
                value={field.value}
                onChange={field.onChange}
                readOnly={readOnly}
                className="grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 max-w-[800px]"
            />
        </InfoBlock>
    );
}

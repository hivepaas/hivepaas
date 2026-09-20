import { cn } from "@lib/utils";
import { Database, Globe, HardDrive, Zap } from "lucide-react";
import { useController, useFormContext } from "react-hook-form";
import { type EAppCategory } from "~/projects/domain";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { type AppConfigKindSettingsFormSchemaInput, type AppConfigKindSettingsFormSchemaOutput } from "../schemas";

interface CategoryOption {
    value: EAppCategory;
    label: string;
    description: string;
    icon: typeof Globe;
}

const CATEGORY_OPTIONS: CategoryOption[] = [
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
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 w-full max-w-[800px]">
                {CATEGORY_OPTIONS.map(opt => {
                    const isSelected = field.value === opt.value;
                    const Icon = opt.icon;

                    return (
                        <div
                            key={opt.value}
                            role="button"
                            tabIndex={readOnly ? -1 : 0}
                            aria-pressed={isSelected}
                            onClick={() => {
                                if (!readOnly) {
                                    field.onChange(opt.value);
                                }
                            }}
                            onKeyDown={e => {
                                if (!readOnly && (e.key === "Enter" || e.key === " ")) {
                                    e.preventDefault();
                                    field.onChange(opt.value);
                                }
                            }}
                            className={cn(
                                "flex flex-col items-start gap-1.5 p-3 rounded-lg border text-left transition-all select-none",
                                readOnly && "opacity-70 pointer-events-none cursor-not-allowed",
                                !readOnly && "cursor-pointer",
                                isSelected
                                    ? "border-primary bg-primary/5 ring-1 ring-primary text-foreground"
                                    : "border-border hover:border-primary/50 hover:bg-accent/40 text-muted-foreground",
                            )}
                        >
                            <div className="flex items-center gap-2">
                                <Icon
                                    className={cn(
                                        "size-4 shrink-0",
                                        isSelected ? "text-primary" : "text-muted-foreground",
                                    )}
                                />
                                <span
                                    className={cn(
                                        "font-medium text-sm",
                                        isSelected ? "text-foreground" : "text-muted-foreground",
                                    )}
                                >
                                    {opt.label}
                                </span>
                            </div>
                            <p className="text-xs text-muted-foreground line-clamp-2 leading-relaxed">
                                {opt.description}
                            </p>
                        </div>
                    );
                })}
            </div>
        </InfoBlock>
    );
}

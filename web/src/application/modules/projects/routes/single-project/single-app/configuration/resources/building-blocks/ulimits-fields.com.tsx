import { useState } from "react";

import { Button, Input } from "@components/ui";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext } from "react-hook-form";

import { InfoBlock, InputNumberWithAddon, LabelWithInfo } from "@application/shared/components";

import { type AppConfigResourcesFormSchemaInput, type AppConfigResourcesFormSchemaOutput } from "../schemas";

const ULIMIT_NAMES = [
    "core",
    "cpu",
    "data",
    "fsize",
    "locks",
    "memlock",
    "msgqueue",
    "nice",
    "nofile",
    "nproc",
    "rss",
    "rtprio",
    "rttime",
    "sigpending",
    "stack",
] as const;

export function UlimitsFields() {
    const { control } = useFormContext<
        AppConfigResourcesFormSchemaInput,
        unknown,
        AppConfigResourcesFormSchemaOutput
    >();

    const { fields, append, remove } = useFieldArray({
        control,
        name: "ulimits",
    });

    const [name, setName] = useState<string>(ULIMIT_NAMES[0]);
    const [soft, setSoft] = useState("");
    const [hard, setHard] = useState("");

    const handleAdd = () => {
        if (!name || !soft.trim() || !hard.trim()) return;
        append({
            name,
            soft: Number(soft),
            hard: Number(hard),
        });
        setSoft("");
        setHard("");
    };

    return (
        <div className="flex flex-col gap-6 px-2">
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Ulimits"
                        content="Per-process resource limits for the container (e.g. nofile, nproc)."
                    />
                }
            >
                <div className="flex flex-col gap-3 max-w-[590px]">
                    <div className="flex gap-3 items-center">
                        <div className="grid grid-cols-10 flex-1 gap-3">
                            <div className="col-span-4 flex items-center rounded-md border border-input h-9 flex-1">
                                <span className="px-3 text-sm border-r border-input bg-muted/50 h-full flex items-center">
                                    Name
                                </span>
                                <Select
                                    value={name}
                                    onValueChange={setName}
                                >
                                    <SelectTrigger className="flex-1 border-0 shadow-none focus:ring-0 rounded-l-none h-9">
                                        <SelectValue placeholder="nofile" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {ULIMIT_NAMES.map(n => (
                                            <SelectItem
                                                key={n}
                                                value={n}
                                            >
                                                {n}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                            <InputNumberWithAddon
                                addonLeft="Soft"
                                value={Number(soft)}
                                onChange={e => {
                                    setSoft(e.target.value);
                                }}
                                placeholder="1024"
                                classNameContainer="col-span-3"
                            />
                            <InputNumberWithAddon
                                addonLeft="Hard"
                                value={Number(hard)}
                                onChange={e => {
                                    setHard(e.target.value);
                                }}
                                placeholder="1024"
                                classNameContainer="col-span-3"
                            />
                        </div>
                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleAdd}
                            disabled={!soft.trim() || !hard.trim()}
                        >
                            <Plus className="size-4" /> Add
                        </Button>
                    </div>

                    <div className="divide-y divide-zinc-200">
                        {fields.map((field, index) => (
                            <div
                                key={field.id}
                                className="flex items-center gap-3 py-2"
                            >
                                <div className="grid grid-cols-10 flex-1 gap-3">
                                    <Input
                                        value={field.name}
                                        disabled
                                        className="col-span-4"
                                    />
                                    <Input
                                        value={String(field.soft)}
                                        disabled
                                        className="col-span-3"
                                    />
                                    <Input
                                        value={String(field.hard)}
                                        disabled
                                        className="col-span-3"
                                    />
                                </div>
                                <div className="w-[76px]">
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="icon"
                                        onClick={() => {
                                            remove(index);
                                        }}
                                    >
                                        <Trash2 className="size-4" />
                                    </Button>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </InfoBlock>
        </div>
    );
}

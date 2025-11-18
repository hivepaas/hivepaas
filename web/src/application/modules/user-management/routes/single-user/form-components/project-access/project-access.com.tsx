import { useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { Button } from "@components/ui/button";
import { Checkbox } from "@components/ui/checkbox";
import { Plus, Trash2 } from "lucide-react";
import { useFieldArray, useFormContext } from "react-hook-form";

import { Combobox, type ComboboxOption } from "@application/shared/components";
import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ProjectsPublicQueries } from "@application/shared/data-public/queries";
import type { ProjectPublic } from "@application/shared/entities";

import { type SingleUserFormSchemaInput, type SingleUserFormSchemaOutput } from "../../schemas";

export function ProjectAccess() {
    const { control, formState } = useFormContext<SingleUserFormSchemaInput, unknown, SingleUserFormSchemaOutput>();

    const { fields, append, update, remove } = useFieldArray({
        control,
        name: "projectAccess",
    });

    const [searchQuery, setSearchQuery] = useState("");
    const [selectedProject, setSelectedProject] = useState<ProjectPublic | null>(null);

    const { data: projectsData, isFetching } = ProjectsPublicQueries.useFindManyPaginated({
        search: searchQuery,
    });

    const projects = projectsData?.data ?? [];
    const selectedProjectIds = fields.map(f => f.id);
    const availableProjects = projects.filter(p => !selectedProjectIds.includes(p.id));

    const comboboxOptions: ComboboxOption<{ id: string; name: string }>[] = useMemo(
        () =>
            availableProjects.map(project => ({
                value: {
                    id: project.id,
                    name: project.name,
                },
                label: project.name,
            })),
        [availableProjects],
    );

    const handleAddProject = (project: ProjectPublic | null) => {
        if (!project) {
            return;
        }

        const newAccess = {
            id: project.id,
            name: project.name,
            access: {
                read: true,
                write: false,
                delete: false,
            },
        };

        append(newAccess);
        setSelectedProject(null);
        setSearchQuery("");
    };

    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Project access"
                    content="Project access description"
                />
            }
        >
            <div className="flex flex-col gap-4">
                {/* Project Selection */}
                <div className="flex items-center gap-2 max-w-[400px]">
                    <Combobox
                        options={comboboxOptions}
                        value={selectedProject?.id ?? null}
                        onChange={(value, option) => {
                            setSelectedProject(option ?? null);
                        }}
                        onSearch={setSearchQuery}
                        placeholder="Select Project(s)"
                        searchable
                        closeOnSelect={false}
                        emptyText="No projects available"
                        className="flex-1"
                        valueKey="id"
                        aria-invalid={!!formState.errors.projectAccess}
                        loading={isFetching}
                    />
                    <Button
                        type="button"
                        onClick={() => {
                            handleAddProject(selectedProject);
                        }}
                        disabled={!selectedProject || selectedProjectIds.includes(selectedProject.id)}
                    >
                        <Plus className="mr-2 h-4 w-4" />
                        Add
                    </Button>
                </div>

                {/* Project List */}
                <div className="divider-y">
                    {fields.length > 0 && (
                        <div className="space-y-0">
                            {fields.map((project, index) => (
                                <div
                                    key={project.id}
                                    className={cn(
                                        "flex items-center gap-4 p-3",
                                        index !== fields.length - 1 && "border-b",
                                    )}
                                >
                                    <div className="flex-1 font-semibold">{project.name}</div>
                                    <div className="flex items-center gap-4">
                                        <div className="flex items-center gap-2">
                                            <Checkbox
                                                checked={project.access.read}
                                                disabled
                                            />
                                            <label
                                                htmlFor="read"
                                                className="text-sm"
                                            >
                                                Read
                                            </label>
                                        </div>
                                        <div className="flex items-center gap-2">
                                            <Checkbox
                                                checked={project.access.write}
                                                onCheckedChange={checked => {
                                                    update(index, {
                                                        ...project,
                                                        access: { ...project.access, write: checked === true },
                                                    });
                                                }}
                                            />
                                            <label
                                                htmlFor="write"
                                                className="text-sm"
                                            >
                                                Write
                                            </label>
                                        </div>
                                        <div className="flex items-center gap-2">
                                            <Checkbox
                                                checked={project.access.delete}
                                                onCheckedChange={checked => {
                                                    update(index, {
                                                        ...project,
                                                        access: { ...project.access, delete: checked === true },
                                                    });
                                                }}
                                            />
                                            <label
                                                htmlFor="delete"
                                                className="text-sm"
                                            >
                                                Delete
                                            </label>
                                        </div>
                                        <Button
                                            type="button"
                                            variant="link"
                                            className="text-destructive hover:opacity-80"
                                            size="icon"
                                            onClick={() => {
                                                remove(index);
                                            }}
                                        >
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </InfoBlock>
    );
}

import React, { useMemo, useState } from "react";

import { Button, Checkbox } from "@components/ui";
import { Plus, Trash2 } from "lucide-react";
import { type Path, useFieldArray, useFormContext } from "react-hook-form";

import { Combobox, type ComboboxOption, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { ProjectsPublicQueries } from "@application/shared/data-public/queries";
import type { ProjectPublic } from "@application/shared/entities";

interface ProjectAccess {
    id: string;
    name: string;
    access: {
        read: boolean;
        write: boolean;
        delete: boolean;
    };
}

function View<T>({ name }: Props<T>) {
    const { control, formState } = useFormContext<Record<string, ProjectAccess[]>>();

    const { fields, append, update, remove } = useFieldArray({
        control,
        name: name as string,
    });

    const [searchQuery, setSearchQuery] = useState("");
    const [selectedProject, setSelectedProject] = useState<ProjectPublic | null>(null);

    const { data: { data: projects } = DEFAULT_PAGINATED_DATA, isFetching } =
        ProjectsPublicQueries.useFindManyPaginated({
            search: searchQuery,
        });

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
                    aria-invalid={!!formState.errors[name as string]}
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
            <div>
                {fields.length > 0 && (
                    <div className="space-y-0 divide-y">
                        {fields.map((project, index) => (
                            <div
                                key={project.id}
                                className="flex items-center gap-4 p-3"
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
    );
}

interface Props<T> {
    name: Path<T>;
}

export const ProjectAccess = React.memo(View) as typeof View;

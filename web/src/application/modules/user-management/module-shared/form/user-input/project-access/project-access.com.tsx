import React, { useMemo, useState } from "react";

import { Button, Checkbox } from "@components/ui";
import { Plus, Trash2 } from "lucide-react";
import { type Path, useFieldArray, useFormContext, useWatch } from "react-hook-form";

import { Combobox, type ComboboxOption } from "@application/shared/components";
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

function View<T>({ name, isAdmin = false }: Props<T>) {
    const { control, formState } = useFormContext<Record<string, ProjectAccess[]>>();

    const { fields, append, update, remove } = useFieldArray({
        control,
        name: name as string,
        keyName: "_id",
    });

    const watchedFields = useWatch({
        control,
        name: name as string,
    });

    const [searchQuery, setSearchQuery] = useState("");
    const [selectedProject, setSelectedProject] = useState<ProjectPublic | null>(null);

    const { data: { data: projects } = DEFAULT_PAGINATED_DATA, isFetching } =
        ProjectsPublicQueries.useFindManyPaginated({
            search: isAdmin ? "" : searchQuery,
        });

    const comboboxOptions: ComboboxOption<{ id: string; name: string }>[] = useMemo(() => {
        if (isAdmin) {
            return [];
        }
        const availableProjects = projects.filter(p => watchedFields.every(f => f.id !== p.id));
        return availableProjects.map(project => ({
            value: {
                id: project.id,
                name: project.name,
            },
            label: project.name,
        }));
    }, [projects, watchedFields, isAdmin]);

    const allProjectsForAdmin = useMemo(() => {
        if (!isAdmin) {
            return [];
        }
        return projects.map(project => ({
            id: project.id,
            name: project.name,
            access: {
                read: true,
                write: true,
                delete: true,
            },
        }));
    }, [projects, isAdmin]);

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

    const projectsToDisplay = isAdmin ? allProjectsForAdmin : fields;

    return (
        <div className="flex flex-col gap-4">
            {/* Project Selection */}
            {!isAdmin && (
                <div className="flex items-center gap-2">
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
                        className="flex-1 md:max-w-[400px]"
                        valueKey="id"
                        aria-invalid={!!formState.errors[name as string]}
                        loading={isFetching}
                    />
                    <Button
                        type="button"
                        onClick={() => {
                            handleAddProject(selectedProject);
                        }}
                        disabled={!selectedProject || watchedFields.some(f => f.id === selectedProject.id)}
                    >
                        <Plus className="size-4" />
                        Add
                    </Button>
                </div>
            )}

            {/* Project List */}
            <div>
                {projectsToDisplay.length > 0 && (
                    <div className="space-y-0 divide-y">
                        {projectsToDisplay.map((project, index) => (
                            <div
                                key={project.id}
                                className="flex items-center flex-wrap gap-4 p-3"
                            >
                                <div className="flex-1 font-semibold">{project.name}</div>
                                <div className="flex items-center gap-4">
                                    <div className="flex items-center gap-2">
                                        <Checkbox
                                            checked={project.access.read}
                                            disabled={isAdmin}
                                        />
                                        <label
                                            htmlFor={`${project.id}-read`}
                                            className="text-sm"
                                        >
                                            Read
                                        </label>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <Checkbox
                                            checked={project.access.write}
                                            disabled={isAdmin}
                                            onCheckedChange={checked => {
                                                if (!isAdmin) {
                                                    update(index, {
                                                        ...project,
                                                        access: { ...project.access, write: checked === true },
                                                    });
                                                }
                                            }}
                                        />
                                        <label
                                            htmlFor={`${project.id}-write`}
                                            className="text-sm"
                                        >
                                            Write
                                        </label>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <Checkbox
                                            checked={project.access.delete}
                                            disabled={isAdmin}
                                            onCheckedChange={checked => {
                                                if (!isAdmin) {
                                                    update(index, {
                                                        ...project,
                                                        access: { ...project.access, delete: checked === true },
                                                    });
                                                }
                                            }}
                                        />
                                        <label
                                            htmlFor={`${project.id}-delete`}
                                            className="text-sm"
                                        >
                                            Delete
                                        </label>
                                    </div>
                                    {!isAdmin && (
                                        <Button
                                            type="button"
                                            variant="link"
                                            className="text-destructive hover:opacity-80"
                                            size="icon"
                                            onClick={() => {
                                                remove(index);
                                            }}
                                        >
                                            <Trash2 className="size-4" />
                                        </Button>
                                    )}
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
    isAdmin?: boolean;
}

export const ProjectAccess = React.memo(View) as typeof View;

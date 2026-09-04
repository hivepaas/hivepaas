import React, { useMemo, useState } from "react";

import { Button, Checkbox } from "@components/ui";
import { CheckCheck, Plus, Trash2 } from "lucide-react";
import { type Path, useFieldArray, useFormContext, useWatch } from "react-hook-form";
import { ProjectEnvBadge } from "~/projects/module-shared/components";
import type {
    AccessActions,
    ProjectAccess as ProjectAccessValue,
    ProjectEnvAccess,
} from "~/user-management/module-shared/schemas";

import { Combobox, type ComboboxOption } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { ProjectsPublicQueries } from "@application/shared/data-public/queries";
import type { ProjectPublic } from "@application/shared/entities";

const NO_ACCESS: AccessActions = { read: false, execute: false, write: false, delete: false };
const READ_ONLY_ACCESS: AccessActions = { read: true, execute: false, write: false, delete: false };
const FULL_ACCESS: AccessActions = { read: true, execute: true, write: true, delete: true };

type ToggleableAction = Exclude<keyof AccessActions, "read">;

const TOGGLEABLE_ACTIONS: { key: ToggleableAction; label: string }[] = [
    { key: "execute", label: "Execute" },
    { key: "write", label: "Write" },
    { key: "delete", label: "Delete" },
];

/** An env is granted when it carries any permission at all. */
function isGranted(access: AccessActions) {
    return access.read || access.execute || access.write || access.delete;
}

function AccessCheckbox({ id, label, checked, disabled, onCheckedChange }: AccessCheckboxProps) {
    return (
        <div className="flex items-center gap-2">
            <Checkbox
                id={id}
                checked={checked}
                disabled={disabled}
                onCheckedChange={onCheckedChange}
            />
            <label
                htmlFor={id}
                className="text-sm"
            >
                {label}
            </label>
        </div>
    );
}

function EnvAccessRow({ envAccess, disabled, onChangeAccess }: EnvAccessRowProps) {
    const { access } = envAccess;
    const granted = isGranted(access);
    const allToggled = TOGGLEABLE_ACTIONS.every(({ key }) => access[key]);

    return (
        <div className="flex flex-wrap items-center gap-4 py-2">
            <div className="flex min-w-0 flex-1 items-center">
                <ProjectEnvBadge
                    name={envAccess.name ? `Env: ${envAccess.name}` : ""}
                    color={envAccess.color}
                    className="max-w-[22ch] whitespace-normal wrap-break-word text-left"
                />
            </div>
            <div className="flex items-center gap-4">
                {/* Read is implied by granting anything at all, so it is never edited directly */}
                <AccessCheckbox
                    id={`${envAccess.id}-read`}
                    label="Read"
                    checked={access.read}
                    disabled
                />
                {TOGGLEABLE_ACTIONS.map(({ key, label }) => (
                    <AccessCheckbox
                        key={key}
                        id={`${envAccess.id}-${key}`}
                        label={label}
                        checked={access[key]}
                        disabled={disabled}
                        onCheckedChange={checked => {
                            if (disabled) {
                                return;
                            }
                            // Granting any action implies read access.
                            onChangeAccess({ ...access, read: true, [key]: checked === true });
                        }}
                    />
                ))}
                {disabled ? (
                    <div className="size-7" />
                ) : (
                    <div className="flex items-center gap-1">
                        <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="size-7 text-zinc-400 hover:text-primary hover:bg-primary/10 rounded-md"
                            title={allToggled ? "Keep read access only" : "Grant every permission"}
                            onClick={() => {
                                onChangeAccess(allToggled ? READ_ONLY_ACCESS : FULL_ACCESS);
                            }}
                        >
                            <CheckCheck className="size-4" />
                        </Button>
                        <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="size-7 text-zinc-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-950/50 dark:hover:text-red-400 rounded-md"
                            title="Revoke access to this environment"
                            disabled={!granted}
                            onClick={() => {
                                onChangeAccess(NO_ACCESS);
                            }}
                        >
                            <Trash2 className="size-4" />
                        </Button>
                    </div>
                )}
            </div>
        </div>
    );
}

function AllProjectsRow() {
    return (
        <div className="flex flex-wrap items-center gap-4 py-2">
            <div className="flex-1 font-semibold">All projects</div>
            <div className="flex items-center gap-4">
                <AccessCheckbox
                    id="all-projects-read"
                    label="Read"
                    checked
                    disabled
                />
                {TOGGLEABLE_ACTIONS.map(({ key, label }) => (
                    <AccessCheckbox
                        key={key}
                        id={`all-projects-${key}`}
                        label={label}
                        checked
                        disabled
                    />
                ))}
            </div>
        </div>
    );
}

function View<T>({ name, isAdmin = false, disabled = false }: Props<T>) {
    const { control, formState } = useFormContext<Record<string, ProjectAccessValue[]>>();

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

    const comboboxOptions: ComboboxOption<ProjectPublic>[] = useMemo(() => {
        if (isAdmin) {
            return [];
        }
        const availableProjects = projects.filter(p => watchedFields.every(f => f.id !== p.id));
        return availableProjects.map(project => ({
            value: project,
            label: project.name,
        }));
    }, [projects, watchedFields, isAdmin]);

    const handleAddProject = (project: ProjectPublic | null) => {
        if (!project) {
            return;
        }
        // Every env of the project is listed so the whole matrix is editable in
        // place; they start with read access, like a freshly granted project did.
        append({
            id: project.id,
            name: project.name,
            envAccesses: project.envs.map(env => ({
                id: env.id,
                name: env.name,
                color: env.color,
                access: { ...READ_ONLY_ACCESS },
            })),
        });
        setSelectedProject(null);
        setSearchQuery("");
    };

    const handleChangeEnvAccess = (projectIndex: number, envIndex: number, access: AccessActions) => {
        const projectAccess = watchedFields[projectIndex];
        if (!projectAccess) {
            return;
        }
        update(projectIndex, {
            ...projectAccess,
            envAccesses: projectAccess.envAccesses.map((envAccess, idx) =>
                idx === envIndex ? { ...envAccess, access } : envAccess,
            ),
        });
    };

    return (
        <div className="flex flex-col gap-4">
            {/* Project Selection */}
            {!isAdmin && !disabled && (
                <div className="flex items-center gap-2">
                    <Combobox
                        options={comboboxOptions}
                        value={selectedProject?.id ?? null}
                        onChange={(_value, option) => {
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

            <div>
                {isAdmin ? (
                    <div className="space-y-0 divide-y divide-border/50">
                        <AllProjectsRow />
                    </div>
                ) : fields.length > 0 ? (
                    <div className="flex flex-col gap-4">
                        {fields.map((projectAccess, projectIndex) => (
                            <div key={projectAccess._id}>
                                {/* Project header - the project itself carries no access */}
                                <div className="flex items-center gap-2 border-b border-border/50 pb-2">
                                    <div className="min-w-0 flex-1 truncate font-semibold">{projectAccess.name}</div>
                                    {!disabled && (
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            className="size-7 text-zinc-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-950/50 dark:hover:text-red-400 rounded-md"
                                            title="Remove this project"
                                            onClick={() => {
                                                remove(projectIndex);
                                            }}
                                        >
                                            <Trash2 className="size-4" />
                                        </Button>
                                    )}
                                </div>

                                {projectAccess.envAccesses.length > 0 ? (
                                    <div className="space-y-0 divide-y divide-border/50 pl-4">
                                        {projectAccess.envAccesses.map((envAccess, envIndex) => (
                                            <EnvAccessRow
                                                key={envAccess.id}
                                                envAccess={envAccess}
                                                disabled={disabled}
                                                onChangeAccess={access => {
                                                    handleChangeEnvAccess(projectIndex, envIndex, access);
                                                }}
                                            />
                                        ))}
                                    </div>
                                ) : (
                                    <p className="py-2 pl-4 text-sm text-muted-foreground">
                                        This project has no environment yet.
                                    </p>
                                )}
                            </div>
                        ))}
                    </div>
                ) : null}
            </div>
        </div>
    );
}

interface AccessCheckboxProps {
    id: string;
    label: string;
    checked: boolean;
    disabled?: boolean;
    onCheckedChange?: (checked: boolean | "indeterminate") => void;
}

interface EnvAccessRowProps {
    envAccess: ProjectEnvAccess;
    disabled?: boolean;
    onChangeAccess: (access: AccessActions) => void;
}

interface Props<T> {
    name: Path<T>;
    isAdmin?: boolean;
    disabled?: boolean;
}

export const ProjectAccess = React.memo(View) as typeof View;

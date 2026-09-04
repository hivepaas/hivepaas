import React, { useMemo, useState } from "react";

import { Button, Checkbox } from "@components/ui";
import { CheckCheck, Plus, Trash2 } from "lucide-react";
import { type Path, useFieldArray, useFormContext, useWatch } from "react-hook-form";
import { ProjectEnvBadge } from "~/projects/module-shared/components";
import { type AccessAction, useGrantableAccess } from "~/user-management/module-shared/hooks";
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

// A user may only hand out permissions they hold themselves; the API enforces it
// and the UI puts the rest visibly out of reach.
const NO_REACH_HINT = "You cannot grant a permission you do not have yourself";

const ACTIONS: { key: AccessAction; label: string }[] = [
    { key: "read", label: "Read" },
    { key: "execute", label: "Execute" },
    { key: "write", label: "Write" },
    { key: "delete", label: "Delete" },
];

function AccessCheckbox({ id, label, checked, disabled, title, onCheckedChange }: AccessCheckboxProps) {
    return (
        <div
            className="flex items-center gap-1.5 shrink-0"
            title={title}
        >
            <Checkbox
                id={id}
                checked={checked}
                disabled={disabled}
                onCheckedChange={onCheckedChange}
            />
            <label
                htmlFor={id}
                className="text-sm select-none cursor-pointer"
            >
                {label}
            </label>
        </div>
    );
}

function EnvAccessRow({ envAccess, disabled = false, canGrant, onChangeAccess }: EnvAccessRowProps) {
    const { access } = envAccess;

    // Select-all only reaches the actions this account may actually change; the
    // others keep whatever an admin granted.
    const grantableKeys = ACTIONS.filter(({ key }) => canGrant(key)).map(({ key }) => key);
    const allSelected = grantableKeys.length > 0 && grantableKeys.every(key => access[key]);

    const toggleAll = () => {
        const next = { ...access };
        for (const key of grantableKeys) {
            next[key] = !allSelected;
        }
        onChangeAccess(next);
    };

    return (
        <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-2 py-2">
            <div className="flex min-w-0 items-center">
                <ProjectEnvBadge
                    name={envAccess.name ? `Env: ${envAccess.name}` : ""}
                    color={envAccess.color}
                    className="max-w-[22ch] whitespace-normal wrap-break-word text-left"
                />
            </div>
            <div className="flex flex-wrap items-center gap-x-2.5 sm:gap-x-3.5 gap-y-1.5">
                {ACTIONS.map(({ key, label }) => (
                    <AccessCheckbox
                        key={key}
                        id={`${envAccess.id}-${key}`}
                        label={label}
                        checked={access[key]}
                        disabled={disabled || !canGrant(key)}
                        title={canGrant(key) ? undefined : NO_REACH_HINT}
                        onCheckedChange={checked => {
                            if (disabled || !canGrant(key)) {
                                return;
                            }
                            onChangeAccess({ ...access, [key]: checked === true });
                        }}
                    />
                ))}
                {!disabled && (
                    <div className="flex items-center gap-1 shrink-0">
                        <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="size-7 text-zinc-400 hover:text-primary hover:bg-primary/10 rounded-md"
                            title={
                                grantableKeys.length === 0 ? NO_REACH_HINT : allSelected ? "Deselect all" : "Select all"
                            }
                            disabled={grantableKeys.length === 0}
                            onClick={toggleAll}
                        >
                            <CheckCheck className="size-4" />
                        </Button>
                    </div>
                )}
            </div>
        </div>
    );
}

function AllProjectsRow() {
    return (
        <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-2 py-2">
            <div className="font-semibold">All projects</div>
            <div className="flex flex-wrap items-center gap-x-2.5 sm:gap-x-3.5 gap-y-1.5">
                {ACTIONS.map(({ key, label }) => (
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
    const { canGrantEnv } = useGrantableAccess();

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
                // Read is the usual starting point, but not on an env this account
                // cannot grant read on - that would be rejected on save.
                access: canGrantEnv(env.id, "read") ? { ...READ_ONLY_ACCESS } : { ...NO_ACCESS },
            })),
        });
        setSelectedProject(null);
        setSearchQuery("");
    };

    // The actions this account may change on one env; the rest keep whatever an
    // admin granted and are never touched by the select-all buttons.
    const grantableEnvKeys = (envAccess: ProjectEnvAccess) =>
        ACTIONS.filter(({ key }) => canGrantEnv(envAccess.id, key)).map(({ key }) => key);

    // A project counts as fully selected once every env has all of its reachable
    // actions on. A project with nothing reachable is reported as not selected so
    // the button stays a no-op.
    const isProjectAllSelected = (projectAccess: ProjectAccessValue) => {
        let hasReachableEnv = false;
        for (const envAccess of projectAccess.envAccesses) {
            const keys = grantableEnvKeys(envAccess);
            if (keys.length === 0) {
                continue;
            }
            hasReachableEnv = true;
            if (!keys.every(key => envAccess.access[key])) {
                return false;
            }
        }
        return hasReachableEnv;
    };

    const handleToggleProject = (projectIndex: number) => {
        const projectAccess = watchedFields[projectIndex];
        if (!projectAccess) {
            return;
        }
        const select = !isProjectAllSelected(projectAccess);

        update(projectIndex, {
            ...projectAccess,
            envAccesses: projectAccess.envAccesses.map(envAccess => {
                const keys = grantableEnvKeys(envAccess);
                if (keys.length === 0) {
                    return envAccess;
                }
                const access = { ...envAccess.access };
                for (const key of keys) {
                    access[key] = select;
                }
                return { ...envAccess, access };
            }),
        });
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
                        {fields.map((projectAccess, projectIndex) => {
                            const projectAllSelected = isProjectAllSelected(projectAccess);
                            const projectHasReachableAction = projectAccess.envAccesses.some(
                                envAccess => grantableEnvKeys(envAccess).length > 0,
                            );

                            return (
                                <div key={projectAccess._id}>
                                    {/* Project header - the project itself carries no access */}
                                    <div className="flex items-center gap-2 border-b border-border/50 pb-2">
                                        <div className="min-w-0 flex-1 truncate font-semibold">
                                            {projectAccess.name}
                                        </div>
                                        {!disabled && (
                                            <div className="flex items-center gap-1 shrink-0">
                                                <Button
                                                    type="button"
                                                    variant="ghost"
                                                    size="icon"
                                                    className="size-7 text-zinc-400 hover:text-primary hover:bg-primary/10 rounded-md"
                                                    title={
                                                        projectHasReachableAction
                                                            ? projectAllSelected
                                                                ? "Deselect all environments"
                                                                : "Select all environments"
                                                            : NO_REACH_HINT
                                                    }
                                                    disabled={!projectHasReachableAction}
                                                    onClick={() => {
                                                        handleToggleProject(projectIndex);
                                                    }}
                                                >
                                                    <CheckCheck className="size-4" />
                                                </Button>
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
                                            </div>
                                        )}
                                    </div>

                                    {projectAccess.envAccesses.length > 0 ? (
                                        <div className="space-y-0 divide-y divide-border/50 pl-2 sm:pl-4">
                                            {projectAccess.envAccesses.map((envAccess, envIndex) => (
                                                <EnvAccessRow
                                                    key={envAccess.id}
                                                    envAccess={envAccess}
                                                    disabled={disabled}
                                                    canGrant={action => canGrantEnv(envAccess.id, action)}
                                                    onChangeAccess={access => {
                                                        handleChangeEnvAccess(projectIndex, envIndex, access);
                                                    }}
                                                />
                                            ))}
                                        </div>
                                    ) : (
                                        <p className="py-2 pl-2 sm:pl-4 text-sm text-muted-foreground">
                                            This project has no environment yet.
                                        </p>
                                    )}
                                </div>
                            );
                        })}
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
    title?: string;
    onCheckedChange?: (checked: boolean | "indeterminate") => void;
}

interface EnvAccessRowProps {
    envAccess: ProjectEnvAccess;
    disabled?: boolean;
    canGrant: (action: AccessAction) => boolean;
    onChangeAccess: (access: AccessActions) => void;
}

interface Props<T> {
    name: Path<T>;
    isAdmin?: boolean;
    disabled?: boolean;
}

export const ProjectAccess = React.memo(View) as typeof View;

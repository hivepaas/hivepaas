import { useCallback, useEffect, useMemo, useState } from "react";

import { Avatar, Button, Checkbox } from "@components/ui";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { CheckCheck, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import type { EnvUserAccessesData, ProjectUserAccessActions, ProjectUserAccessBase } from "~/projects/api/services";
import { ProjectUserAccessesCommands } from "~/projects/data/commands";
import { ProjectUserAccessesQueries } from "~/projects/data/queries";
import type { ProjectEnvEntity } from "~/projects/domain";
import { ProjectEnvBadge } from "~/projects/module-shared/components";

import { AppLoader, Combobox, type ComboboxOption, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { UsersPublicQueries } from "@application/shared/data-public/queries";
import { EUserRole } from "@application/shared/enums";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { Separator } from "@/components/ui/separator";

import { useProjectUserAccessesDialogState } from "../hooks";

type UserAccessOption = Record<string, unknown> & {
    id: string;
    username: string;
    email: string;
    fullName: string;
    photo: string | null;
    role: EUserRole;
};

type EnvAccessesState = Record<string, ProjectUserAccessBase[]>;

function getUserDisplayName(user: Pick<ProjectUserAccessBase, "email" | "fullName" | "username">) {
    return user.fullName || user.email || user.username;
}

function createDefaultAccess(): ProjectUserAccessActions {
    return {
        read: true,
        execute: false,
        write: false,
        delete: false,
    };
}

function createUserFromOption(selectedUser: UserAccessOption): ProjectUserAccessBase {
    return {
        id: selectedUser.id,
        username: selectedUser.username,
        email: selectedUser.email,
        fullName: selectedUser.fullName,
        photo: selectedUser.photo,
        role: selectedUser.role,
        access: createDefaultAccess(),
    };
}

function buildInitialEnvAccesses(envs: ProjectEnvEntity[], envUserAccesses: EnvUserAccessesData[]): EnvAccessesState {
    const accessByName = new Map(envUserAccesses.map(envAccess => [envAccess.name, envAccess.userAccesses]));

    return Object.fromEntries(envs.map(env => [env.name, accessByName.get(env.name) ?? []]));
}

function UserInfo({ user }: UserInfoProps) {
    return (
        <div className="flex min-w-0 items-center gap-3">
            <Avatar
                name={getUserDisplayName(user)}
                src={user.photo}
                className="size-8"
                borderless
            />
            <div className="flex min-w-0 flex-col">
                <span className="truncate text-sm font-medium">{getUserDisplayName(user)}</span>
                <span className="truncate text-xs text-muted-foreground">{user.email || user.username}</span>
            </div>
        </div>
    );
}

function AccessCheckbox({ checked, disabled, id, label, onCheckedChange }: AccessCheckboxProps) {
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

function EnvAccessSection({
    env,
    users,
    selectedUser,
    userOptions,
    isFetchingUsers,
    canUpdateProjectAccess,
    onSelectUser,
    onSearchUser,
    onAdd,
    onToggleAll,
    onChangeAccess,
    onRemoveUser,
}: EnvAccessSectionProps) {
    const handleSearch = useCallback(
        (query: string) => {
            onSearchUser(env.name, query);
        },
        [env.name, onSearchUser],
    );

    const handleChange = useCallback(
        (_value: string | null, option: UserAccessOption | null) => {
            onSelectUser(env.name, option);
        },
        [env.name, onSelectUser],
    );

    return (
        <div>
            <Separator className="mb-4 opacity-50" />
            <InfoBlock
                title={
                    <div className="flex items-center gap-2">
                        Env:
                        <ProjectEnvBadge
                            name={env.name}
                            color={env.color}
                            className="max-w-[16ch] whitespace-normal wrap-break-word text-left"
                        />
                    </div>
                }
                titleWidth={180}
            >
                <div className="flex flex-col gap-4">
                    <div className="flex flex-wrap items-center gap-2">
                        <Combobox
                            options={userOptions}
                            value={selectedUser?.id ?? null}
                            onChange={handleChange}
                            onSearch={handleSearch}
                            placeholder="Select User(s)"
                            searchable
                            emptyText="No users available"
                            className="flex-1 md:max-w-[420px]"
                            valueKey="id"
                            loading={isFetchingUsers}
                            disabled={!canUpdateProjectAccess}
                        />
                        <PermissionTooltipAction
                            id={MODULE_IDS.Project}
                            action="write"
                        >
                            {({ isDenied }) => (
                                <Button
                                    type="button"
                                    variant="outline"
                                    disabled={isDenied || !selectedUser || !canUpdateProjectAccess}
                                    onClick={() => {
                                        onAdd(env.name);
                                    }}
                                >
                                    <Plus className="size-4" />
                                    Add
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    </div>

                    {users.length > 0 && (
                        <div className="divide-y">
                            {users.map(user => (
                                <EnvUserAccessRow
                                    key={user.id}
                                    envName={env.name}
                                    user={user}
                                    canUpdateProjectAccess={canUpdateProjectAccess}
                                    onToggleAll={userId => {
                                        onToggleAll(env.name, userId);
                                    }}
                                    onChangeAccess={(userId, key, checked) => {
                                        onChangeAccess(env.name, userId, key, checked);
                                    }}
                                    onRemove={userId => {
                                        onRemoveUser(env.name, userId);
                                    }}
                                />
                            ))}
                        </div>
                    )}
                </div>
            </InfoBlock>
        </div>
    );
}

function EnvUserAccessRow({
    envName,
    user,
    canUpdateProjectAccess,
    onToggleAll,
    onChangeAccess,
    onRemove,
}: EnvUserAccessRowProps) {
    return (
        <div className="flex flex-wrap items-center gap-4 py-3">
            <div className="min-w-[220px] flex-1">
                <UserInfo user={user} />
            </div>
            <div className="flex items-center gap-4">
                <AccessCheckbox
                    id={`${envName}-${user.id}-read`}
                    checked={user.access.read}
                    disabled
                    label="Read"
                />
                <AccessCheckbox
                    id={`${envName}-${user.id}-execute`}
                    checked={user.access.execute}
                    disabled={!canUpdateProjectAccess}
                    label="Execute"
                    onCheckedChange={checked => {
                        onChangeAccess(user.id, "execute", checked === true);
                    }}
                />
                <AccessCheckbox
                    id={`${envName}-${user.id}-write`}
                    checked={user.access.write}
                    disabled={!canUpdateProjectAccess}
                    label="Write"
                    onCheckedChange={checked => {
                        onChangeAccess(user.id, "write", checked === true);
                    }}
                />
                <AccessCheckbox
                    id={`${envName}-${user.id}-delete`}
                    checked={user.access.delete}
                    disabled={!canUpdateProjectAccess}
                    label="Delete"
                    onCheckedChange={checked => {
                        onChangeAccess(user.id, "delete", checked === true);
                    }}
                />
                <div className="flex items-center gap-1">
                    <PermissionTooltipAction
                        id={MODULE_IDS.Project}
                        action="write"
                        triggerClassName="inline-flex"
                    >
                        {({ isDenied }) => (
                            <Button
                                type="button"
                                variant="link"
                                className="size-7 p-0 text-foreground"
                                aria-label="Toggle execute, write and delete access"
                                title="Toggle execute, write and delete access"
                                disabled={isDenied || !canUpdateProjectAccess}
                                onClick={() => {
                                    onToggleAll(user.id);
                                }}
                            >
                                <CheckCheck className="size-4" />
                            </Button>
                        )}
                    </PermissionTooltipAction>
                    <PermissionTooltipAction
                        id={MODULE_IDS.Project}
                        action="write"
                        triggerClassName="inline-flex"
                    >
                        {({ isDenied }) => (
                            <Button
                                type="button"
                                variant="link"
                                className="size-7 p-0 text-destructive"
                                aria-label="Remove user access"
                                title="Remove user access"
                                disabled={isDenied || !canUpdateProjectAccess}
                                onClick={() => {
                                    onRemove(user.id);
                                }}
                            >
                                <Trash2 className="size-4" />
                            </Button>
                        )}
                    </PermissionTooltipAction>
                </div>
            </div>
        </div>
    );
}

export function ProjectUserAccessesDialog() {
    const { state, props: dialogOptions, close: closeDialog, clear: clearDialog } = useProjectUserAccessesDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const [updateVer, setUpdateVer] = useState(0);
    const [globalSearchQuery, setGlobalSearchQuery] = useState("");
    const [selectedGlobalUser, setSelectedGlobalUser] = useState<UserAccessOption | null>(null);
    const [envSearchQueries, setEnvSearchQueries] = useState<Record<string, string>>({});
    const [selectedEnvUsers, setSelectedEnvUsers] = useState<Record<string, UserAccessOption | null>>({});
    const [envAccesses, setEnvAccesses] = useState<EnvAccessesState>({});
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    const open = state.mode !== "closed";
    const projectId = state.mode === "open" ? state.projectId : "";
    const stateEnvs = state.mode === "open" ? state.envs : null;
    const projectEnvs = useMemo(() => stateEnvs ?? [], [stateEnvs]);

    const accessQuery = ProjectUserAccessesQueries.useFindOne(
        { projectID: projectId },
        {
            enabled: open,
        },
    );

    const ownerAccess = accessQuery.data?.data.ownerAccess;
    const canUpdateProjectAccess =
        canWrite && accessQuery.data?.data.currentUserActions.canUpdateProjectUserAccesses === true;
    const canViewModuleAccess = accessQuery.data?.data.currentUserActions.canViewModuleUserAccesses === true;

    const activeSearchQuery = useMemo(() => {
        if (globalSearchQuery) {
            return globalSearchQuery;
        }

        return Object.values(envSearchQueries).find(query => query.length > 0) ?? "";
    }, [envSearchQueries, globalSearchQuery]);

    const { data: usersData, isFetching: isFetchingUsers } = UsersPublicQueries.useFindManyBase(
        {
            search: activeSearchQuery,
            role: EUserRole.Member,
        },
        {
            enabled: open && canUpdateProjectAccess,
        },
    );

    const { mutate: updateProjectUserAccesses, isPending: isUpdating } = ProjectUserAccessesCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project user accesses updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            setUpdateVer(0);
            setGlobalSearchQuery("");
            setSelectedGlobalUser(null);
            setEnvSearchQueries({});
            setSelectedEnvUsers({});
            setEnvAccesses({});
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    useEffect(() => {
        if (!open || !accessQuery.data || hasChanges) {
            return;
        }

        setEnvAccesses(buildInitialEnvAccesses(projectEnvs, accessQuery.data.data.envUserAccesses));
        setUpdateVer(accessQuery.data.data.updateVer);
        setHasChanges(false);
        setSelectedGlobalUser(null);
        setSelectedEnvUsers({});
    }, [accessQuery.data, hasChanges, open, projectEnvs]);

    const globalUserOptions = useMemo<ComboboxOption<UserAccessOption>[]>(() => {
        const unavailableUserIds = new Set<string>();

        if (ownerAccess) {
            unavailableUserIds.add(ownerAccess.id);
        }

        for (const user of usersData?.data ?? []) {
            const isInEveryEnv =
                projectEnvs.length > 0 &&
                projectEnvs.every(env => (envAccesses[env.name] ?? []).some(accessUser => accessUser.id === user.id));

            if (isInEveryEnv) {
                unavailableUserIds.add(user.id);
            }
        }

        return (usersData?.data ?? [])
            .filter(user => !unavailableUserIds.has(user.id))
            .map(user => ({
                value: {
                    id: user.id,
                    username: user.username,
                    email: user.email,
                    fullName: user.fullName,
                    photo: user.photo,
                    role: user.role,
                },
                label: getUserDisplayName(user),
            }));
    }, [envAccesses, ownerAccess, projectEnvs, usersData]);

    const envUserOptionsMap = useMemo(() => {
        const map: Record<string, ComboboxOption<UserAccessOption>[]> = {};

        for (const env of projectEnvs) {
            const unavailableUserIds = new Set((envAccesses[env.name] ?? []).map(user => user.id));

            if (ownerAccess) {
                unavailableUserIds.add(ownerAccess.id);
            }

            map[env.name] = (usersData?.data ?? [])
                .filter(user => !unavailableUserIds.has(user.id))
                .map(user => ({
                    value: {
                        id: user.id,
                        username: user.username,
                        email: user.email,
                        fullName: user.fullName,
                        photo: user.photo,
                        role: user.role,
                    },
                    label: getUserDisplayName(user),
                }));
        }

        return map;
    }, [envAccesses, ownerAccess, projectEnvs, usersData]);

    function handleAddToAllEnvs() {
        if (!selectedGlobalUser || !canUpdateProjectAccess || state.mode !== "open") {
            return;
        }

        setEnvAccesses(current => {
            const next = { ...current };

            for (const env of state.envs) {
                const existing = next[env.name] ?? [];

                if (existing.some(user => user.id === selectedGlobalUser.id)) {
                    continue;
                }

                next[env.name] = [...existing, createUserFromOption(selectedGlobalUser)];
            }

            return next;
        });
        setSelectedGlobalUser(null);
        setGlobalSearchQuery("");
        setHasChanges(true);
    }

    const handleEnvSearchUser = useCallback((envName: string, query: string) => {
        setEnvSearchQueries(current => ({
            ...current,
            [envName]: query,
        }));
    }, []);

    const handleEnvSelectUser = useCallback((envName: string, option: UserAccessOption | null) => {
        setSelectedEnvUsers(current => ({
            ...current,
            [envName]: option,
        }));
    }, []);

    function handleAddToEnv(envName: string) {
        const selectedUser = selectedEnvUsers[envName];

        if (!selectedUser || !canUpdateProjectAccess) {
            return;
        }

        setEnvAccesses(current => {
            const existing = current[envName] ?? [];

            if (existing.some(user => user.id === selectedUser.id)) {
                return current;
            }

            return {
                ...current,
                [envName]: [...existing, createUserFromOption(selectedUser)],
            };
        });
        setSelectedEnvUsers(current => ({
            ...current,
            [envName]: null,
        }));
        setEnvSearchQueries(current => ({
            ...current,
            [envName]: "",
        }));
        setHasChanges(true);
    }

    function handleToggleAll(envName: string, userId: string) {
        if (!canUpdateProjectAccess) {
            return;
        }

        setEnvAccesses(current =>
            Object.fromEntries(
                Object.entries(current).map(([name, users]) => {
                    if (name !== envName) {
                        return [name, users];
                    }

                    return [
                        name,
                        users.map(user => {
                            if (user.id !== userId) {
                                return user;
                            }

                            const shouldCheck = !(user.access.execute && user.access.write && user.access.delete);

                            return {
                                ...user,
                                access: {
                                    ...user.access,
                                    execute: shouldCheck,
                                    write: shouldCheck,
                                    delete: shouldCheck,
                                },
                            };
                        }),
                    ];
                }),
            ),
        );
        setHasChanges(true);
    }

    function handleChangeAccess(
        envName: string,
        userId: string,
        key: "execute" | "write" | "delete",
        checked: boolean,
    ) {
        if (!canUpdateProjectAccess) {
            return;
        }

        setEnvAccesses(current => ({
            ...current,
            [envName]: (current[envName] ?? []).map(user =>
                user.id === userId
                    ? {
                          ...user,
                          access: {
                              ...user.access,
                              [key]: checked,
                          },
                      }
                    : user,
            ),
        }));
        setHasChanges(true);
    }

    function handleRemoveUser(envName: string, userId: string) {
        if (!canUpdateProjectAccess) {
            return;
        }

        setEnvAccesses(current => ({
            ...current,
            [envName]: (current[envName] ?? []).filter(user => user.id !== userId),
        }));
        setHasChanges(true);
    }

    function handleSubmit() {
        if (state.mode === "closed" || !canUpdateProjectAccess) {
            return;
        }

        updateProjectUserAccesses({
            projectID: state.projectId,
            updateVer,
            envUserAccesses: state.envs.map(env => ({
                name: env.name,
                userAccesses: (envAccesses[env.name] ?? []).map(user => ({
                    id: user.id,
                    access: user.access,
                })),
            })),
        });
    }

    function handleClose() {
        if (isUpdating) {
            return;
        }

        if (canWrite && hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) {
            return;
        }

        closeDialog();
        dialogOptions?.onClose?.();
    }

    const moduleAccesses = accessQuery.data?.data.moduleUserAccesses ?? [];
    const projectName = state.mode === "open" ? state.projectName : "";
    const showAccessContent = !(accessQuery.isFetching && !accessQuery.data) && !accessQuery.error;

    return (
        <Dialog
            open={open}
            onOpenChange={nextOpen => {
                if (!nextOpen) {
                    handleClose();
                }
            }}
        >
            <DialogFixedContent className="min-w-[390px] w-[1000px]">
                <DialogHeader>
                    <DialogTitle>User accesses on project {projectName}</DialogTitle>
                </DialogHeader>
                <div className="px-6">
                    <Separator className="opacity-50" />
                </div>

                <DialogBody>
                    {accessQuery.isFetching && !accessQuery.data ? (
                        <AppLoader />
                    ) : accessQuery.error ? (
                        <div className="flex min-h-32 flex-col items-center justify-center gap-3 text-sm text-muted-foreground">
                            <span>Unable to load project user accesses.</span>
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => {
                                    void accessQuery.refetch();
                                }}
                            >
                                Retry
                            </Button>
                        </div>
                    ) : (
                        <div className="flex flex-col gap-6">
                            {ownerAccess && (
                                <>
                                    <InfoBlock
                                        title={
                                            <LabelWithInfo
                                                label="Project Owner"
                                                content="Project owner has full access to this project."
                                            />
                                        }
                                        titleWidth={180}
                                    >
                                        <div className="flex flex-wrap items-center gap-4 py-3">
                                            <div className="min-w-[220px] flex-1">
                                                <UserInfo user={ownerAccess} />
                                            </div>
                                            <div className="flex items-center gap-4">
                                                <AccessCheckbox
                                                    id={`owner-${ownerAccess.id}-read`}
                                                    checked={ownerAccess.access.read}
                                                    disabled
                                                    label="Read"
                                                />
                                                <AccessCheckbox
                                                    id={`owner-${ownerAccess.id}-execute`}
                                                    checked={ownerAccess.access.execute}
                                                    disabled
                                                    label="Execute"
                                                />
                                                <AccessCheckbox
                                                    id={`owner-${ownerAccess.id}-write`}
                                                    checked={ownerAccess.access.write}
                                                    disabled
                                                    label="Write"
                                                />
                                                <AccessCheckbox
                                                    id={`owner-${ownerAccess.id}-delete`}
                                                    checked={ownerAccess.access.delete}
                                                    disabled
                                                    label="Delete"
                                                />
                                                <div className="w-[60px]" />
                                            </div>
                                        </div>
                                    </InfoBlock>
                                    <Separator className="opacity-50" />
                                </>
                            )}

                            <InfoBlock
                                title={
                                    <LabelWithInfo
                                        label="Project Access"
                                        content="Add a user to all environments that do not already have them."
                                    />
                                }
                                titleWidth={180}
                            >
                                <div className="flex flex-wrap items-center gap-2">
                                    <Combobox
                                        options={globalUserOptions}
                                        value={selectedGlobalUser?.id ?? null}
                                        onChange={(_value, option) => {
                                            setSelectedGlobalUser(option);
                                        }}
                                        onSearch={setGlobalSearchQuery}
                                        placeholder="Select User(s)"
                                        searchable
                                        emptyText="No users available"
                                        className="flex-1 md:max-w-[420px]"
                                        valueKey="id"
                                        loading={isFetchingUsers}
                                        disabled={!canUpdateProjectAccess || projectEnvs.length === 0}
                                    />
                                    <PermissionTooltipAction
                                        id={MODULE_IDS.Project}
                                        action="write"
                                    >
                                        {({ isDenied }) => (
                                            <Button
                                                type="button"
                                                variant="outline"
                                                disabled={
                                                    isDenied ||
                                                    !selectedGlobalUser ||
                                                    !canUpdateProjectAccess ||
                                                    projectEnvs.length === 0
                                                }
                                                onClick={handleAddToAllEnvs}
                                            >
                                                <Plus className="size-4" />
                                                Add to All Envs
                                            </Button>
                                        )}
                                    </PermissionTooltipAction>
                                </div>
                            </InfoBlock>

                            {projectEnvs.map(env => (
                                <EnvAccessSection
                                    key={env.name}
                                    env={env}
                                    users={envAccesses[env.name] ?? []}
                                    selectedUser={selectedEnvUsers[env.name] ?? null}
                                    userOptions={envUserOptionsMap[env.name] ?? []}
                                    isFetchingUsers={isFetchingUsers}
                                    canUpdateProjectAccess={canUpdateProjectAccess}
                                    onSelectUser={handleEnvSelectUser}
                                    onSearchUser={handleEnvSearchUser}
                                    onAdd={handleAddToEnv}
                                    onToggleAll={handleToggleAll}
                                    onChangeAccess={handleChangeAccess}
                                    onRemoveUser={handleRemoveUser}
                                />
                            ))}

                            {canViewModuleAccess && (
                                <>
                                    <Separator className="opacity-50" />
                                    <InfoBlock
                                        title={
                                            <LabelWithInfo
                                                label="Module Access"
                                                content="Module access description"
                                            />
                                        }
                                        titleWidth={180}
                                    >
                                        {moduleAccesses.length > 0 && (
                                            <div className="divide-y">
                                                {moduleAccesses.map(user => (
                                                    <div
                                                        key={user.id}
                                                        className="flex flex-wrap items-center gap-4 py-3"
                                                    >
                                                        <div className="min-w-[220px] flex-1">
                                                            <UserInfo user={user} />
                                                        </div>
                                                        <div className="flex items-center gap-4">
                                                            <AccessCheckbox
                                                                id={`module-${user.id}-read`}
                                                                checked={user.access.read}
                                                                disabled
                                                                label="Read"
                                                            />
                                                            <AccessCheckbox
                                                                id={`module-${user.id}-execute`}
                                                                checked={user.access.execute}
                                                                disabled
                                                                label="Execute"
                                                            />
                                                            <AccessCheckbox
                                                                id={`module-${user.id}-write`}
                                                                checked={user.access.write}
                                                                disabled
                                                                label="Write"
                                                            />
                                                            <AccessCheckbox
                                                                id={`module-${user.id}-delete`}
                                                                checked={user.access.delete}
                                                                disabled
                                                                label="Delete"
                                                            />
                                                            <a
                                                                href={ROUTE.userManagement.users.single.$route(user.id)}
                                                                target="_blank"
                                                                rel="noreferrer"
                                                                className="text-sm font-medium text-primary hover:underline"
                                                            >
                                                                Settings
                                                            </a>
                                                        </div>
                                                    </div>
                                                ))}
                                            </div>
                                        )}
                                    </InfoBlock>
                                </>
                            )}
                        </div>
                    )}
                </DialogBody>
                {showAccessContent ? (
                    <DialogActionFooter className="flex justify-end">
                        <PermissionTooltipAction
                            id={MODULE_IDS.Project}
                            action="write"
                        >
                            {({ isDenied }) => (
                                <Button
                                    type="button"
                                    className="min-w-[100px]"
                                    disabled={isDenied || !canUpdateProjectAccess}
                                    isLoading={isUpdating}
                                    onClick={handleSubmit}
                                >
                                    Save
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    </DialogActionFooter>
                ) : null}
            </DialogFixedContent>
        </Dialog>
    );
}

interface UserInfoProps {
    user: Pick<ProjectUserAccessBase, "email" | "fullName" | "photo" | "username">;
}

interface AccessCheckboxProps {
    id: string;
    checked: boolean;
    disabled: boolean;
    label: string;
    onCheckedChange?: (checked: boolean | "indeterminate") => void;
}

interface EnvAccessSectionProps {
    env: ProjectEnvEntity;
    users: ProjectUserAccessBase[];
    selectedUser: UserAccessOption | null;
    userOptions: ComboboxOption<UserAccessOption>[];
    isFetchingUsers: boolean;
    canUpdateProjectAccess: boolean;
    onSelectUser: (envName: string, option: UserAccessOption | null) => void;
    onSearchUser: (envName: string, query: string) => void;
    onAdd: (envName: string) => void;
    onToggleAll: (envName: string, userId: string) => void;
    onChangeAccess: (envName: string, userId: string, key: "execute" | "write" | "delete", checked: boolean) => void;
    onRemoveUser: (envName: string, userId: string) => void;
}

interface EnvUserAccessRowProps {
    envName: string;
    user: ProjectUserAccessBase;
    canUpdateProjectAccess: boolean;
    onToggleAll: (userId: string) => void;
    onChangeAccess: (userId: string, key: "execute" | "write" | "delete", checked: boolean) => void;
    onRemove: (userId: string) => void;
}

import { useEffect, useMemo, useState } from "react";

import { Avatar, Button, Checkbox } from "@components/ui";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { CheckCheck, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import type { ProjectUserAccessBase } from "~/projects/api/services";
import { ProjectUserAccessesCommands } from "~/projects/data/commands";
import { ProjectUserAccessesQueries } from "~/projects/data/queries";

import { AppLoader, Combobox, type ComboboxOption, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { UsersPublicQueries } from "@application/shared/data-public/queries";
import { EUserRole } from "@application/shared/enums";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { useProjectUserAccessesDialogState } from "../hooks";

type UserAccessOption = Record<string, unknown> & {
    id: string;
    username: string;
    email: string;
    fullName: string;
    photo: string | null;
    role: EUserRole;
};

function getUserDisplayName(user: Pick<ProjectUserAccessBase, "email" | "fullName" | "username">) {
    return user.fullName || user.email || user.username;
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

export function ProjectUserAccessesDialog() {
    const { state, props: dialogOptions, close: closeDialog, clear: clearDialog } = useProjectUserAccessesDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const [searchQuery, setSearchQuery] = useState("");
    const [selectedUser, setSelectedUser] = useState<UserAccessOption | null>(null);
    const [projectAccesses, setProjectAccesses] = useState<ProjectUserAccessBase[]>([]);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    const open = state.mode !== "closed";
    const projectId = state.mode === "open" ? state.projectId : "";

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

    const { data: usersData, isFetching: isFetchingUsers } = UsersPublicQueries.useFindManyBase(
        {
            search: searchQuery,
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
            setSearchQuery("");
            setSelectedUser(null);
            setProjectAccesses([]);
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    useEffect(() => {
        if (!open || !accessQuery.data || hasChanges) {
            return;
        }

        setProjectAccesses(accessQuery.data.data.userAccesses);
        setHasChanges(false);
        setSelectedUser(null);
    }, [accessQuery.data, hasChanges, open]);

    const userOptions = useMemo<ComboboxOption<UserAccessOption>[]>(() => {
        const unavailableUserIds = new Set(projectAccesses.map(user => user.id));

        if (ownerAccess) {
            unavailableUserIds.add(ownerAccess.id);
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
    }, [ownerAccess, projectAccesses, usersData]);

    function handleAddUser() {
        if (!selectedUser || !canUpdateProjectAccess) {
            return;
        }

        setProjectAccesses(current => {
            if (current.some(user => user.id === selectedUser.id)) {
                return current;
            }

            return [
                ...current,
                {
                    id: selectedUser.id,
                    username: selectedUser.username,
                    email: selectedUser.email,
                    fullName: selectedUser.fullName,
                    photo: selectedUser.photo,
                    role: selectedUser.role,
                    access: {
                        read: true,
                        execute: false,
                        write: false,
                        delete: false,
                    },
                },
            ];
        });
        setSelectedUser(null);
        setSearchQuery("");
        setHasChanges(true);
    }

    function handleToggleAll(userId: string) {
        if (!canUpdateProjectAccess) {
            return;
        }

        setProjectAccesses(current =>
            current.map(user => {
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
        );
        setHasChanges(true);
    }

    function handleChangeAccess(userId: string, key: "execute" | "write" | "delete", checked: boolean) {
        if (!canUpdateProjectAccess) {
            return;
        }

        setProjectAccesses(current =>
            current.map(user =>
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
        );
        setHasChanges(true);
    }

    function handleRemoveUser(userId: string) {
        if (!canUpdateProjectAccess) {
            return;
        }

        setProjectAccesses(current => current.filter(user => user.id !== userId));
        setHasChanges(true);
    }

    function handleSubmit() {
        if (state.mode === "closed" || !canUpdateProjectAccess) {
            return;
        }

        updateProjectUserAccesses({
            projectID: state.projectId,
            userAccesses: projectAccesses.map(user => ({
                id: user.id,
                access: user.access,
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

    return (
        <Dialog
            open={open}
            onOpenChange={nextOpen => {
                if (!nextOpen) {
                    handleClose();
                }
            }}
        >
            <DialogContent className="min-w-[390px] w-[860px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>User accesses on project {projectName}</DialogTitle>
                </DialogHeader>

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
                                <hr className="border-border" />
                            </>
                        )}
                        <InfoBlock
                            title={
                                <LabelWithInfo
                                    label="Project Access"
                                    content="Project access description"
                                />
                            }
                            titleWidth={180}
                        >
                            <div className="flex flex-col gap-4">
                                <div className="flex flex-wrap items-center gap-2">
                                    <Combobox
                                        options={userOptions}
                                        value={selectedUser?.id ?? null}
                                        onChange={(_value, option) => {
                                            setSelectedUser(option);
                                        }}
                                        onSearch={setSearchQuery}
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
                                                onClick={handleAddUser}
                                            >
                                                <Plus className="size-4" />
                                                Add
                                            </Button>
                                        )}
                                    </PermissionTooltipAction>
                                </div>

                                {projectAccesses.length > 0 && (
                                    <div className="divide-y">
                                        {projectAccesses.map(user => (
                                            <div
                                                key={user.id}
                                                className="flex flex-wrap items-center gap-4 py-3"
                                            >
                                                <div className="min-w-[220px] flex-1">
                                                    <UserInfo user={user} />
                                                </div>
                                                <div className="flex items-center gap-4">
                                                    <AccessCheckbox
                                                        id={`project-${user.id}-read`}
                                                        checked={user.access.read}
                                                        disabled
                                                        label="Read"
                                                    />
                                                    <AccessCheckbox
                                                        id={`project-${user.id}-execute`}
                                                        checked={user.access.execute}
                                                        disabled={!canUpdateProjectAccess}
                                                        label="Execute"
                                                        onCheckedChange={checked => {
                                                            handleChangeAccess(user.id, "execute", checked === true);
                                                        }}
                                                    />
                                                    <AccessCheckbox
                                                        id={`project-${user.id}-write`}
                                                        checked={user.access.write}
                                                        disabled={!canUpdateProjectAccess}
                                                        label="Write"
                                                        onCheckedChange={checked => {
                                                            handleChangeAccess(user.id, "write", checked === true);
                                                        }}
                                                    />
                                                    <AccessCheckbox
                                                        id={`project-${user.id}-delete`}
                                                        checked={user.access.delete}
                                                        disabled={!canUpdateProjectAccess}
                                                        label="Delete"
                                                        onCheckedChange={checked => {
                                                            handleChangeAccess(user.id, "delete", checked === true);
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
                                                                        handleToggleAll(user.id);
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
                                                                        handleRemoveUser(user.id);
                                                                    }}
                                                                >
                                                                    <Trash2 className="size-4" />
                                                                </Button>
                                                            )}
                                                        </PermissionTooltipAction>
                                                    </div>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </div>
                        </InfoBlock>

                        {canViewModuleAccess && (
                            <>
                                <hr className="border-border" />
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

                        <div className="flex justify-end border-t pt-4">
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
                        </div>
                    </div>
                )}
            </DialogContent>
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

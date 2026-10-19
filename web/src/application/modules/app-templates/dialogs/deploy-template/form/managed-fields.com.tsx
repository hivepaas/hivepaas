import { AlertTriangle, DatabaseBackup } from "lucide-react";
import { type Control, Controller } from "react-hook-form";
import { managedEngineOf } from "~/projects/module-shared/utils";

import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
    Select,
    SelectContent,
    SelectGroup,
    SelectItem,
    SelectLabel,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";

import type { AppTemplateManagedDefaults } from "../../../api";

import type { DeployTemplateFormState } from "./deploy-template.form.types";
import { type ManagedArchiveOption, type ManagedFormState, NO_DUMP, NO_KEY } from "./managed-fields.state";

interface ManagedFieldsProps {
    control: Control<DeployTemplateFormState>;
    /** The template's engine, which the copy names what it works with after. */
    engine: string;
    archive: string;
    archives: ManagedArchiveOption[];
    keyAuths: { id: string; name: string }[];
    defaults?: AppTemplateManagedDefaults | null;
    readOnly: boolean;
}

/**
 * Where a managed database archives and how long its engine keeps what. The archive and its key are chosen here
 * for good: the prefix the archive is under never moves.
 */
export function ManagedFields({
    control,
    engine: engineName,
    archive,
    archives,
    keyAuths,
    defaults,
    readOnly,
}: ManagedFieldsProps) {
    const engine = managedEngineOf(engineName);
    const chosen = archives.find(option => `${option.kind}:${option.id}` === archive);
    const repos = archives.filter(option => option.kind === "repo");
    const storages = archives.filter(option => option.kind === "storage");

    return (
        <div className="space-y-4 rounded-xl border border-violet-500/30 bg-violet-500/5 p-4 shadow-2xs">
            <div className="flex items-center gap-2 border-b border-border/40 pb-2.5">
                <DatabaseBackup className="size-4 text-violet-500" />
                <h4 className="text-sm font-semibold text-foreground tracking-tight">Archive and backups</h4>
            </div>
            <p className="text-xs text-muted-foreground leading-relaxed">
                Every change is archived as it happens, so a failure loses about the archive timeout&apos;s worth of
                writes. The data volume is pinned to a node, where the engine&apos;s tools are put, or is on storage
                every node reaches: the database may then start on any node, and archives to object storage.
            </p>

            <Controller
                name="managed.archive"
                control={control}
                rules={{ required: "Choose where the database archives" }}
                render={({ field, fieldState }) => (
                    <Field data-invalid={fieldState.invalid}>
                        <FieldLabel className="text-xs font-medium text-foreground">
                            Archive to <span className="text-destructive">*</span>
                        </FieldLabel>
                        <Select
                            value={field.value}
                            onValueChange={field.onChange}
                            disabled={readOnly}
                        >
                            <SelectTrigger
                                className="h-9 text-xs"
                                aria-label="Archive to"
                            >
                                <SelectValue placeholder="A backup repository or a cloud storage" />
                            </SelectTrigger>
                            <SelectContent>
                                {repos.length > 0 && (
                                    <SelectGroup>
                                        <SelectLabel className="text-[11px]">Backup repositories</SelectLabel>
                                        {repos.map(option => (
                                            <SelectItem
                                                key={`repo:${option.id}`}
                                                value={`repo:${option.id}`}
                                                className="text-xs"
                                            >
                                                {option.name}
                                                {option.onVolume ? " (volume)" : " (S3)"}
                                            </SelectItem>
                                        ))}
                                    </SelectGroup>
                                )}
                                {storages.length > 0 && (
                                    <SelectGroup>
                                        <SelectLabel className="text-[11px]">Cloud storages</SelectLabel>
                                        {storages.map(option => (
                                            <SelectItem
                                                key={`storage:${option.id}`}
                                                value={`storage:${option.id}`}
                                                className="text-xs"
                                            >
                                                {option.name}
                                            </SelectItem>
                                        ))}
                                    </SelectGroup>
                                )}
                                {archives.length === 0 && (
                                    <div className="px-2 py-1.5 text-xs text-muted-foreground">
                                        No backup repository or cloud storage in this project yet.
                                    </div>
                                )}
                            </SelectContent>
                        </Select>
                        {fieldState.error && (
                            <FieldError className="text-[11px] text-destructive">{fieldState.error.message}</FieldError>
                        )}
                    </Field>
                )}
            />

            {chosen?.onVolume && (
                <div
                    role="note"
                    className="flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-xs text-amber-800 dark:text-amber-300"
                >
                    <AlertTriangle className="mt-0.5 size-3.5 shrink-0" />
                    <span>
                        An archive on a volume is only as safe as that volume is apart from the database: on the
                        database&apos;s own node, losing the node loses both. Object storage keeps it off-site.
                    </span>
                </div>
            )}

            {chosen && !chosen.onVolume && (
                <Controller
                    name="managed.keyAuth"
                    control={control}
                    render={({ field }) => (
                        <Field>
                            <FieldLabel className="text-xs font-medium text-foreground">Key</FieldLabel>
                            <Select
                                value={field.value === "" ? NO_KEY : field.value}
                                onValueChange={field.onChange}
                                disabled={readOnly}
                            >
                                <SelectTrigger
                                    className="h-9 text-xs"
                                    aria-label="Key"
                                >
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem
                                        value={NO_KEY}
                                        className="text-xs"
                                    >
                                        The storage&apos;s own
                                    </SelectItem>
                                    {keyAuths.map(key => (
                                        <SelectItem
                                            key={key.id}
                                            value={key.id}
                                            className="text-xs"
                                        >
                                            {key.name}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            <FieldDescription className="text-[11px]">
                                A repository&apos;s key can also delete its snapshots. A key that writes only under this
                                archive&apos;s prefix keeps a compromised database from reaching them.
                            </FieldDescription>
                        </Field>
                    )}
                />
            )}

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <ManagedTextField
                    control={control}
                    name="managed.baseBackupSchedule"
                    label="Base backup schedule"
                    placeholder={orDefault(defaults?.baseBackupSchedule, "0 2 * * *")}
                    description="A cron expression: when the nightly base backup and retention run."
                    readOnly={readOnly}
                />
                <ManagedTextField
                    control={control}
                    name="managed.retainFull"
                    label="Full backups kept"
                    placeholder={defaults?.retainFull ? String(defaults.retainFull) : "7"}
                    description="How far back a restore can go, in base backups: about a day each."
                    readOnly={readOnly}
                    pattern={/^\d{0,3}$/}
                    patternMessage="A whole number, 1 to 365"
                />
                <ManagedTextField
                    control={control}
                    name="managed.restoreTestSchedule"
                    label="Restore test schedule"
                    placeholder={orDefault(defaults?.restoreTestSchedule, "0 4 * * 0")}
                    description="When the archive is restored, to prove it restores."
                    readOnly={readOnly}
                />
                <ManagedTextField
                    control={control}
                    name="managed.archiveTimeout"
                    label="Archive timeout"
                    placeholder={orDefault(defaults?.archiveTimeout, "60s")}
                    description={
                        `The most a failure can lose: a quiet database's ${engine.archiveNoun} ` +
                        `${engine.archiveNounPlural ? "are" : "is"} archived at least this often.`
                    }
                    readOnly={readOnly}
                    pattern={/^(\d+(s|m|h))*$/}
                    patternMessage="A duration such as 60s or 5m"
                />
            </div>

            <ManagedTextField
                control={control}
                name="managed.restoreTestQuery"
                label="Restore test check"
                placeholder={engine.checkQueryExample}
                description="A query the restored copy must answer, besides listing its databases. Optional."
                readOnly={readOnly}
            />

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Controller
                    name="managed.dumpRepo"
                    control={control}
                    render={({ field }) => (
                        <Field>
                            <FieldLabel className="text-xs font-medium text-foreground">Dump into</FieldLabel>
                            <Select
                                value={field.value === "" ? DUMP_DEFAULT : field.value}
                                onValueChange={value => {
                                    field.onChange(value === DUMP_DEFAULT ? "" : value);
                                }}
                                disabled={readOnly}
                            >
                                <SelectTrigger
                                    className="h-9 text-xs"
                                    aria-label="Dump into"
                                >
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem
                                        value={DUMP_DEFAULT}
                                        className="text-xs"
                                    >
                                        {chosen?.kind === "repo"
                                            ? `The archive's repository (${chosen.name})`
                                            : "No dumps"}
                                    </SelectItem>
                                    {chosen?.kind === "repo" && (
                                        <SelectItem
                                            value={NO_DUMP}
                                            className="text-xs"
                                        >
                                            No dumps
                                        </SelectItem>
                                    )}
                                    {repos
                                        .filter(option => option.id !== chosen?.id)
                                        .map(option => (
                                            <SelectItem
                                                key={option.id}
                                                value={option.id}
                                                className="text-xs"
                                            >
                                                {option.name}
                                                {option.onVolume ? " (volume)" : " (S3)"}
                                            </SelectItem>
                                        ))}
                                </SelectContent>
                            </Select>
                            <FieldDescription className="text-[11px]">
                                A {engine.dumpTool} into a backup repository, beside the archive: a copy any{" "}
                                {engine.name} can load, restored from the app&apos;s backup snapshots.
                            </FieldDescription>
                        </Field>
                    )}
                />
                <ManagedTextField
                    control={control}
                    name="managed.dumpSchedule"
                    label="Dump schedule"
                    placeholder="0 3 * * *"
                    description="A cron expression: when the dump runs, at any frequency - 0 */6 * * * is every 6 hours."
                    readOnly={readOnly}
                />
            </div>
        </div>
    );
}

/** The dump select's value for what the request leaves to the server: the archive's repository, if it is one. */
const DUMP_DEFAULT = "__dump-default__";

interface ManagedTextFieldProps {
    control: Control<DeployTemplateFormState>;
    name: `managed.${Exclude<keyof ManagedFormState, "archive" | "keyAuth" | "dumpRepo">}`;
    label: string;
    placeholder: string;
    description: string;
    readOnly: boolean;
    pattern?: RegExp;
    patternMessage?: string;
}

function ManagedTextField({
    control,
    name,
    label,
    placeholder,
    description,
    readOnly,
    pattern,
    patternMessage,
}: ManagedTextFieldProps) {
    return (
        <Controller
            name={name}
            control={control}
            rules={pattern ? { pattern: { value: pattern, message: patternMessage ?? "Not valid" } } : undefined}
            render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                    <FieldLabel className="text-xs font-medium text-foreground">{label}</FieldLabel>
                    <Input
                        {...field}
                        aria-label={label}
                        placeholder={placeholder}
                        disabled={readOnly}
                        className="h-9 text-xs"
                    />
                    <FieldDescription className="text-[11px]">{description}</FieldDescription>
                    {fieldState.error && (
                        <FieldError className="text-[11px] text-destructive">{fieldState.error.message}</FieldError>
                    )}
                </Field>
            )}
        />
    );
}

/** A template's default, or the engine's when the template leaves it empty. */
function orDefault(value: string | undefined, fallback: string): string {
    return value !== undefined && value !== "" ? value : fallback;
}

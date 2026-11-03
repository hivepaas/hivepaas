import { type ReactNode, useRef } from "react";

import { Button, Checkbox, FieldError, Input, Label } from "@components/ui";
import { Textarea } from "@components/ui/textarea";
import { FileUpIcon, FolderOpenIcon, XIcon } from "lucide-react";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

const COMPOSE_PLACEHOLDER = `services:
  web:
    image: nginx:1.27
    ports: ["8080:80"]`;

/**
 * The compose file and what it is read with: its .env, the project's name and
 * env - or, for an existing project, the target given - the profiles whose
 * services are created.
 */
export function ComposeSource({
    compose,
    onComposeChange,
    dotEnv,
    onDotEnvChange,
    projectName,
    onProjectNameChange,
    fileName,
    envName,
    onEnvNameChange,
    profiles,
    selectedProfiles,
    onProfilesChange,
    error,
    nameError,
    envError,
    target,
    folder,
    onOpenFolder,
    onCloseFolder,
    onUseEnvExample,
    isReading,
}: Props) {
    const fileInputRef = useRef<HTMLInputElement>(null);
    const folderInputRef = useRef<HTMLInputElement>(null);

    return (
        <div className="flex w-full flex-col gap-6">
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Compose File"
                        content="A docker-compose.yml: paste it, open the file, or open its folder - the files it reads are then taken from there, as the review asks for them. It is read on the server without touching anything there - no file of the server's, nor its environment."
                    />
                }
            >
                <div className="flex w-full max-w-[900px] flex-col gap-2">
                    <Textarea
                        value={compose}
                        onChange={event => {
                            onComposeChange(event.target.value);
                        }}
                        placeholder={COMPOSE_PLACEHOLDER}
                        minRows={14}
                        maxRows={32}
                        spellCheck={false}
                        className="font-mono text-xs"
                    />
                    <div className="flex items-center gap-3">
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => fileInputRef.current?.click()}
                        >
                            <FileUpIcon className="size-4" />
                            Open file
                        </Button>
                        <input
                            ref={fileInputRef}
                            type="file"
                            accept=".yml,.yaml,text/yaml"
                            className="hidden"
                            onChange={event => {
                                const file = event.target.files?.[0];
                                if (fileInputRef.current) {
                                    fileInputRef.current.value = "";
                                }
                                if (file) {
                                    void file.text().then(onComposeChange);
                                }
                            }}
                        />
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => folderInputRef.current?.click()}
                        >
                            <FolderOpenIcon className="size-4" />
                            Open folder
                        </Button>
                        <input
                            ref={element => {
                                folderInputRef.current = element;
                                // Not in React's attributes: the browser's own, for picking a directory.
                                element?.setAttribute("webkitdirectory", "");
                            }}
                            type="file"
                            multiple
                            className="hidden"
                            onChange={event => {
                                const files = event.target.files ? Array.from(event.target.files) : [];
                                if (folderInputRef.current) {
                                    folderInputRef.current.value = "";
                                }
                                if (files.length > 0) {
                                    onOpenFolder(files);
                                }
                            }}
                        />
                        {isReading && <span className="text-xs text-muted-foreground">Reading…</span>}
                    </div>
                    {folder && (
                        <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                            <FolderOpenIcon className="size-4" />
                            <span>
                                <code className="font-mono text-foreground">
                                    {[folder.name, folder.base, folder.composeName].filter(Boolean).join("/")}
                                </code>{" "}
                                and the {folder.fileCount} {folder.fileCount === 1 ? "file" : "files"} beside it: those
                                the file reads are given from the folder.
                            </span>
                            <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                onClick={onCloseFolder}
                            >
                                <XIcon className="size-4" />
                                Close folder
                            </Button>
                        </div>
                    )}
                    {error && <p className="max-w-[720px] text-xs text-destructive">{error}</p>}
                </div>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label=".env"
                        content="The variables the file uses, as its .env holds them. A required one with no value is asked for below. One that reads as a secret is kept as an env secret, encrypted."
                    />
                }
            >
                <div className="flex w-full max-w-[900px] flex-col gap-2">
                    <Textarea
                        value={dotEnv}
                        onChange={event => {
                            onDotEnvChange(event.target.value);
                        }}
                        placeholder="DB_PASSWORD=..."
                        minRows={3}
                        maxRows={12}
                        spellCheck={false}
                        className="font-mono text-xs"
                    />
                    {folder?.envExampleName && dotEnv.trim() === "" && (
                        <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                            The folder has no .env, but has {folder.envExampleName}.
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                onClick={onUseEnvExample}
                            >
                                Use {folder.envExampleName}
                            </Button>
                            <span>Change the passwords it gives: they are the example&apos;s, known to anybody.</span>
                        </div>
                    )}
                </div>
            </InfoBlock>

            {target ?? (
                <>
                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Project Name"
                                content="Empty takes the file's own name, `name:`."
                            />
                        }
                    >
                        <Input
                            value={projectName}
                            onChange={event => {
                                onProjectNameChange(event.target.value);
                            }}
                            placeholder={fileName || "my-project"}
                            aria-invalid={nameError !== undefined}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                        />
                        <FieldError errors={[nameError === undefined ? undefined : { message: nameError }]} />
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={220}
                        title={
                            <LabelWithInfo
                                label="Environment"
                                content="The project's one env, where every service becomes an app."
                            />
                        }
                    >
                        <Input
                            value={envName}
                            onChange={event => {
                                onEnvNameChange(event.target.value);
                            }}
                            placeholder="production"
                            aria-invalid={envError !== undefined}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                        />
                        <FieldError errors={[envError === undefined ? undefined : { message: envError }]} />
                    </InfoBlock>
                </>
            )}

            {profiles.length > 0 && (
                <InfoBlock
                    titleWidth={220}
                    title={
                        <LabelWithInfo
                            label="Profiles"
                            content="The services of the profiles chosen are created, beside those in none - as `docker compose --profile` does."
                        />
                    }
                >
                    <div className="flex flex-wrap gap-4">
                        {profiles.map(profile => (
                            <div
                                key={profile}
                                className="flex items-center gap-2"
                            >
                                <Checkbox
                                    id={`compose-profile-${profile}`}
                                    checked={selectedProfiles.includes(profile)}
                                    onCheckedChange={value => {
                                        onProfilesChange(
                                            value === true
                                                ? [...selectedProfiles, profile]
                                                : selectedProfiles.filter(item => item !== profile),
                                        );
                                    }}
                                />
                                <Label
                                    htmlFor={`compose-profile-${profile}`}
                                    className="font-mono text-xs font-normal"
                                >
                                    {profile}
                                </Label>
                            </div>
                        ))}
                    </div>
                </InfoBlock>
            )}
        </div>
    );
}

interface Props {
    compose: string;
    onComposeChange: (value: string) => void;
    dotEnv: string;
    onDotEnvChange: (value: string) => void;
    projectName: string;
    onProjectNameChange: (value: string) => void;
    /** The file's own name for the project, offered when none is typed. */
    fileName: string;
    envName: string;
    onEnvNameChange: (value: string) => void;
    profiles: string[];
    selectedProfiles: string[];
    onProfilesChange: (value: string[]) => void;
    /** What the file itself was refused for. */
    error?: string;
    /** What the project's name and env were refused for, shown under each. */
    nameError?: string;
    envError?: string;
    /** Where the services go in an existing project, instead of a new project's name and env. */
    target?: ReactNode;
    /** The folder opened, if one is. */
    folder?: ComposeFolderSummary;
    onOpenFolder: (files: File[]) => void;
    onCloseFolder: () => void;
    onUseEnvExample: () => void;
    isReading: boolean;
}

/** What the page shows of a folder opened. */
export interface ComposeFolderSummary {
    name: string;
    /** The compose file's directory in the folder; empty for the folder itself. */
    base: string;
    composeName: string;
    fileCount: number;
    /** The example .env beside the compose file, if any. */
    envExampleName?: string;
}

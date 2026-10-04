import { useRef } from "react";

import { Button, Checkbox, Input, Label } from "@components/ui";
import { Textarea } from "@components/ui/textarea";
import { FileUpIcon } from "lucide-react";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

const COMPOSE_PLACEHOLDER = `services:
  web:
    image: nginx:1.27
    ports: ["8080:80"]`;

/**
 * The compose file and what it is read with: its .env, the project's name and
 * env, the profiles whose services are created.
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
    isReading,
}: Props) {
    const fileInputRef = useRef<HTMLInputElement>(null);

    return (
        <div className="flex w-full flex-col gap-6">
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Compose File"
                        content="A docker-compose.yml: paste it, or open the file. It is read on the server without touching anything there - no file of the server's, nor its environment."
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
                        {isReading && <span className="text-xs text-muted-foreground">Reading…</span>}
                    </div>
                    {error && <p className="max-w-[720px] text-xs text-destructive">{error.message}</p>}
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
                <Textarea
                    value={dotEnv}
                    onChange={event => {
                        onDotEnvChange(event.target.value);
                    }}
                    placeholder="DB_PASSWORD=..."
                    minRows={3}
                    maxRows={12}
                    spellCheck={false}
                    className="max-w-[900px] font-mono text-xs"
                />
            </InfoBlock>

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
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                />
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
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                />
            </InfoBlock>

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
    error?: Error;
    isReading: boolean;
}

import { Button, Checkbox, Input } from "@components/ui";
import { Badge } from "@components/ui/badge";
import { DicesIcon } from "lucide-react";
import type { ComposeVariableInput, ComposeVariableView } from "~/operations/domain";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

/** 32 hex characters from the browser's random source, shown where they are typed. */
function randomSecret(): string {
    const bytes = new Uint8Array(16);
    window.crypto.getRandomValues(bytes);

    return Array.from(bytes, byte => byte.toString(16).padStart(2, "0")).join("");
}

/**
 * The variables the file uses: a value over the .env's, whether it is kept as
 * a secret, and one generated for a secret left empty.
 */
export function ComposeVariables({ variables, inputs, onChange }: Props) {
    if (variables.length === 0) {
        return null;
    }

    const update = (name: string, input: ComposeVariableInput) => {
        onChange({ ...inputs, [name]: { ...inputs[name], ...input } });
    };

    return (
        <InfoBlock
            titleWidth={220}
            title={
                <LabelWithInfo
                    label="Variables"
                    content="What the file's ${VARIABLES} are given, over the .env. A secret one is kept as an env secret, encrypted, which the apps' variables refer to; the containers get the same value."
                />
            }
        >
            <div className="flex w-full max-w-[900px] flex-col divide-y rounded-md border">
                {variables.map(variable => {
                    const input = inputs[variable.name] ?? {};
                    const secret = input.secret ?? variable.secret;
                    const missing = variable.required && !variable.given && !input.value;

                    return (
                        <div
                            key={variable.name}
                            className="flex flex-wrap items-center gap-3 px-3 py-2"
                        >
                            <div className="flex w-[220px] items-center gap-2">
                                <code className="truncate font-mono text-xs text-foreground">{variable.name}</code>
                                {variable.required && (
                                    <Badge
                                        variant="outline"
                                        className={missing ? "border-destructive/40 text-destructive" : undefined}
                                    >
                                        required
                                    </Badge>
                                )}
                            </div>
                            <Input
                                value={input.value ?? ""}
                                type={secret ? "password" : "text"}
                                autoComplete="off"
                                onChange={event => {
                                    update(variable.name, {
                                        value: event.target.value === "" ? undefined : event.target.value,
                                    });
                                }}
                                placeholder={
                                    variable.given ? "from the .env" : variable.default || (missing ? "a value" : "")
                                }
                                className="min-w-[200px] flex-1 font-mono text-xs"
                            />
                            <div className="flex items-center gap-2">
                                <Checkbox
                                    id={`compose-variable-secret-${variable.name}`}
                                    checked={secret}
                                    onCheckedChange={value => {
                                        update(variable.name, { secret: value === true });
                                    }}
                                />
                                <label
                                    htmlFor={`compose-variable-secret-${variable.name}`}
                                    className="text-xs text-muted-foreground"
                                >
                                    Secret
                                </label>
                            </div>
                            {secret && (
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="sm"
                                    title="Generate a random value"
                                    onClick={() => {
                                        update(variable.name, { value: randomSecret() });
                                    }}
                                >
                                    <DicesIcon className="size-4" />
                                    Generate
                                </Button>
                            )}
                        </div>
                    );
                })}
            </div>
        </InfoBlock>
    );
}

interface Props {
    variables: ComposeVariableView[];
    inputs: Record<string, ComposeVariableInput>;
    onChange: (inputs: Record<string, ComposeVariableInput>) => void;
}

/**
 * Helper functions for parsing and stringifying environment variables
 * between array format and text format (key=value format)
 */

export type EnvVarRecord = {
    key: string;
    value: string;
    isLiteral: boolean;
};

export const MULTILINE_ENV_SEPARATOR = "-----MULTILINE-ENV-----";

function normalizeLineEndings(text: string): string {
    return text.replace(/\r\n/g, "\n").replace(/\r/g, "\n");
}

function isSeparatorLine(line: string): boolean {
    return line.trim() === MULTILINE_ENV_SEPARATOR;
}

function parseSingleLineEnvVar(line: string): EnvVarRecord | null {
    const trimmedLine = line.trim();

    if (trimmedLine === "" || trimmedLine.startsWith("#") || isSeparatorLine(line)) {
        return null;
    }

    const equalsIndex = line.indexOf("=");

    if (equalsIndex === -1) {
        return {
            key: trimmedLine,
            value: "",
            isLiteral: false,
        };
    }

    const key = line.substring(0, equalsIndex).trim();

    if (key === "" || key === MULTILINE_ENV_SEPARATOR) {
        return null;
    }

    return {
        key,
        value: line.substring(equalsIndex + 1),
        isLiteral: false,
    };
}

function parseMultilineEnvVar(lines: string[]): EnvVarRecord | null {
    if (lines.length === 0) {
        return null;
    }

    const [firstLine = "", ...restLines] = lines;
    const equalsIndex = firstLine.indexOf("=");
    const key = equalsIndex === -1 ? firstLine.trim() : firstLine.substring(0, equalsIndex).trim();

    if (key === "" || key === MULTILINE_ENV_SEPARATOR || key.startsWith("#")) {
        return null;
    }

    const valueLines = equalsIndex === -1 ? restLines : [firstLine.substring(equalsIndex + 1), ...restLines];

    return {
        key,
        value: valueLines.join("\n"),
        isLiteral: false,
    };
}

/**
 * Convert array of env var records to text format (key=value format)
 * @param records Array of env var records
 * @returns String in format: key1=value1\nkey2=value2
 */
export function stringifyEnvVars(records: EnvVarRecord[]): string {
    if (records.length === 0) {
        return "";
    }

    return records
        .filter(record => {
            // Skip records with empty keys
            return record.key.trim() !== "";
        })
        .map(record => {
            const key = record.key.trim();
            const { value } = record;
            const serializedRecord = `${key}=${value}`;

            if (value.includes("\n")) {
                return [MULTILINE_ENV_SEPARATOR, serializedRecord, MULTILINE_ENV_SEPARATOR].join("\n");
            }

            return serializedRecord;
        })
        .join("\n");
}

/**
 * Parse text format (key=value format) to array of env var records
 * @param text Text in format: key1=value1\nkey2=value2
 * @returns Array of env var records
 */
export function parseEnvVars(text: string): EnvVarRecord[] {
    if (!text || text.trim() === "") {
        return [];
    }

    const lines = normalizeLineEndings(text).split("\n");
    const records: EnvVarRecord[] = [];
    let normalLines: string[] = [];

    const flushNormalLines = () => {
        for (const line of normalLines) {
            const record = parseSingleLineEnvVar(line);

            if (record) {
                records.push(record);
            }
        }

        normalLines = [];
    };

    for (let index = 0; index < lines.length; index++) {
        const line = lines[index] ?? "";

        if (!isSeparatorLine(line)) {
            normalLines.push(line);
            continue;
        }

        flushNormalLines();

        const multilineLines: string[] = [];
        index++;

        while (index < lines.length) {
            const multilineLine = lines[index];

            if (multilineLine === undefined || isSeparatorLine(multilineLine)) {
                break;
            }

            multilineLines.push(multilineLine);
            index++;
        }

        const record = parseMultilineEnvVar(multilineLines);

        if (record) {
            records.push(record);
        }
    }

    flushNormalLines();

    return records;
}

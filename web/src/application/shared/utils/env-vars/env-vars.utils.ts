/**
 * Helper functions for parsing and stringifying environment variables
 * between array format and text format (key=value format)
 */

export type EnvVarRecord = {
    key: string;
    value: string;
    isLiteral: boolean;
};

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
            const value = record.value.trim();
            // Format: key=value (value can contain =, newlines, etc.)
            return `${key}=${value}`;
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

    const lines = text.split("\n");
    const records: EnvVarRecord[] = [];

    for (const line of lines) {
        const trimmedLine = line.trim();

        // Skip blank lines
        if (trimmedLine === "") {
            continue;
        }

        // Skip comment lines (optional - lines starting with #)
        if (trimmedLine.startsWith("#")) {
            continue;
        }

        // Find the first = sign to split key and value
        const equalsIndex = trimmedLine.indexOf("=");

        if (equalsIndex === -1) {
            // No = sign found, treat entire line as key with empty value
            if (trimmedLine !== "") {
                records.push({
                    key: trimmedLine,
                    value: "",
                    isLiteral: false,
                });
            }
        } else {
            // Split at first = sign
            const key = trimmedLine.substring(0, equalsIndex).trim();
            const value = trimmedLine.substring(equalsIndex + 1);

            // Skip if key is empty
            if (key === "") {
                continue;
            }

            records.push({
                key,
                value,
                isLiteral: false, // Default to false, can be updated later if needed
            });
        }
    }

    return records;
}

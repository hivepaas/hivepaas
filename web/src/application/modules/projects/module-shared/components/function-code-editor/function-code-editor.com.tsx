import { useId } from "react";

import Prism from "prismjs";
import "prismjs/components/prism-go";
import "prismjs/components/prism-json";
import "prismjs/components/prism-python";
import "prismjs/components/prism-typescript";
import "prismjs/themes/prism-tomorrow.css";
import Editor from "react-simple-code-editor";

import { MONO_FONT_FAMILY } from "@application/shared/utils";

import { type CodeLanguage, languageOfPath } from "../../utils";

function highlight(code: string, language: CodeLanguage): string {
    const grammar = language === "plain" ? undefined : Prism.languages[language];

    if (!grammar) {
        return code.replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;");
    }

    return Prism.highlight(code, grammar, language);
}

/**
 * An editor for one file of a function's code, highlighted after its extension:
 * JavaScript, TypeScript, Python, Go, JSON.
 */
export function FunctionCodeEditor({ path, value, onChange, readOnly = false }: Props) {
    const generatedId = useId();
    const language = languageOfPath(path);

    return (
        <div className="min-h-0 flex-1 overflow-auto rounded-md bg-[#1e1e1e]">
            <Editor
                value={value}
                onValueChange={onChange}
                highlight={code => highlight(code, language)}
                padding={16}
                textareaId={`function-code-${generatedId}`}
                readOnly={readOnly}
                style={{
                    minHeight: "100%",
                    fontFamily: MONO_FONT_FAMILY,
                    fontSize: 14,
                    lineHeight: 1.6,
                    backgroundColor: "#1e1e1e",
                    color: "#f8f8f2",
                    outline: "none",
                }}
            />
        </div>
    );
}

interface Props {
    path: string;
    value: string;
    onChange: (value: string) => void;
    readOnly?: boolean;
}

import { useCallback } from "react";

import { useTranslation } from "react-i18next";

import { type Language } from "@i18n/types";

function createHook() {
    return function useI18n() {
        const { t, i18n } = useTranslation("translation");

        const changeLanguage = useCallback(
            (lang: Language) => {
                localStorage.setItem("lang", lang);

                return i18n.changeLanguage(lang);
            },
            [i18n],
        );

        return {
            language: i18n.language as Language,
            changeLanguage,
            t,
        };
    };
}

export const useI18n = createHook();

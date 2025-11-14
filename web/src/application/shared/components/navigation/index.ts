import { AppLinkBasic, AppLinkModules } from "./app-link";
import { AppNavLinkBasic, AppNavLinkModules } from "./app-nav-link";
import { AppNavigateBasic, AppNavigateModules } from "./app-navigate";

export * from "./tab-navigation";

export const AppLink = {
    Basic: AppLinkBasic,
    Modules: AppLinkModules,
};

export const AppNavLink = {
    Basic: AppNavLinkBasic,
    Modules: AppNavLinkModules,
};

export const AppNavigate = {
    Basic: AppNavigateBasic,
    Modules: AppNavigateModules,
};

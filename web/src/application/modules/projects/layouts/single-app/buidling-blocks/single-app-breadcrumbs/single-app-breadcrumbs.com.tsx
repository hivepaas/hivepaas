import { Fragment, memo } from "react";

import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from "@components/ui";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

function View({ app, appRoute, project, items = [] }: Props) {
    const hasExtraItems = items.length > 0;
    const appLabel = app.name === "" ? "<unset>" : app.name;

    return (
        <Breadcrumb>
            <BreadcrumbList>
                <BreadcrumbItem>
                    <AppLink.Modules to={ROUTE.projects.list.$route}>Projects</AppLink.Modules>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                    <AppLink.Modules to={ROUTE.projects.single.apps.$route(project.id)}>{project.name}</AppLink.Modules>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                    {hasExtraItems && appRoute ? (
                        <AppLink.Modules to={appRoute}>{appLabel}</AppLink.Modules>
                    ) : (
                        <BreadcrumbPage>{appLabel}</BreadcrumbPage>
                    )}
                </BreadcrumbItem>
                {items.map(item => (
                    <Fragment key={`${item.label}-${item.to ?? "current"}`}>
                        <BreadcrumbSeparator />
                        <BreadcrumbItem>
                            {item.to ? (
                                <AppLink.Modules to={item.to}>{item.label}</AppLink.Modules>
                            ) : (
                                <BreadcrumbPage>{item.label}</BreadcrumbPage>
                            )}
                        </BreadcrumbItem>
                    </Fragment>
                ))}
            </BreadcrumbList>
        </Breadcrumb>
    );
}

interface SingleAppBreadcrumbItem {
    label: string;
    to?: string;
}

interface Props {
    app: {
        name: string;
    };
    appRoute?: string;
    project: {
        id: string;
        name: string;
    };
    items?: SingleAppBreadcrumbItem[];
}

export const SingleAppBreadcrumbs = memo(View);

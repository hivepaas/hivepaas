import { memo } from "react";

import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from "@components/ui";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

function View({ node }: Props) {
    return (
        <Breadcrumb>
            <BreadcrumbList>
                <BreadcrumbItem>
                    <AppLink.Modules to={ROUTE.cluster.nodes.$route}>Nodes</AppLink.Modules>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                    <BreadcrumbPage>{node.name}</BreadcrumbPage>
                </BreadcrumbItem>
            </BreadcrumbList>
        </Breadcrumb>
    );
}

interface Props {
    node: {
        name: string;
    };
}

export const SingleNodeBreadcrumbs = memo(View);

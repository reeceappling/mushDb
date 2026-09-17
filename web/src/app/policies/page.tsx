import React, {Suspense} from "react";
import PageWrapper from "@/app/components/clientGeneric";
import AuthArea from "@/app/components/authClient";
import {Metadata} from "next";
import {mushDbTitle} from "@/app/components/Constants";
import {GetReaderWriterNames} from "@/app/components/serverActions";
import {CookiesProvider} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {MainViewArea} from "@/app/view/[itemType]/[idEncoded]/client";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";

// TODO: Cache Components adoption. Refactor this route so this opt-out can be removed.
// See: https://nextjs.org/docs/app/guides/migrating-to-cache-components
//export const instant = false;

export const metadata: Metadata = {
    title: `Privacy Policy`,
    description: "Privacy Policy",
};

export default async function Page() {
    let err: string | undefined = undefined
    let readers: string[] = []
    try {
        readers = await GetReaderWriterNames() // Done on the server
    } catch (e) {
        err = "Failed to load page wrapper component: "+JSON.stringify(e)
    }
    const body = <>
        <h1>{"Policies"}</h1>
        <ErrorDisplay err={err}/>
        <ul>
            <li><a href={"/policies/cookies"}>{"Cookies Policy"}</a></li>
            <li><a href={"/policies/privacy"}>{"Privacy Policy"}</a></li>
            <li><a href={"/policies/security"}>{"Security Policy"}</a></li>{/* TODO: ensure list updated! make dynamic!*/}
        </ul>
    </>
    return <PageWrapper props={{pageType: "policies", readers: readers}}>
        <Suspense fallback={body}>
            {body}
        </Suspense>
    </PageWrapper>
}
import {Metadata} from "next";
import {GetReaderWriterNames} from "@/app/components/serverActions";
import PageWrapper from "@/app/components/clientGeneric";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import React from "react";
import "@/app/ui/policies.css"
import {BaseExternalUrl} from "@/app/components/Constants";

// TODO: Cache Components adoption. Refactor this route so this opt-out can be removed.
// See: https://nextjs.org/docs/app/guides/migrating-to-cache-components
//export const instant = false;

type Props = {
    params: Promise<{ policy: string }>
};

export async function generateMetadata(
    {params}: Props,
): Promise<Metadata> {
    // read route params
    const {policy} = await params

    // fetch data
    let text = policy + " policy"
    const capitalized = CapitalizeFirstLetter(text)
    return {
        title: capitalized,
        description: capitalized,
        alternates: {
            canonical: `${BaseExternalUrl}/policies/${policy}`,
            languages: {
                "en-US": `${BaseExternalUrl}/policies/${policy}`,
            }
        },
        robots: {
            index: true,
            follow: true,
            nocache: false,
        },
    }
}

function CapitalizeFirstLetter(s: string) {
    return s.charAt(0).toUpperCase() + s.slice(1);
}

export default async function Page({
                                       params,
                                   }:
                                   Props
) {
    const {policy} = await params
    let err: string | undefined = undefined
    let readers: string[] = []
    try {
        readers = await GetReaderWriterNames() // Done on the server
    } catch (e) {
        err = "Failed to load page wrapper component: " + JSON.stringify(e)
    }
    return <PageWrapper props={{pageType: "policies", readers: readers}}>
        <div>
            <ErrorDisplay err={err}/>
            <h1 className={"centerH"}>{CapitalizeFirstLetter(policy + " Policy")}</h1>
            <PolicyTextDisplay policy={policy}/>
        </div>
    </PageWrapper>
}

export async function PolicyTextDisplay({policy}: { policy: string }) {
    //const url = `${await BaseInternalUrl()}/staticContent/policies/${policy}.txt`
    const url = `http://web:3000/staticContent/policies/${policy}.txt`
    try {
        const resp = await fetch(url, {
                method: 'GET',
                headers: {
                    credentials: 'include', // TODO: maybe not
                    'Accept': 'text/html',
                },
                next: {revalidate:3600}, // Cache for an hour (3600)
            }
        )
        const paragraphs = (await resp.text()).split('\n\n');
        return <div className={"policyParagraphs"}>
                {paragraphs.map((paragraph, index) => {
                    // Only render the paragraph if it's not an empty line
                    return <p key={index} className={"policyParagraph"}>
                        {paragraph.trim()}
                    </p>
                })}
            </div>
    } catch(e){
        return <ErrorDisplay err={"failed to get policy text: "+JSON.stringify(e)}/>
    }
}
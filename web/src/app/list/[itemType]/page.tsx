import PageWrapper from "@/app/components/clientGeneric";
import React from "react";
import {GetReaderWriterNames} from "@/app/components/serverActions";
import {cookies} from "next/headers";
import {BaseExternalUrl} from "@/app/components/Constants";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import ListDisplay from "@/app/list/[itemType]/client";
import {CookiesProvider} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {Metadata} from "next";

// TODO: Cache Components adoption. Refactor this route so this opt-out can be removed.
// See: https://nextjs.org/docs/app/guides/migrating-to-cache-components
//export const instant = false;

type Props = {
    params: Promise<{ itemType: string }>
};

// Next.js runs this first to set the tab title
export async function generateMetadata({params}: Props): Promise<Metadata> {
    const {itemType} = await params
    return {
        title: itemType + ` list`,
        description: "List page for all " + itemType + "s" // TODO: msss, speciess, and subspeciess are incorrect here. maybe more
    };
}

export default async function Page({
                                       params,
                                   }: Props,
) {
    const itemType = (await params).itemType
    const cookieStore = await cookies()
    const session = cookieStore.get('_gothic_session')
    const allCookies = cookieStore.getAll().map(cookie => `${cookie.name}=${cookie.value}`).join('; ');

    const getData: (a1: string) => Promise<any> = async (itemTypeA: string) => {
        return new Promise<any>((accept, reject) => {
            fetch(BaseExternalUrl + "/db/list/" + itemTypeA, {
                method: 'Get',
                credentials: 'include',
                headers: {
                    'Accept': 'application/json',
                    'Cookie': allCookies, // TODO: necessary?
                },
            }).then((res) => {
                if (!res.ok) {
                    return res.text().then(txt => {
                        reject("response not ok: " + txt)
                    }).catch(err => {
                        reject("response not ok and failed to decode: " + JSON.stringify(err))
                    })
                }
                res.json().then(data => {
                    accept(data)
                }).catch(err1 => {
                    throw (err1)
                })
            }).catch(err2 => {
                reject(JSON.stringify(err2))
            })
        })
    }
    try {
        const readers = await GetReaderWriterNames() // Done on the server
        const data = await getData(itemType)

        return <PageWrapper props={{pageType: "list", readers: readers}}>
            <CookiesProvider cookies={cookieStore.getAll()} session={session?.value}>
                {/*<div className={"fullPage"}>*/}
                    <ListDisplay itemType={itemType} inpData={data}/>
                {/*</div>*/}
            </CookiesProvider>
        </PageWrapper>
    } catch (e) {
        return <PageWrapper props={{pageType: "error", readers: []}}>
            <ErrorDisplay err={"Error loading data: " + String(e)}/>
        </PageWrapper>
    }
}

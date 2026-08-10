import React from "react";
import {BaseExternalUrl, mushDbTitle} from "@/app/components/Constants";
import {GetReaderWriterNames} from "@/app/components/serverActions";
import PageWrapper from "@/app/components/clientGeneric";
import {cookies} from 'next/headers'
import {MainViewArea} from "@/app/view/[itemType]/[idEncoded]/client";
import {CookiesProvider} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {Metadata} from "next";

type Props = {
    params: Promise<{
        itemType: string
        idEncoded: string // urlEncoded
    }>
};
// Next.js runs this first to set the tab title
export async function generateMetadata({ params }: Props): Promise<Metadata> { // TODO: add generateMetadata on all pages!
    const {itemType, idEncoded} = await params
    const decoded = decodeURI(idEncoded)
    const desc = 'Viewpage for '+itemType+` `+decoded
    if (itemType == "species" || itemType == "subspecies") {
        return {
            title: {
                absolute: decoded,
            },
            description: desc,
        };
    }
    return {
        title: itemType+` `+decoded,
        description: desc, // TODO: fix for things like MSS
    };
}

export default async function Page({
                                       params,
                                   }: {
    params: Promise<{
        itemType: string
        idEncoded: string // urlEncoded
    }>,
}) {
    const {itemType, idEncoded} = await params
    const cookieStore = await cookies()
    const session = cookieStore.get('_gothic_session')
    const allCookies = cookieStore.getAll().map(cookie => `${cookie.name}=${cookie.value}`).join('; ');


    const getData: (a1: string, a2: string) => Promise<any> = async (itemTypeA: string, idEnc: string) => {
        return new Promise<React.JSX.Element>((accept, reject) => { // TODO: REIMPLEMENT!
            fetch(BaseExternalUrl + "/db/get/" + itemTypeA + "/" + idEnc, {
                method: 'Get',
                credentials: 'include',
                headers: {
                    'Accept': 'application/json',
                    //'Access-Control-Allow-Origin': BaseExternalUrl || "*", // TODO: ENSURE OK! maybe "*"?
                    'Cookie': allCookies, // REQUIRED // TODO: can we drop this because we have included creds?
                    // TODO: set Origin header to web? or should this be BaseExternalUrl?
                },
            }).then((res) => {
                console.log("got response " + JSON.stringify(res))
                if (!res.ok) {
                    return res.text().then(txt => {
                        throw new Error("response not ok: " + txt + ". Status " + res.status)
                    }).catch(err => {
                        throw new Error("response not ok and failed to decode: " + JSON.stringify(err) + ". Status " + res.status)
                    })
                }
                res.json().then((data) => {
                    console.log(data)
                    accept(data)
                }).catch(err1 => {
                    console.log("failed to resolve json data from result, " + JSON.stringify(err1))
                    reject(err1)
                })
            }).catch(err1 => {
                reject(err1)
            })
        })
    }
    try {
        const data = await getData(itemType, idEncoded)
        const readers = await GetReaderWriterNames()
        return <PageWrapper props={{pageType: "view", readers: readers}}>
            <CookiesProvider cookies={cookieStore.getAll()} session={session?.value}> {/* TODO: validate working*/}
                <div className={"fullPage"}>
                    <MainViewArea itemType={itemType} inpData={data}/>
                </div>
            </CookiesProvider>
        </PageWrapper>
    } catch (e) {
        return <PageWrapper props={{pageType: "view", readers: []}}>
            <div className={"fullPage"}>
                <div>{"Page not loaded. Nonexistent or unauthorized entry: "}</div> {/* TODO: STYLING*/}
                <div>{JSON.stringify(e)/* TODO: CHANGE!*/}</div>
            </div>
        </PageWrapper>
    }

}


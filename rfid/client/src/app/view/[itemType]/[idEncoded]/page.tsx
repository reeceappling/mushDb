import {BaseExternalUrl} from "@/app/components/Constants";
import React from "react";
import {GetReaderWriterNames} from "@/app/view/[itemType]/[idEncoded]/serverActions";
import PageWrapper from "@/app/components/clientGeneric";
import {cookies} from 'next/headers'
import {MainViewArea} from "@/app/view/[itemType]/[idEncoded]/client";

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
                    // TODO: ensure all session headers exist here
                    'Accept': 'application/json',
                    'Cookie': allCookies, // TODO: ensure this is everywhere
                },
            }).then((res) => {
                if (!res.ok) {
                    return res.text().then(txt => {
                        throw new Error("response not ok: " + txt)
                    }).catch(err => {
                        throw new Error("response not ok and failed to decode: "+JSON.stringify(err));
                    })
                }
                console.log("got response")
                res.json().then((data) => {
                    console.log(data)
                    accept(data)
                }).catch(err1 => {
                    console.log(data)
                    reject(err1)
                })
            }).catch(err1 => {
                reject(err1)
            })
        })
    }
    const data = await getData(itemType, idEncoded)
    const readers = await GetReaderWriterNames()
    return <PageWrapper props={{pageType: "view", readers: readers}}>
        <div className={"fullPage"}>
            <MainViewArea itemType={itemType} inpData={data} allCookies={allCookies}/>
        </div>
    </PageWrapper>
}


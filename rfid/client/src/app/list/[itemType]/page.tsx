import PageWrapper from "@/app/components/clientGeneric";
import React from "react";
import {GetReaderWriterNames} from "@/app/components/serverActions";
import {cookies} from "next/headers";
import {BaseExternalUrl, TopPageHeaderLevel} from "@/app/components/Constants";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import ListDisplay from "@/app/list/[itemType]/client";
import {SessionProvider} from "@/app/components/formSubcomponents/sessionContext/session";

export default async function Page({
                                       params,
                                   }: {
    params: Promise<{ itemType: string }>,
}) {
    const itemType = (await params).itemType
    const cookieStore = await cookies()
    const session = cookieStore.get('_gothic_session')
    const allCookies = cookieStore.getAll().map(cookie => `${cookie.name}=${cookie.value}`).join('; ');

    const getData:(a1:string)=>Promise<any>= async (itemTypeA: string)=>{
        return new Promise<any>((accept, reject)=> {
            fetch(BaseExternalUrl + "/db/list/" + itemTypeA, {
                method: 'Get',
                credentials: 'include',
                headers: {
                    //'credentials': 'include', // TODO: ok?
                    'Accept': 'application/json',
                    'Cookie': allCookies,
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
                }).catch(err1 => {throw(err1)})
            }).catch(err2 => {
                reject(JSON.stringify(err2))
            })
        })
    }
    try {
        const readers = await GetReaderWriterNames() // TODO: DO THIS ON SERVER
        const data = await getData(itemType)

        return <PageWrapper props={{pageType:"list",readers:readers}}>
            <SessionProvider session={session?.value}>
            <div className={"fullPage"}>
                <ListDisplay itemType={itemType} inpData={data}/>
            </div>
            </SessionProvider>
        </PageWrapper>
    } catch (e) {
        return <div className={"fullPage"}>
            <ErrorDisplay err={"Error loading data: "+String(e)} headerLevel={TopPageHeaderLevel}/>
        </div>
    }
}

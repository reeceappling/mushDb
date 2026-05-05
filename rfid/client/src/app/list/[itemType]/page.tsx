import PageWrapper from "@/app/components/clientGeneric";
import React from "react";
import {GetReaderWriterNames} from "@/app/view/[itemType]/[idEncoded]/serverActions";
import {cookies} from "next/headers";
import {BaseExternalUrl, TopPageHeaderLevel} from "@/app/components/Constants";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import ListDisplay from "@/app/list/[itemType]/client";

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
        return new Promise<any>((accept, reject)=>{ // TODO: REIMPLEMENT!
            fetch(BaseExternalUrl + "/db/list/" + itemTypeA, {
                method: 'Get',
                credentials: 'include',
                headers: {
                    'Accept': 'application/json',
                    'Cookie': allCookies,
                },
            }).then((res) => {
                if(!res.ok){
                    return res.text().then(txt=>{
                        throw new Error("response not ok: "+txt);
                    }).catch(err=>{
                        throw new Error("response not ok and failed to decode: ")
                    })
                }
                console.log("got response")
                res.json().then((data) => {
                    console.log(data)
                    accept(data)
                }).catch(err1 => {
                    console.log(err1)
                    reject(err1)
                })
            }).catch(err1 => {
                reject(err1)
            })
        })
    }
    try {
        const data = await getData(itemType)
        const readers = await GetReaderWriterNames() // TODO: DO THIS ON SERVER
        return <PageWrapper props={{pageType:"list",readers:readers}}>
            <div className={"fullPage"}>
                <ListDisplay itemType={itemType} inpData={data}/>
            </div>
        </PageWrapper>
    } catch (e) {
        return <div className={"fullPage"}>
            <ErrorDisplay err={"Error loading data: "+String(e)} headerLevel={TopPageHeaderLevel}/>
        </div>
    }
}

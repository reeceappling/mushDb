// import {SelectorProps} from "@/app/components/selector";
// import {JSX, ReactElement, useState} from "react";
// import {InlineProps} from "@/app/components/common";
// import H from "@/app/components/formSubcomponents/utils/headers";
// import EntryLink from "@/app/components/formSubcomponents/entryLink";
// import {redirect} from "next/navigation";
// import {BaseUrl} from "@/app/components/Constants";


import DateArea from "@/app/components/formSubcomponents/date";
import {DisposedDisplay} from "@/app/components/formSubcomponents/commonClient";

export default function Centered({children}:{children:React.ReactNode}){
    return <div className={"centerH"}>{children}</div>
}
export function CapitalizeFirstLetter(str: string): string {
    if (!str || str==="") return str;
    return str.charAt(0).toUpperCase() + str.slice(1);
}

export function CreatedUpdatedDisposedArea(
    {created, updated, initialDisposed, readonly, setDisposedOnParent}: {
        created: number,
        updated: number,
        initialDisposed?: number,
        readonly: boolean,
        setDisposedOnParent?: (n?: number) => void,
    }
) {
    return <>
        <DateArea pre={"Created: "} when={created} readonly={true}/>
        <DateArea pre={"Updated: "} when={updated} readonly={true}/>
        <DisposedDisplay readonly={readonly} initial={initialDisposed} setDisposedOnParent={setDisposedOnParent}/>
    </>
}
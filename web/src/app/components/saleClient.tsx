'use client'

import React, { useContext, useState} from "react";
import {
    IsValidNote,
    NewEntryNotes,
    Note, NotesFormArea
} from "@/app/components/formSubcomponents/notes";
import {AllEntries} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest, DoUpdateRequest,
    FlexedArea,
    FlexedSinglesGroup,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    NewColumn,
    NewEntryFormWrapper,
    NumberToDateStr,
    OptionalArrayOfType, RequiredKey,
    viewUrlFor
} from "@/app/components/common";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {SaleData} from "@/app/components/saleServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {
    AclDisplay,
    MarshalAcl,
    TogglableAreaWithDepth,
    UnmarshalAcl
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";

// TODO: list page not working
// TODO: ensure display page doing what we want

export function AssertSale(input: any): asserts input is SaleData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Sale assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Sale assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Sale assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export default function SaleDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput<SaleData>) {
        const [initial, setInitial] = useState(data)
        
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL>(initial.acl)
        const updateInitial = (updated: SaleData)=>{
            setInitial(updated)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
            setErr(undefined)
        }
        // TODO: BE ABLE TO MODIFY SALES PERMS, USE WHOLE BODY
        const cookies = useContext(CookiesContext)
        const saleUpdateSubmit = () => {
            const body: any = {
                notes:notes,
                acl:MarshalAcl(acl),
            }
            DoUpdateRequest("sale",data._id, body, AssertSale, allCookies(cookies))
                .then(v=>{
                    updateInitial(new SaleData(v))
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                })
        }
        return (
            <DisplayFormWrapper entryType={"sale"}>
                <ErrorDisplay err={err}/>
                <ID props={{id:data._id, txt:"Sale", entryType:"sale"}}/>

                <FlexedArea>
                    <FlexedSinglesGroup>
                        <DateArea pre={"Sold on: "} when={initial.creationDate} readonly={true}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    </FlexedSinglesGroup>
                </FlexedArea>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay initial={initial.acl} readonly={readonly} updateParent={setAcl} />
                </TogglableAreaWithDepth>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    saleUpdateSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
}

export function NewSaleForm( // TODO: overhaul! Need Id available?
    {
        onCreate
    }: {
        onCreate?: (s: SaleData)=>void
    }) {
    // id/lot, saleDate, lastUpdated are done on server
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()

    const cookies = useContext(CookiesContext)
    const createSale = (e: React.MouseEvent) => {
        e.preventDefault()


        const body = {
            notes: notes,
        }
        DoCreateRequest("sale", body, AssertSale, allCookies(cookies))
            .then(v=>{
                onCreate ? onCreate(v) : window.location.assign(viewUrlFor("sale", v._id))/*redirect() TODO: ensure ok*/ // TODO: del if working
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    return <NewEntryFormWrapper entryType={"sale"}>
        <ErrorDisplay err={err}/>
        <NewEntryNotes setNotes={setNotes} />
        <button className={"greenButton"} onClick={createSale}>{"Create Sale"}</button>
    </NewEntryFormWrapper>

}

export function SaleArea(
    {readonly, sale, setSale, canCreateSale}:{
        readonly: boolean,
        sale?: string,
        setSale?: (s: string) => void,
        canCreateSale: boolean
    }
){
    return null  // TODO: REENABLE EVENTUALLY

    // const [open, setOpen] = useState(false)
    // const saleCreated = (newsale: string)=>{
    //     setSale && setSale(newsale)
    //     setOpen(false)
    // }
    // if (sale === undefined) {
    //     return <div>
    //         <TestAndValidate todos={["overhaul this for sales"]}>
    //         <div className={"saleArea"}>
    //             <div className={"inline"}>{"Available to sell"}</div>
    //             {/* TODO: ensure that mark as sold marks as disposed as well, except in cases where things can be multi-sold*/}
    //             {!readonly && <button className={"basicButtonSmall inline"} onClick={()=>{setOpen(!open)}}>{open?"Close sale creation area":"Mark as sold"}</button> /* TODO: FIX ME SO THIS CREATES A NEW SALE!*/}
    //         </div>
    //         {open &&
    //             <NewSaleForm onCreate={s=>saleCreated(s._id)} />
    //         }
    //         </TestAndValidate>
    //     </div>
    // }
    // const b58id = sale
    // return <div className={"areaWrapper"}>
    //     <div className={"areaHeader"}>{"Sold: "}</div>
    //     <div>
    //         <EntryLinkForId props={{displayId: b58id, linkId: b58id, entryType:"sale", openInNewTab:true}}/>
    //     </div>
    // </div>
}

export function SalesArea(
    {sales, allowCreate, headerLevel, readonly, offset, setEntries}:{
        sales?: string[],
        allowCreate: boolean
        headerLevel?: number,
        readonly?: boolean
        setEntries?:(ps:string[])=>void
        offset?:number
    }) {
    return null  // TODO: REENABLE EVENTUALLY
    // const [current, setCurrent] = useState(sales || [])
    // const [creatorOpen, setCreatorOpen] = useState(false)
    // const updateEntries=(items: string[])=>{
    //     setEntries && setEntries(items)
    //     setCurrent(items)
    // }
    // const addArea = ()=>{
    //     return <div>
    //         <div>
    //             <button className={"basicButton"} onClick={() => {
    //                 setCreatorOpen(!creatorOpen)
    //             }}>{creatorOpen ? "Close Sale Creator" : "Create a new Sale"}</button>
    //         </div>
    //         {creatorOpen && <NewSaleForm onCreate={(s)=>{
    //             updateEntries([...current,s._id])}
    //         }/> }
    //     </div>
    // }
    // return <div>
    //         <div>{"Associated Sales: "}</div>
    //         {(sales || []).map(s=>{
    //             return <div key={s}>
    //                 <EntryLinkForId props={{displayId:s,linkId:s,entryType:"sale",openInNewTab:true}}/>
    //             </div>
    //         })}
    //         {(!readonly&&allowCreate) && addArea()}
    //     </div>
}

export function SaleListPageTable({data, onClick, withLink}: ListPageItems<SaleData>) {
    let cols: ListTableColumn<SaleData>[] = [
        NewColumn("ID", (v)=>v._id, true),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }, true),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }), // TODO: fit?
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SaleData)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new SaleData(v)}}/>
}

export function SaleSelectorTable({data, onClick}: ListPageItems<SaleData>) {
    return <SaleListPageTable data={data} onClick={onClick} withLink={true} />
}
// // TODO: likely get rid of
// export function SaleSelector(
//     {
//         doSelect,
//     }: {
//         doSelect: (val: SaleData | undefined) => void,
//     }) {
//     const table = (items: SaleData[]):JSX.Element=>{
//         return <SaleSelectorTable data={items} onClick={doSelect}/>
//     }
//
//     return <ExistingRecentSelector entryType={"sale"} entryTypes={"sales"} doSelect={doSelect} asserter={AssertSale}
//                                    table={table}>
//     </ExistingRecentSelector>
// }
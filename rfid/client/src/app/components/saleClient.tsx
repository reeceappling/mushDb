'use client'

import React, {useState} from "react";
import NotesAreaOld, {IsValidNote, NewEntryNotes, Note, NoteEntriesGroup} from "@/app/components/formSubcomponents/notes";
import {AllEntries, Data} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea, {NumberToDate} from "@/app/components/formSubcomponents/date";
import {
    DisplayInput, HandleJsonResponse,
    HeaderLevel,
    InlineProps,
    OptionalArrayOfType,
    OptionalKey
} from "@/app/components/common";
import {redirect} from "next/navigation";
import {ErrorDisplay, OpenMainPage} from "@/app/components/formSubcomponents/commonClient";
import {SaleData} from "@/app/components/saleServer";
import {BaseExternalUrl} from "@/app/components/Constants";
import EntryLink from "@/app/components/formSubcomponents/entryLink";
import {useCookies} from "react-cookie";
import {dataFor} from "@/app/components/agarRecipeClient";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {DisplayFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {FlexedArea, FlexedSinglesGroup, NotesFormArea} from "@/app/components/agarBatchClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";

export function AssertSale(input: any): asserts input is SaleData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Sale assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
       ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Plate assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Sale assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export default function SaleDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput) {
    try {
        AssertSale(data)
        const [initial, setInitial] = useState(data)
        
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const updateInitial = (updated: SaleData)=>{
            setInitial(updated)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
        }
        ////const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
        // TODO: BE ABLE TO MODIFY SALES PERMS, USE WHOLE BODY
        //,
        //             perms: perms, // TODO: validate on insert
        //         }
        const saleUpdateSubmit = () => {
            fetch(BaseExternalUrl+"/db/update/sale/"+data._id, {
                method: "POST",
                headers: {
                    credentials: 'include',
                    // TODO: may need 'Cookie': cookies,
                    'Content-type': 'application/json'
                },
                body: JSON.stringify({notes:notes,acl:acl,}) // TODO: used to just be notes. Fix in go
            }).then(HandleJsonResponse)
                .then((entry)=>{
                    AssertSale(entry)
                    updateInitial(entry)
                }).catch((er)=>{
                    setErr(JSON.stringify(er))
                })
        }
        return (
            <DisplayFormWrapper entryType={"sale"}>
                <ErrorDisplay err={err} headerLevel={headerLevel} />
                <ID id={data._id} txt={"Sale"} entryType={"sale"}/>
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
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
                </TogglableAreaWithDepth>
                {readonly ? null : <input type="submit" value="Update" onClick={saleUpdateSubmit} onSubmit={(e)=>{e.preventDefault();}}/>}
                {/* TODO: ?<OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>*/}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Sale data format incorrect: " + err}</div>
    }
}

export function NewSaleForm(
    {
        headerLevel, onCreate
    }: {
        headerLevel?: number
        onCreate?: (s: SaleData)=>void
    }) {
    // id/lot, saleDate, lastUpdated are done on server
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    //const [perms, setPerms] = useState<EntryPerms | undefined>()
    ////const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
    const createSale = (e: React.MouseEvent) => {
        e.preventDefault()


        let body = {
            notes: notes,
            //perms: perms, // TODO: KEEP PERMS FROM PARENT?
        }
        fetch(BaseExternalUrl+"/create/sale", {
            method: "POST",
            headers: {
                credentials: 'include',
                // TODO: may need 'Cookie': cookies,
                'Content-type': 'application/json'
            },
            body: JSON.stringify(body)
        })
            .then(HandleJsonResponse)
            .then((sale) => {
                AssertSale(sale)
                onCreate?onCreate(sale):redirect(BaseExternalUrl+"/view/sale/"+sale._id)
            })
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
    }
    return <NewEntryFormWrapper entryType={"sale"}>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <NewEntryNotes setNotes={setNotes} />
        {/*<EntryPermsArea setEntryPerms={setPerms}/> /!* TODO: perms from parent??? *!/*/}
        <button className={"greenButton"} onClick={createSale}>{"Create Sale"}</button>
    </NewEntryFormWrapper>

}

export function SaleInline({data, expandByDefault, onClick, showMainPageButton, idIsLink}: InlineProps<SaleData>) {
    const notes = data.notes || []
    const [expanded, setExpanded] = useState(expandByDefault)
    const expansionArea = () => {
        if (expanded) {
            return null
        }
        return <div className={"fullWidth"}>
            <div>{"Notes: "}</div>
            {notes.map((n) => {
                return <div>
                    {NumberToDate(new Date(n.time)) + " - " + n.note}
                </div>
            })}
        </div>
    }
    const b58id = data._id
    return <div>
        <div>
            <ID id={b58id} txt={"Sale"} entryType={"sale"} allowOpenMainPage={showMainPageButton} linkPage={idIsLink}/>
            <DateArea pre={"Sold on: "} when={data.creationDate} readonly={true}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
            <button className={"basicButton"} onClick={() => {
                setExpanded(!expanded)
            }}>{expanded ? "See Less" : "See More"}</button>
        </div>
        {expansionArea()}
    </div>
}

export function SaleArea(
    {readonly, headerLevel, sale, setSale, canCreateSale}:{
        readonly: boolean,
        sale?: string,
        setSale?: (s: string) => void,
        headerLevel?: number,
        canCreateSale: boolean
    }
){
    // TODO: does saleArea need incremented depth? probably not
    const [open, setOpen] = useState(false)
    const saleCreated = (newsale: string)=>{
        setSale && setSale(newsale)
        setOpen(false)
    }
    if (sale === undefined) {
        return <div>
            <TestAndValidate todos={["fully test this"]}>
            <div className={"saleArea"}>
                <div className={"inline"}>{"Available to sell"}</div>
                {!readonly && <button className={"basicButton inline"/* TODO: ok? */} onClick={()=>{setOpen(!open)}}>{open?"Close sale creation area":"Mark as sold"}</button> /* TODO: FIX ME SO THIS CREATES A NEW SALE!*/}
            </div>
            {open &&
                <NewSaleForm headerLevel={headerLevel} onCreate={s=>saleCreated(s._id)} />
            }
            </TestAndValidate>
        </div>
    }
    const b58id = sale
    return <div className={"areaWrapper"}>
        <div className={"areaHeader"}>{"Sold: "}</div>
        <div>
            <EntryLink props={{displayedId: b58id, linkId: b58id, entryType:"sale", openInNewTab:true}}>{b58id}</EntryLink>
        </div>
    </div>
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
    // TODO: does salesArea need incremented depth? probably not
    const [current, setCurrent] = useState(sales || [])
    const [creatorOpen, setCreatorOpen] = useState(false)
    const updateEntries=(items: string[])=>{
        setEntries && setEntries(items)
        setCurrent(items)
    }
    const addArea = ()=>{
        return <div>
            <div>
                <button className={"basicButton"} onClick={() => {
                    setCreatorOpen(!creatorOpen)
                }}>{creatorOpen ? "Close Sale Creator" : "Create a new Sale"}</button>
            </div>
            {creatorOpen && <NewSaleForm headerLevel={headerLevel} onCreate={(s)=>{
                updateEntries([...current,s._id])}
            }/> }
        </div>
    }
    return <div>
            <div>{"Associated Sales: "}</div>
            {(sales || []).map(s=>{
                const b58id = s
                return <div>
                    <EntryLink props={{displayedId:b58id,linkId:b58id,entryType:"sale",openInNewTab:true}}>
                        <div>{b58id}</div>
                    </EntryLink>
                </div>
            })}
            {(!readonly&&allowCreate) && addArea()}
        </div>
}
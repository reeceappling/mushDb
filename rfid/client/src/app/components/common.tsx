'use client'

import {defaultHeaderLevel} from "@/app/components/formSubcomponents/utils/headers";
import {JSX, ReactNode, SetStateAction, SyntheticEvent, useState} from "react";
import {
    Contamination,
    ContaminationForm,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {NumberToDate} from "@/app/components/formSubcomponents/date";
import {SplitAllEntries} from "@/app/components/formSubcomponents/shared";
import {NewPicWithNotesForm, PicWithNotesForm} from "@/app/components/formSubcomponents/picWithNotes";
import {BaseExternalUrl} from "@/app/components/Constants";
import ReaderWriterSelector, {
    ReadTagFunc,
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {useRfidReaderContext} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import {
    AssertDualListResult,
    AssertSubstrateRecipe,
    validatorForAssertion
} from "@/app/components/substrateRecipeClient";
import TestAndValidate from "@/app/components/testing/untested";
import * as React from "react";
import {InputTextInlineTitle} from "@/app/components/formSubcomponents/numericInput";
import {AssertAgarRecipe} from "@/app/components/agarRecipeClient";
import {AssertAgarBatch} from "@/app/components/agarBatchClient";
import {AssertBag} from "@/app/components/bagClient";
import {AssertFruit} from "@/app/components/fruitClient";
import {AssertFruitingChamber} from "@/app/components/fruitingChamberClient";
import {AssertGrainBatch} from "@/app/components/grainBatchClient";
import {AssertJar} from "@/app/components/jarClient";
import {AssertJarRecipe} from "@/app/components/jarRecipeClient";
import {AssertLcRecipe} from "@/app/components/lcRecipeClient";
import {AssertLc} from "@/app/components/lcClient";
import {AssertLcSyringe} from "@/app/components/lcSyringeClient";
import {AssertMss} from "@/app/components/mssClient";
import {AssertPcRun} from "@/app/components/pcRunClient";
import {AssertPlate} from "@/app/components/plateClient";
import {AssertProject} from "@/app/components/projectClient";
import {AssertSale} from "@/app/components/saleClient";
import {AssertSlant} from "@/app/components/slantClient";
import {AssertSpecies} from "@/app/components/speciesClient";
import {AssertSporePrint} from "@/app/components/sporePrintClient";
import {AssertSporeSwab} from "@/app/components/sporeSwabClient";
import {AssertStasisTube} from "@/app/components/stasisTubeClient";
import {AssertSubspecies} from "@/app/components/subspeciesClient";
import {AssertSubstrateBatch} from "@/app/components/substrateBatchClient";
import { AssertUser } from "./userClient";
import {AssertWaterJar} from "@/app/components/waterJarClient";
import {AssertTransfer} from "@/app/components/transferClient";

export function SendMultipartRequest(url: string, cookies: string, formData: FormData) {
    return fetch(url, {
        method: 'Post',
        body: formData,
        credentials: 'include',
        headers: {
            credentials: 'include',
            'Cookie': cookies, // TODO: does this need to be here? I think so for multipart
            'Access-Control-Allow-Origin': '*',
        },
    })
}

// TODO: USE THIS!
export function MainCollectionInputOrRead({label, onIdSelected, copyText}: {
    label?: string,
    onIdSelected: (s: string) => void
    copyText?: string,
}) {
    const {state, dispatch} = useRfidReaderContext()
    const [id, setId] = useState<string>("");
    const updateId = (newId: string) => {
        setId(newId)
        onIdSelected(newId)
    }
    return <div>
        {/* INPUT FOR MAINCOLLECTIONID */}
        <InputTextInlineTitle label={"ID TO:"} value={id} readonly={false} errorMessage={undefined/* TODO: ???*/} placeholder={"Destination"} onChange={(s)=>updateId(s || "")}/>
        {/*<TextBox label={label || "Main Collection Id Input: "} value={id} fieldName={"mainCollIdInput"}*/}
        {/*         updateTextHandler={updateId} readonly={false}/>*/}
        {/* BUTTON TO READ MAIN COLL ID */}
        <ReaderWriterSelector txt={"select rfid reader"} onSelect={(wr)=>{ // TODO: wr ok here or state.selected?
            ReadTagFunc(dispatch, undefined, wr).then(updateId)
        }} />
        {/*<RfidSelectorWithReadButton handleTagRead={updateId} readButtonTxt={"read from current tag reader"}*/}
        {/*                            readerWriterTxt={"select rfid reader"} onWriterSelect={(wr)=>{*/}
        {/*    ReadTagFunc(dispatch, undefined, state.selected).then(updateId)*/}
        {/*}}/>*/}
        {/* BUTTON TO USE LAST READ ID */}
        {state.lastReadTag !== undefined && <div>
            <button className={"basicButton"} onClick={() => {
                state.lastReadTag && setId(state.lastReadTag)
            }}>{copyText || "Copy last read id"}</button>
        </div>}
    </div>
}

export function AssertArrayResult<T>(input: any, validateEntry: (inp: any) => void): asserts input is T[] {
    if (!Array.isArray(input)) {
        throw new Error('not an array');
    }
    try {
        if (!CheckArrayType(input, validatorForAssertion(validateEntry))) {
            throw new Error('incorrect item types in array');
        }
    } catch (e) {
        throw e;
    }

    return
}

export function CheckArrayType<T>(arr: T[], typeChecker: (item: T) => boolean): boolean {
    return arr.every(typeChecker);
}

export function RequiredKey(key: string, input: any, validateType: (val: any) => boolean): boolean {
    return key in input && validateType(input[key])
}

// // TODO: Redundant get rid of?
// export function RequiredSimpleKey(key: string, input: any, expType: string): boolean {
//     return RequiredKey(key, input, (val: any) => {
//         return typeof val === expType
//     })
// }

export function OptionalKey(key: string, input: any, validateIfExists: (inp: any) => boolean): boolean {
    return (key in input) ? validateIfExists(input[key]) : true
}

export function OptionalSimpleKey(key: string, input: any, expType: string): boolean {
    return OptionalKey(key, input, IsType(expType))
}

export function IsType(finalType: string): (inpt: any) => boolean {
    return (inp: any) => {
        return typeof inp === finalType
    }
}

export function RequiredArrayOfType(key: string, input: any, validateChildren: (child: any) => boolean): boolean {
    return RequiredKey(key, input, (val: any): boolean => {
        return Array.isArray(val) && CheckArrayType(val, validateChildren)
    })
}

export function OptionalArrayOfType(key: string, input: any, validateChildren: (child: any) => boolean): boolean {
    return OptionalKey(key, input, (chd: any): boolean => {
        return Array.isArray(chd) && CheckArrayType(chd, validateChildren)
    })
}

// TODO: delete if unneeded
// export function OptionalMapOfType(key: string, input: any, validateChildren: (child: any) => boolean): boolean {
//     if (typeof input !== 'object') {
//         throw new Error('Input is not an object! Input is ' + typeof input);
//     }
//     return OptionalKey(key, input, (chd: any): boolean => {
//         return CheckArrayType(chd, validateChildren)
//     })
// }

export function ViewInNewTabButton({entryType,id}:{entryType:string,id:string}){
    return <EntryLinkWrapper props={{linkId: encodeURI(encodeURI(id)), entryType: entryType, openInNewTab: true}}>
        <button className={"basicButtonSmall"}>{"View"}</button>
    </EntryLinkWrapper>
}

export function ListItemsRequest(entryType:string){
    return fetch(BaseExternalUrl + "/db/list/"+entryType, {
        method: 'Get',
        credentials: 'include',
        headers: {
            credentials: 'include',
            'Accept': 'application/json',
        },
    }).then((res) => {
        if(!res.ok){
            throw new Error('response not ok. Status='+res.status+', body='+res.text())
        }
        return res.json().then(result=>{
            let asserter:(x:any)=>void = ()=>false
            switch(entryType){
                case "agarBatches": asserter = AssertAgarBatch; break;
                case "agarRecipes":asserter = AssertAgarRecipe; break;
                case "bags": asserter = AssertBag; break;
                case "fruits": asserter = AssertFruit; break;
                case "fruitingChambers": asserter = AssertFruitingChamber; break;
                case "grainBatches": asserter = AssertGrainBatch; break;
                case "jars": asserter = AssertJar; break;
                case "jarRecipes": asserter = AssertJarRecipe; break;
                case "lcs": asserter = AssertLc; break;
                case "lcRecipes": asserter = AssertLcRecipe; break;
                case "lcSyringes": asserter = AssertLcSyringe; break;
                case "mss": asserter = AssertMss; break;
                case "pcRuns": asserter = AssertPcRun; break;
                case "plates": asserter = AssertPlate; break;
                case "projects": asserter = AssertProject; break;
                case "sales": asserter = AssertSale; break;
                case "slants": asserter = AssertSlant; break;
                case "species": asserter = AssertSpecies; break;
                case "sporePrints": asserter = AssertSporePrint; break;
                case "sporeSwabs": asserter = AssertSporeSwab; break;
                case "stasisTubes": asserter = AssertStasisTube; break;
                case "subspecies": asserter = AssertSubspecies; break;
                case "substrateBatches": asserter = AssertSubstrateBatch; break;
                case "substrateRecipes": asserter = AssertSubstrateRecipe; break;
                case "transfers": asserter = AssertTransfer; break;
                case "users": asserter = AssertUser; break;
                case "waterJars": asserter = AssertWaterJar; break;
                default:
                    throw new Error("invalid type but got response. Should never happen"); break;
            }
            switch(entryType){
                case "agarRecipes":
                case "jarRecipes":
                case "lcRecipes":
                case "substrateRecipes":
                    AssertDualListResult(result, asserter); break;
                default:
                    AssertArrayResult(result, asserter); break;
            }
            return result
        })
    })
}

export function IsString(item: any): boolean {
    return typeof item === 'string'
}

export function IsBool(item: any): boolean {
    return typeof item === 'boolean'
}

export function HeaderLevel(lvl?: number) {
    return lvl || defaultHeaderLevel
}

export interface ListPageItems<T> {
    data: T[],
    onClick?: (v: T) => void
    withLink?: boolean,
}

export interface InlineProps<T> {
    data: T,
    expandByDefault?: boolean,
    onClick?: (v?: T) => void
    headerLevel?: number
    idIsLink?:boolean
    showMainPageButton?:boolean
}

export interface SingleListProps<T> {
    data: T[],
    onClick: (v: T) => void
}

export interface TwoListProps<T> {
    recent: T[],
    standard: T[],
    onClick: (v: T) => void
}

export function InlineSubArea(
    {
        props, children
    }: {
        props: {
            className?: string
        },
        children: ReactNode,
    }) {
    return <div data-cy-id="InlineSubAreaWrapper" className={props.className}>
        <div data-cy-id="InlineSubArea" className={"inlineSubArea"}>
            {children}
        </div>
    </div>
}

export function InlineExpansionArea(
    {
        props, children
    }: {
        props: {
            expanded?: boolean
        },
        children: ReactNode,
    }) {
    if (!props.expanded) {
        return null
    }
    return <InlineSubArea data-cy-id="InlineExpansionArea" props={{}}>
        {children}
    </InlineSubArea>
}

export function InlineExpansionButton(
    {
        setExpanded, expanded
    }: {
        setExpanded: (value: SetStateAction<boolean | undefined>) => void
        expanded?: boolean
    }) {
    return <div data-cy-id="InlineExpansionButtonWrapper">
        <button className={"basicButton"} data-cy-id="InlineExpansionButton" onClick={(e) => {
            e.stopPropagation();
            setExpanded(!expanded)
        }}>{expanded ? "See less" : "See more"}</button>
    </div>
}
export function TwoValuePlusUnknownSelector({pre, updateParent, initial, trueStr, falseStr,className}: {
    pre: string,
    updateParent?: (v?: boolean) => void,
    initial?: boolean,
    trueStr: string,
    falseStr: string
    className?: string
}) {
    if (initial !== undefined) {
        return <div>{pre + (initial ? trueStr : falseStr)}</div>
    }
    const strForBool = (s?: boolean) => {
        return ((s === undefined) ? "unknown" : (s ? trueStr : falseStr))
    }
    const [selected, setSelected] = useState<boolean | undefined>(initial)
    const boolForStr = (s: string) => {
        return ((s === "unknown") ? undefined : (s === trueStr))
    }

    const selectHandler = (e: SyntheticEvent<HTMLSelectElement, Event>) => {
        let val = boolForStr(e.currentTarget.value)
        updateParent && updateParent(val)
        setSelected(val)
    }
    return <div className={className}>
        <div>{pre}</div>
        <select className={"tailwindSelector"} value={strForBool(selected)} onChange={selectHandler}>
            <option value={"unknown"}>{"unknown"}</option>
            <option value={trueStr}>{trueStr}</option>
            <option value={falseStr}>{falseStr}</option>
        </select>
    </div>
}

export function ConfirmedCleanSelector(// TODO: validate works now via a test LC
    {updateParent, initial}:
    {
        updateParent:(b?:boolean)=>void, initial?: boolean
    }) {
    return <TwoValuePlusUnknownSelector pre={"Confirmed Clean: "} updateParent={updateParent} initial={initial} trueStr={"clean"} falseStr={"contaminated"}/>

    // const strForBool = (s?: boolean) => {
    //     return ((s === undefined) ? "unknown" : (s ? "clean" : "contaminated"))
    // }
    // const [selected, setSelected] = useState<string>(strForBool(initial))
    // const boolForStr = (s: string) => {
    //     return ((s === "unknown") ? undefined : (s === "clean"))
    // }
    //
    // const selectHandler = (e: SyntheticEvent<HTMLSelectElement, Event>) => {
    //     let val = e.currentTarget.value
    //     selProps.doSelect(boolForStr(val))
    //     setSelected(val)
    // }
    // return <div className={"confirmedCleanSelector"}>{/* TODO: STYLING!!!!*/}
    //     <div>{"Confirmed Clean: "}</div>
    //     <select className={"tailwindSelector"} value={selected} onChange={selectHandler}>
    //         <option value={"unknown"}>{"unknown"}</option>
    //         <option value={"clean"}>{"clean"}</option>
    //         <option value={"contaminated"}>{"contaminated"}</option>
    //     </select>
    // </div>
}

export function YesNoSelector({pre, updateParent, initial,className}: {
    pre: string,
    updateParent?: (v?: boolean) => void,
    initial?: boolean
    className?:string
}) {
    return <TwoValuePlusUnknownSelector pre={pre} updateParent={updateParent} initial={initial} trueStr={"yes"} falseStr={"no"} className={className}/>
}

export function ConfirmedCleanArea(
    {
        readonly, initial, headerLevel, onSelect
    }: {
        readonly?: boolean
        initial?: boolean
        headerLevel?: number
        onSelect?: (c?: boolean) => void
    }) {
    return <YesNoSelector pre={"Confirmed Clean:"} initial={initial} updateParent={onSelect}/>
    // if (readonly) {
    //     return <div className={"confirmedCleanArea"}>
    //         <div>{"Confirmed Clean:"}</div>
    //         <div>{(initial === undefined) ? "Unknown" : (initial ? "Yes" : "No")}</div>
    //     </div>
    // }
    // return <div className={"confirmedCleanArea"}><ConfirmedCleanSelector initial={initial} updateParent={(v)=> {
    //     onSelect && onSelect(v)
    // }
    // }/></div>

}

export type DisplayInput = {
    id: string;
    readonly: boolean;
    data: any
    headerLevel?: number
    isTopLevel: boolean
    cookies: string
}

export type ImportDisplayInput = {
    headerLevel: number
    cookies: string
}

export function DisposedContamArea( // TODO: THIS AND USE THIS WHEN NEEDED!!!
    {
        headerLevel, disposed, contams
    }: {
        disposed?: number
        contams?: Contamination[]
        headerLevel?: number
    }) {
    return <div>
        <TestAndValidate todos={["DisposedContamArea NOT IMPLEMENTED!"]}>
            {"DisposedContamArea NOT IMPLEMENTED!"}
        </TestAndValidate>
    </div> // TODO: THIS!
}

export function DisposedSaleContamArea(
    {
        contams, sale, disposed, headerLevel
    }: {
        contams?: Contamination[]
        sale?: string
        disposed?: number
        headerLevel?: number
    }) {
    const sectionHeader = <div>{"Status: "}</div>
    if (sale) {
        const displayId = sale
        return <div>
            {sectionHeader}
            <div>{"Sold in sale "}
                <EntryLink props={{
                    displayedId: displayId,
                    linkId: displayId,
                    entryType: "sale",
                    openInNewTab: true
                }}>{displayId}</EntryLink>
            </div>
        </div>
    }
    let contamToUse: Contamination = {time: 0, confirmed: false, mold: false, bacteria: false, location: ""}
    if (contams !== undefined && contams.length === 0) {
        contamToUse.time = contams[contams.length - 1].time
        for (let i = 0; i < contams.length; i++) {
            if (!contamToUse.confirmed && contams[i].confirmed) {
                contamToUse.confirmed = true
            }
            if (!contamToUse.mold && contams[i].mold) {
                contamToUse.mold = true
            }
            if (!contamToUse.bacteria && contams[i].bacteria) {
                contamToUse.bacteria = true
            }
            if (contams[i].location) {
                contamToUse.location = contams[i].location
            }
        }
    }
    let contamLine: JSX.Element | null = null
    if (contamToUse.mold || contamToUse.bacteria) {
        let contamType = contamToUse.mold ? "mold" : "bacteria"
        if (contamToUse.mold && contamToUse.bacteria) {
            contamType = "mold, bacteria"
        }
        let lastContamPart = (" last cited " + NumberToDate(new Date(contamToUse.time)))
        contamLine = <div>
            <div>{(contamToUse.confirmed ? "Confirmed" : "Unconfirmed") + " contamination (" + contamType + ")" + lastContamPart}</div>
        </div>
    }
    let disposedSection = <div>
        {disposed ? "Disposed on " + NumberToDate(new Date(disposed)) : "Available"}{/* TODO: DIFFERENT STYLING BASED ON ANSWER?*/}
    </div>
    return <div>
        <div>{sectionHeader}</div>
        {contamLine}
        {disposedSection}
    </div>
}

// TODO: del if unused
// export function SaleAndDisposedArea({sale, disposed, headerLevel, readonly}: { // TODO: USE THIS WHERE NEEDED!!!!
//     sale?: string,
//     disposed?: number,
//     headerLevel?: number,
//     readonly: boolean
// }) {
//     if (sale) {
//         return <SaleArea sale={sale} readonly={true} headerLevel={headerLevel} canCreateSale={false}/>
//     }
// }

export interface NewEntryIdInput {
    headerLevel?: number,
    onCreate?: (id: string) => void
    redirectOnCreate: boolean
}


export interface NewEntryInput<T> {
    isTopLevel: boolean
    onCreate?: (newItem: T) => void
}

// TODO: MOVE THIS
export async function getTypeFor(id: string) { // TODO: ensure this works????
    // TODO: USE EXAMPLE ITEMS FOR DEV ENVIRONMENT!
    return await fetch(BaseExternalUrl + "/typeOf/" + id, {
        method: "GET",
        headers: {
            credentials: 'include',
            SessionId: "FIXME!!!", // TODO; THIS
        },
    }).then(HandleTxtResponse)
        .then((entryType) => {
            return entryType
        })
        .catch((error) => {
            throw error
        });
}

function webUrl(subPath: string) {
    return BaseExternalUrl + subPath
}

function apiUrl(subPath: string) {
    return BaseExternalUrl + "/db" + subPath
}

// TODO: use all of these all over the place!
export function viewUrlFor(itemType: string, newId: string) {
    return webUrl("/view/" + itemType + "/" + newId)
}

export function viewApiUrlFor(itemType: string, id: string) {
    return apiUrl("/get/" + itemType + "/" + id)
}

export function createUrlFor(itemType: string) {
    return webUrl("/create/" + itemType)
}

export function createApiUrlFor(itemType: string) {
    return apiUrl("/create/" + itemType)
}

export function importUrlFor(itemType: string) {
    return webUrl("/import/" + itemType)
}

export function importApiUrlFor(itemType: string) {
    return apiUrl("/import/" + itemType)
}

export function updateApiUrlFor(itemType: string, id: string) {
    return apiUrl("/update/" + itemType + "/" + id) // TODO: ensure ok
}


// TODO: fix inputs and use this everywhere we can???
export function CreateNewEntryButton(handler: { onSubmit: () => void }) {
    return <button className={"greenButton buttonFullWidth"} onClick={handler.onSubmit}>{"Create!"}</button>
}

export function resolvePicsFormData(picsIn: SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>) {
    let newImages: File[] = new Array(picsIn.new.length)
    let dataOut = {existing: picsIn.existing, new: new Array(picsIn.new.length)}
    for (let i = 0; i < picsIn.new.length; i++) {
        let toSend = picsIn.new[i].data
        if (toSend.img === undefined) {
            throw new Error("new image " + i + " is undefined")
            //setErr("new image " + i + " is undefined")
            //return
        }
        newImages[i] = toSend.img
        dataOut.new[i] = {
            time: toSend.time, notes: toSend.notes.new.map(n => {
                return n.data
            })
        }
    }
    return {
        images: newImages,
        obj: dataOut,
    }
}

export function resolveContamsFormData(inp: SplitAllEntries<ContaminationForm, NewContaminationForm>) {
    let conts: (File | undefined)[] = new Array(inp.new.length)
    let dataOut = {existing: inp.existing, new: new Array(inp.new.length)}
    for (let i = 0; i < inp.new.length; i++) {
        let toSend = inp.new[i].data
        dataOut.new[i] = {
            time: toSend.time,
            confirmed: toSend.confirmed,
            bacteria: toSend.bacteria,
            mold: toSend.mold,
            notes: toSend.notes
        }
        conts[i] = toSend.file
    }
    return {
        images: conts,
        obj: dataOut,
    }
}

export function setFormImages(formData: FormData, filePrefix: string, pics: any[]) {
    for (let i = 0; i < pics.length; i++) {
        const fileName = filePrefix + "-" + i
        if (pics[i] === undefined) {
            console.log("Picture undefined, " + fileName)
            continue
        }
        console.log("Picture set, " + fileName)
        formData.set(fileName, pics[i], fileName)
    }
}

export function setFormData(formData: FormData, dataObj: any) {
    formData.set("data", JSON.stringify(dataObj))
}

export function HandleJsonResponse(res: Response): Promise<any> {
    checkResponseStatus(res)
    return res.json()
}

export function HandleTxtResponse(res: Response): Promise<string> {
    checkResponseStatus(res)
    return res.text()
}

function checkResponseStatus(res: Response) {
    if (!res.ok || res.status !== 200) {
        throw "[(response status " + res.status + " " + res.statusText + ")]"
    }
    return
}
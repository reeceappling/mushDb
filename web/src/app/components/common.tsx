'use client'

import {defaultHeaderLevel} from "@/app/components/formSubcomponents/utils/headers";
import * as React from "react";
import {JSX, ReactNode, SetStateAction, SyntheticEvent, useContext, useEffect, useState} from "react";
import {
    ContaminationForm,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import EntryLinkForId, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {Data, ListResult, SplitAllEntries} from "@/app/components/formSubcomponents/shared";
import {NewPicWithNotesForm, PicWithNotesForm} from "@/app/components/formSubcomponents/picWithNotes";
import {BaseExternalUrl} from "@/app/components/Constants";
import ReaderWriterSelector, {
    ReadTagFunc,
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {useRfidReaderContext} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import {
    AssertSubstrateRecipe,
} from "@/app/components/substrateRecipeClient";
import { InputTextInlineTitle} from "@/app/components/formSubcomponents/numericInput";
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
import {AssertUser} from "./userClient";
import {AssertWaterJar} from "@/app/components/waterJarClient";
import {AssertTransfer} from "@/app/components/transferClient";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {DepthContext, DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import {FruitData} from "@/app/components/fruitServer";
import {Actions, ActionTypes} from "@/app/components/formSubcomponents/modalContext/modal";
import {allCookies} from "@/app/components/formSubcomponents/cookiesContext/cookies";

export const clientPostRequestHeaders = {
    credentials: 'include',
    'Access-Control-Allow-Origin': BaseExternalUrl || "*", // TODO: FIXME! maybe "*"?
    'Content-type': "application/json",
    'Accept': "application/json", // TODO: ensure ok
}
export const clientGetRequestHeaders = {
    credentials: 'include',
    'Access-Control-Allow-Origin': BaseExternalUrl || "*", // TODO: FIXME! maybe "*"?
    'Accept': "application/json", // TODO: ensure ok
}

export const clientPostRequestHeadersMultipart = {
    credentials: 'include',
    'Access-Control-Allow-Origin': BaseExternalUrl || "*", // TODO: FIXME! maybe "*"?
    // 'Content-type': "multipart/form-data", // If this is set, it will not work (bounds not auto-calculated)
    'Accept': "application/json",
}

// TODO: remove cookies from args if it works without
export function SendMultipartRequest(url: string, formData: FormData, cookies: string) {
    return fetch(url, {
        method: 'Post',
        body: formData,
        credentials: 'include',
        headers: clientPostRequestHeadersMultipart,
    })
}



// TODO: USE THIS!
export function MainCollectionInputOrRead({label, placeholder, onIdSelected, copyText}: {
    label?: string,
    placeholder?: string,
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
        <InputTextInlineTitle label={(label || "ID TO")+":"} value={id} readonly={false} errorMessage={undefined/* TODO: ???*/}
                              placeholder={placeholder || "Destination"} onChange={(s) => updateId(s || "")}/>
        {/*<TextBox label={label || "Main Collection Id Input: "} value={id} fieldName={"mainCollIdInput"}*/}
        {/*         updateTextHandler={updateId} readonly={false}/>*/}
        {/* BUTTON TO READ MAIN COLL ID */}
        <ReaderWriterSelector txt={"select rfid reader"} onSelect={(wr) => { // TODO: wr ok here or state.selected?
            ReadTagFunc(dispatch, undefined, wr).then(updateId)
        }}/>
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

export function OptionalKeyNew(key: string, input: any, validate: (inp: any) => void): void {
    if (key in input && (!(input[key] === undefined || input[key] === null))) {
        validate(input[key])
        return
    }
    return
}

export function OptionalSimpleKey(key: string, input: any, expType: string): boolean {
    return OptionalKey(key, input, IsType(expType))
}

export function OptionalSimpleKeyNew(key: string, input: any, expType: string): void {
    return OptionalKeyNew(key, input, IsTypeNew(expType))
}

export function IsType(finalType: string): (inpt: any) => boolean {
    return (inp: any) => {
        return typeof inp === finalType
    }
}

export function IsTypeNew(finalType: string): (inpt: any) => void {
    return (inp: any) => {
        const typ = typeof inp
        if (!(typ === finalType)) {
            throw 'field type was not '+finalType+", was "+typ
        }
        return
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

export function ViewInNewTabButton<T extends Entry>({entry}: { entry:T}) {
    return <EntryLinkWrapper props={{entry:entry, openInNewTab: true}}>
        <button className={"basicButtonSmall"}>{"View"}</button>
    </EntryLinkWrapper>
}

export function ListItemsRequest(entryType: string, hideDisposed: boolean = false) {
    return fetch(BaseExternalUrl + "/db/list/" + entryType+(hideDisposed?"?hideDisposed=true":""), { // TODO: ensure hiding disposed works!
        method: 'Get',
        credentials: 'include',
        headers: clientPostRequestHeaders,
    }).then((res) => {
        if (!res.ok) {
            throw new Error('response not ok. Status=' + res.status + ', body=' + res.text())
        }
        return res.json().then(result => {
            let asserter: (x: any) => void = () => false
            switch (entryType) {
                case "agarBatches":
                    asserter = AssertAgarBatch;
                    break;
                case "agarRecipes":
                    asserter = AssertAgarRecipe;
                    break;
                case "bags":
                    asserter = AssertBag;
                    break;
                case "fruits":
                    asserter = AssertFruit;
                    break;
                case "fruitingChambers":
                    asserter = AssertFruitingChamber;
                    break;
                case "grainBatches":
                    asserter = AssertGrainBatch;
                    break;
                case "jars":
                    asserter = AssertJar;
                    break;
                case "jarRecipes":
                    asserter = AssertJarRecipe;
                    break;
                case "lcs":
                    asserter = AssertLc;
                    break;
                case "lcRecipes":
                    asserter = AssertLcRecipe;
                    break;
                case "lcSyringes":
                    asserter = AssertLcSyringe;
                    break;
                case "mss":
                    asserter = AssertMss;
                    break;
                case "pcRuns":
                    asserter = AssertPcRun;
                    break;
                case "plates":
                    asserter = AssertPlate;
                    break;
                case "projects":
                    asserter = AssertProject;
                    break;
                case "sales":
                    asserter = AssertSale;
                    break;
                case "slants":
                    asserter = AssertSlant;
                    break;
                case "species":
                    asserter = AssertSpecies;
                    break;
                case "sporePrints":
                    asserter = AssertSporePrint;
                    break;
                case "sporeSwabs":
                    asserter = AssertSporeSwab;
                    break;
                case "stasisTubes":
                    asserter = AssertStasisTube;
                    break;
                case "subspecies":
                    asserter = AssertSubspecies;
                    break;
                case "substrateBatches":
                    asserter = AssertSubstrateBatch;
                    break;
                case "substrateRecipes":
                    asserter = AssertSubstrateRecipe;
                    break;
                case "transfers":
                    asserter = AssertTransfer;
                    break;
                case "users":
                    asserter = AssertUser;
                    break;
                case "waterJars":
                    asserter = AssertWaterJar;
                    break;
                default:
                    throw new Error("invalid type but got response. Should never happen");
                    break;
            }
            switch (entryType) {
                case "agarRecipes":
                case "jarRecipes":
                case "lcRecipes":
                case "substrateRecipes":
                    AssertDualListResult(result, asserter);
                    break;
                default:
                    AssertArrayResult(result, asserter);
                    break;
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
    idIsLink?: boolean
    showMainPageButton?: boolean
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

export function TwoValuePlusUnknownSelector({pre, updateParent, initial, trueStr, falseStr, className}: {
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
        const val = boolForStr(e.currentTarget.value)
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
        updateParent: (b?: boolean) => void, initial?: boolean
    }) {
    return <TwoValuePlusUnknownSelector pre={"Confirmed Clean: "} updateParent={updateParent} initial={initial}
                                        trueStr={"clean"} falseStr={"contaminated"}/>
}

export function YesNoSelector({pre, updateParent, initial, className}: {
    pre: string,
    updateParent?: (v?: boolean) => void,
    initial?: boolean
    className?: string
}) {
    return <TwoValuePlusUnknownSelector pre={pre} updateParent={updateParent} initial={initial} trueStr={"yes"}
                                        falseStr={"no"} className={className}/>
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
    if (readonly) {
        return <div className={"inlineChildren"}>
            <div>{"Confirmed Clean:"}</div>
            <div>{(initial === undefined) ? "Unknown" : (initial ? "Yes" : "No")}</div>
        </div>
    }
    return <YesNoSelector pre={"Confirmed Clean:"} initial={initial} updateParent={onSelect}/>


}

export type DisplayInput<T extends Entry> = {
    id: string;
    readonly: boolean;
    data: T
    headerLevel?: number
    isTopLevel: boolean
}

export type ImportDisplayInput = {
    headerLevel: number
}

// export function DisposedContamArea( // TODO: THIS AND USE THIS WHEN NEEDED!!!
//     {
//         headerLevel, disposed, contams
//     }: {
//         disposed?: number
//         contams?: Contamination[]
//         headerLevel?: number
//     }) {
//     return <div>
//         <TestAndValidate todos={["DisposedContamArea NOT IMPLEMENTED!"]}>
//             {"DisposedContamArea NOT IMPLEMENTED!"}
//         </TestAndValidate>
//     </div> // TODO: THIS!
// }
//
// export function DisposedSaleContamArea(
//     {
//         contams, sale, disposed, headerLevel
//     }: {
//         contams?: Contamination[]
//         sale?: string
//         disposed?: number
//         headerLevel?: number
//     }) {
//     const sectionHeader = <div>{"Status: "}</div>
//     if (sale) {
//         const displayId = sale
//         return <div>
//             {sectionHeader}
//             <div>{"Sold in sale "}
//                 <EntryLinkForId props={{
//                     displayId: displayId,
//                     linkId: displayId,
//                     entryType: "sale",
//                     openInNewTab: true
//                 }}/>
//             </div>
//         </div>
//     }
//     const contamToUse: Contamination = {time: 0, confirmed: false, mold: false, bacteria: false, location: ""}
//     if (contams !== undefined && contams.length === 0) {
//         contamToUse.time = contams[contams.length - 1].time
//         for (let i = 0; i < contams.length; i++) {
//             if (!contamToUse.confirmed && contams[i].confirmed) {
//                 contamToUse.confirmed = true
//             }
//             if (!contamToUse.mold && contams[i].mold) {
//                 contamToUse.mold = true
//             }
//             if (!contamToUse.bacteria && contams[i].bacteria) {
//                 contamToUse.bacteria = true
//             }
//             if (contams[i].location) {
//                 contamToUse.location = contams[i].location
//             }
//         }
//     }
//     let contamLine: JSX.Element | null = null
//     if (contamToUse.mold || contamToUse.bacteria) {
//         let contamType = contamToUse.mold ? "mold" : "bacteria"
//         if (contamToUse.mold && contamToUse.bacteria) {
//             contamType = "mold, bacteria"
//         }
//         const lastContamPart = (" last cited " + NumberToDate(new Date(contamToUse.time)))
//         contamLine = <div>
//             <div>{(contamToUse.confirmed ? "Confirmed" : "Unconfirmed") + " contamination (" + contamType + ")" + lastContamPart}</div>
//         </div>
//     }
//     const disposedSection = <div>
//         {disposed ? "Disposed on " + NumberToDate(new Date(disposed)) : "Available"}{/* TODO: DIFFERENT STYLING BASED ON ANSWER?*/}
//     </div>
//     return <div>
//         <div>{sectionHeader}</div>
//         {contamLine}
//         {disposedSection}
//     </div>
// }

// TODO: del if unused
// export function SaleAndDisposedArea({sale, disposed, headerLevel, readonly}: { // TODO: USE THIS WHERE NEEDED!!!!
//     sale?: string,
//     disposed?: number,
//     headerLevel?: number,
//     readonly: boolean
// }) {
//     if (sale) {
//         return <SaleArea sale={sale} readonly={true} canCreateSale={false}/>
//     }
// }

export interface NewEntryInput<T> {
    isTopLevel: boolean
    onCreate?: (newItem: T) => void
}

// // TODO: MOVE THIS
// export async function getTypeFor(id: string) { // TODO: ensure this works????
//     // TODO: USE EXAMPLE ITEMS FOR DEV ENVIRONMENT!
//     return await fetch(BaseExternalUrl + "/typeOf/" + id, {
//         method: "GET",
//         headers: clientPostRequestHeaders,
//     }).then(HandleTxtResponse)
//         .then((entryType) => {
//             return entryType
//         })
//         .catch((error) => {
//             throw error
//         });
// }

export async function getPathFor(id: string) { // TODO: ensure this works????
    const resp = await fetch(BaseExternalUrl + "/db/pathFor/" + id, {
        method: "GET",
        headers: clientPostRequestHeaders,
    })
    if (!resp.ok) {
        throw "failed to get path for id"
    }
    return await resp.text()
}

export function webUrl(subPath: string) {
    return BaseExternalUrl + subPath
}

function apiUrl(subPath: string) {
    return BaseExternalUrl + "/db" + subPath
}

export function viewUrlFor(itemType: string, newId: string) {
    return webUrl("/view/" + itemType + "/" + newId)
}

export function viewApiUrlFor(itemType: string, id: string) {
    return apiUrl("/get/" + itemType + "/" + id)
}
export function getUrlFor(itemType: string, id: string) {
    return viewApiUrlFor(itemType, id)
}

export function createUrlFor(itemType: string) {
    return webUrl("/new/" + itemType)
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
    return apiUrl("/update/" + itemType + "/" + id)
}


// TODO: fix inputs and use this everywhere we can???
export function CreateNewEntryButton(handler: { onSubmit: () => void }) {
    return <button className={"greenButton buttonFullWidth"} onClick={handler.onSubmit}>{"Create!"}</button>
}

export function resolvePicsFormData(picsIn: SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>) {
    const newImages: File[] = new Array(picsIn.new.length)
    const dataOut = {existing: picsIn.existing, new: new Array(picsIn.new.length)}
    for (let i = 0; i < picsIn.new.length; i++) {
        const toSend = picsIn.new[i]
        if (toSend.img === undefined) {
            throw new Error("new image " + i + " is undefined")
        } else {
            newImages[i] = toSend.img
        }
        dataOut.new[i] = {
            time: toSend.time,
            notes: toSend.notes.new.map(n => {
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
    const conts: (File | undefined)[] = new Array(inp.new.length)
    const dataOut = {existing: inp.existing, new: new Array(inp.new.length)}
    for (let i = 0; i < inp.new.length; i++) {
        dataOut.new[i] = {
            time: inp.new[i].time,
            confirmed: inp.new[i].confirmed,
            bacteria: inp.new[i].bacteria,
            mold: inp.new[i].mold,
            notes: inp.new[i].notes
        }
        conts[i] = inp.new[i].file
    }
    return {
        images: conts,
        obj: dataOut,
    }
}

export function setFormImages(filePrefix: string, formData: FormData, pics: any[]) {
    for (let i = 0; i < pics.length; i++) {
        const fileName = filePrefix + "-" + i
        if (pics[i] === undefined) {
            console.log("Picture undefined, " + fileName) // TODO: DEL
            continue
        }
        console.log("Picture set, " + fileName) // TODO: DEL
        formData.set(fileName, pics[i], fileName)
    }
}

export function setFormFull(formData: FormData, dataObj: any, pics?: any[], contams?: any[], flushes?: any[]) {
    formData.set("data", JSON.stringify(dataObj))
    if (pics && pics.length > 0){
        setFormImages("newPic", formData, pics)
    }
    if (contams && contams.length > 0){
        setFormImages("newContam", formData, contams)
    }
    if (flushes && flushes.length > 0){
        setFormImages("newFlush", formData, flushes)
    }
}

export function setFormData(formData: FormData, dataObj: any) {
    formData.set("data", JSON.stringify(dataObj))
}

export function HandleJsonResponse(res: Response): Promise<any> {
    checkResponseStatus(res)
    return res.json()
}

export interface Importable { // TODO: USED IN IMPORT PAGES
    _id: string
}

export function EntryUrlId(item: Entry){
    return (item && typeof (item as any).getIdUrlEncoded === "function") ? (item as any).getIdUrlEncoded() : item.getId()
}

export interface Entry { // TODO: USE!
    getId(): string;
    entryType(): string;
}
export interface StringNameEntry extends Entry { // TODO: USE!
    getIdUrlEncoded(): string;
}
type TypeAsserter<T> = (value: unknown) => asserts value is T; // TODO: USE THIS! MOVE THIS!

export function ImportResponseHandler<T extends Importable>(asserter: TypeAsserter<T>, typeStr: string, setErr: (e:any)=>void): (res: Response)=>void {
    return (res: Response)=>{
        HandleJsonResponse(res)
            .then(item=>{
                asserter(item)
                window.location.assign(viewUrlFor(typeStr, item._id))
            })
            .catch(ErrHandler(setErr))
    }
}

// TODO: use everywhere! validate working!
export function DoImportRequest<T extends Importable>(body: any, typeStr: string, asserter: TypeAsserter<T>, setErr: (e:any)=>void, cookies: string) {
    fetch(importApiUrlFor(typeStr), {
        method: "POST",
        headers: clientPostRequestHeaders,
        body: JSON.stringify(body)
    })
        .then(HandleJsonResponse)
        .then(newItem => {
            asserter(newItem)
            window.location.assign(viewUrlFor(typeStr, newItem._id))
            // redirect(viewUrlFor(typeStr, newItem._id)) // TODO: del if working
        })
        .catch(ErrHandler(setErr));
}

export function DoGetRequest<T extends Entry>(itemType: string, typeStr: string, asserter: TypeAsserter<T>, setErr: (e:any)=>void): Promise<T|undefined> {
    return fetch(viewApiUrlFor(itemType, typeStr), {
        method: "GET",
        headers: clientPostRequestHeaders,
    }).then(HandleJsonResponse)
        .then(newItem => {
            asserter(newItem)
            return newItem
        })
        .catch(e=>{
                ErrHandler(setErr)(e)
            return undefined
        });
}

export function DoMultipartImportRequest<T extends Importable>(formData: FormData, typeStr: string, asserter: TypeAsserter<T>, setErr: (e:any)=>void, cookies: string, dispatchUpdate:(isErr:boolean,text:string)=>void) {
    SendMultipartRequest(importApiUrlFor(typeStr), formData, cookies)
        .then(ImportResponseHandler(asserter,typeStr, setErr))
        .catch(caughtErr=>{
            const newErr = JSON.stringify(caughtErr)
            setErr(newErr)
            dispatchUpdate(true, newErr)
    })
}

export function HandleTxtResponse(res: Response): Promise<string> {
    checkResponseStatus(res)
    return res.text()
}

export function ErrHandler(setErr: (err:any)=>void): (err:any)=>void {
    return (e: any) => {
        setErr("error: "+JSON.stringify(e))
    }
}

function checkResponseStatus(res: Response) {
    if (!res.ok || res.status !== 200) {
        throw "[(response status " + res.status + " " + res.statusText + ")]"
    }
    return
}

export function FlexedArea(props: React.PropsWithChildren<{}>) {
    return <div className={"flexedArea"}>{props.children}</div>
}

export function FlexedSinglesGroup(props: React.PropsWithChildren<{}>) {
    return <div className={"flexedSinglesGroup"}>{props.children}</div>
}

export function ListPageTableRow<T>(props: React.PropsWithChildren<{ data: T, onClick: (item: T) => void, className?: string }>) {
    return <tr className={"listPageTableRow nonHeaderRow"+(props.className?" "+props.className : "")} onClick={() => {
        props.onClick && props.onClick(props.data)
    }}>{props.children}</tr>
}

export interface ListTableColumn<T> {
    key: string
    f: (v:T)=>string
    fit:boolean
}

export function NewColumn<T>(key:string,f:(v:T)=>any,fit?:boolean):ListTableColumn<T> {
    return {key:key,f:f,fit:fit||false}
}

export function ListPageTable<T extends Entry>({data, onClick, cols,className, newClass}: {
    data: T[],
    onClick?: (v: T) => void,
    cols: ListTableColumn<T>[],
    className?: string,
    newClass: (inp: any)=>T,
    // TODO: give this a reload button????
}){
    //const [hidden, setHidden] = useState<boolean[]>(data.map(d=>false))
    const classes = cols.map(c=>{
        return "text-left"+(c.fit ? " fit" : "")
    })
    return <table className={"listPageTable"}>
        <tr className={"listPageTableRow headerRow"}>
            {cols.map((col,i)=>{
                // if (hidden[i]) {
                //     return null
                // }
                return <th className={classes[i]} key={i} >{col.key}</th>
            })}
        </tr>
        {data.map(newClass).map((item,i) => {
            return <ListPageTableRow className={className} key={i} data={item} onClick={(v)=>{onClick && onClick(v)}}>{/* TODO: ADD EXPANSION???*/}
                {cols.map((col,i)=>{
                    // if (hidden[i]) {
                    //     return null
                    // }
                    return <td className={classes[i]} key={i}>{col.f(item)}</td>
                })}
            </ListPageTableRow>
        })}
    </table>
}

export function NumberToDateStr(n: number): string {
    const d = new Date(n)
    return (d.getMonth()+1)+"/"+d.getDate()+"/"+d.getFullYear()
}

export function ExistingDualSelector<T>(props: React.PropsWithChildren<{
    doSelect: (val?: T) => void,
    table: (items: T[],onSelect: (v?: T)=>void) => JSX.Element,
    entryType:string,
    entryTypes:string,
    asserter: (val: any)=>void
}>){
    const [err, setErr] = useState<string | undefined>(undefined)
    const [loaded, setLoaded] = React.useState(false);
    const [data, setData] = React.useState<ListResult<T> | undefined>(undefined);
    useEffect(()=>{ListItemsRequest(props.entryTypes).then((result) => {
        try {
            AssertDualListResult<T>(result, props.asserter)
            setData(result)
            setLoaded(true)
            return
        } catch (e) {
            console.error(JSON.stringify(e))
            throw e
        }
    }).catch(e => {
        console.error(JSON.stringify(e))
        setErr("error on listItems request: " + JSON.stringify(e))
    })},[])
    if (!loaded || data === undefined) {
        return <div>
            <ErrorDisplay err={err}/>
            <div>{"Loading "+props.entryType+" Selector"}</div>
        </div>
    }
    return <Subform>
        <ErrorDisplay err={err}/>
        <SelectorTableWithHeader header={"Recent"} data={data?.recent} onSelect={props.doSelect} table={props.table}/>
        <SelectorTableWithHeader header={"Standard"} data={data?.standard} onSelect={props.doSelect} table={props.table}/>
        <SelectorCreationArea>{props.children}</SelectorCreationArea>
    </Subform>
}

export function SelectorCreationArea(props:React.PropsWithChildren<{}>){
    const [creatorOpen, setCreatorOpen] = React.useState(false);
    if (!props.children){
        return null
    }
    if (!creatorOpen){
        return <button className={"buttonFullWidth basicButtonSmall"} onClick={e=>{
            e.stopPropagation();
            setCreatorOpen(true);
        }}>{"Create one instead"}</button>
    }
    return <><button className={"basicButtonSmall"} onClick={e=>{
        e.stopPropagation();
        setCreatorOpen(false);
    }}>{"Close creator"}</button>
        {props.children}
    </>
}

export function ExistingRecentSelector<T extends Entry>(props: React.PropsWithChildren<{
    doSelect: (val?: T) => void,
    table: (items: T[],onSelect: (v?: T)=>void) => JSX.Element,
    entryType:string,
    entryTypes:string,
    asserter: (val: any)=>void,
    hideDisposed?: boolean,
}>){
    const [err, setErr] = useState<string | undefined>(undefined)
    const [loaded, setLoaded] = React.useState(false);
    const [data, setData] = React.useState<T[] | undefined>(undefined);
    useEffect(()=>{ListItemsRequest(props.entryTypes, props.hideDisposed).then((result) => {
        try {
            AssertArrayResult<T>(result, props.asserter)
            setLoaded(true)
            setData(result)
        } catch (e) {
            throw e
        }
    }).catch(e => {
        setErr("error on listItems request: " + JSON.stringify(e))
    })},[])
    if (!loaded || data === undefined) {
        return <div>
            <ErrorDisplay err={err}/>
            <div>{"Loading "+props.entryType+" Selector"}</div>
        </div>
    }
    return <Subform>
        <ErrorDisplay err={err}/>
        <SelectorTableWithHeader header={"Recent"} data={data} onSelect={props.doSelect} table={props.table}/>
        <SelectorCreationArea>{props.children}</SelectorCreationArea>
    </Subform>
}

export function SelectorTableWithHeader<T>({header, data,table,onSelect}:{
    header: string,
    data?: T[],
    onSelect: (val?: T) => void,
    table:(items: T[], onSelect: (v?: T) => void)=>JSX.Element
}){
    if(!data || data.length===0){
        return null
    }
    return <>
        <div className={"text-xl"}>{header}</div>
        {table(data,onSelect)}
    </>
}

// // TODO: may disappear
// export function InlineEntry(props: React.PropsWithChildren<{ onClick?: () => void }>) { // TODO: ADD THIS TO ALL INLINES!!!!!
//     return <div className={"inlineEntry"} onClick={(e) => {
//         e.stopPropagation()
//         props.onClick && props.onClick()
//     }}>
//         {props.children}
//     </div>
// }
//
export function dataFor<Type>(vals?: Type[]): Data<Type>[] {
    return (vals || []).map((l) => {
        return {data: l, disabled: false}
    })
}

//InputDecimal
// export function FloatInput({initial, onChange}: { initial?: number, onChange: (value: number) => void }) {
//     const [val, setVal] = useState<number>(initial || 0)
//     const updateNumber = (s: string) => {
//         try {
//             const n = Number(s)
//             setVal(n)
//             onChange(n)
//         } catch (e) {
//             console.error("failed to set float input for "+s+" "+JSON.stringify(e))
//         }
//     }
//     return <div>
//         <TestAndValidate todos={["validate working properly"]}>
//             <InputNumber min={0} max={10000} onChange={s => {
//                 s && updateNumber(s)
//             }} step={1} mode={Modes.floating} value={val.toString()} readonly={false}/>
//         </TestAndValidate>
//     </div>
// }

//InputDecimal
// export function DecimalInput({initial, onChange}: { initial?: number, onChange: (value: number) => void }) {
//     const [val, setVal] = useState<number>(initial || 0)
//     const updateNumber = (s: string) => {
//         try {
//             const n = Number(s)
//             setVal(n)
//             onChange(n)
//         } catch (e) {
//             console.error("failed to set float input for "+s+" "+JSON.stringify(e))
//         }
//     }
//     return <div>
//         <TestAndValidate todos={["validate working properly"]}>
//             <InputNumber min={0} max={10000} onChange={s => {
//                 s && updateNumber(s)
//             }} step={1} mode={"floating"} value={val.toString()} readonly={false}/>
//         </TestAndValidate>
//     </div>
// }

export function SelectorWrapper<T>(props: React.PropsWithChildren<{
    title: string,
    current?: T,
    nameFunc: (item: T) => string
}>) {
    const [isOpen, setIsOpen] = useState(!props.current);
    useEffect(() => {
        setIsOpen(false) // TODO: it does not like that we are calling a setState in a useEffect
    }, [props.current])
    if (isOpen) {
        return <div>
            <div>{props.title}</div>
            <button className={"basicButtonSmall"} onClick={e => {
                e.stopPropagation();
                setIsOpen(false);
            }}>{"close selector"}</button>
            {props.children}
        </div>
    }
    if (props.current === undefined) {
        return <div className={"inlineChildren"}>
            <div>{props.title + ": "}</div>
            <button className={"basicButtonSmall"} onClick={e => {
                e.stopPropagation();
                setIsOpen(true);
            }}>{"select"}</button>
        </div>
    } else {
        return <div className={"inlineChildren"}>
            <div>{props.title + ": " + props.nameFunc(props.current)}</div>
            <button className={"basicButtonSmall"} onClick={e => {
                e.stopPropagation();
                setIsOpen(true);
            }}>{"select another"}</button>
        </div>
    }
}

function depthAndEntryClasses(depth: number, entryType?: string) {
    return " depth" + depth + (entryType ? " " + entryType : "")
}

export function NewEntryFormWrapper(props: React.PropsWithChildren<{ entryType: string, className?: string }>) { // TODO: USE THIS EVERYWHERE!
    const depth = useContext(DepthContext)
    return <DepthProvider>
        <div
            className={"subForm newEntryForm" + depthAndEntryClasses(depth, props.entryType) + (props.className ? " " + props.className : "")}>{/* TODO: likely not working as expected. +1?*/}
            {props.children}
        </div>
    </DepthProvider>
}

export function ImportEntryFormWrapper(props: React.PropsWithChildren<{ entryType: string }>) { // TODO: USE THIS EVERYWHERE!
    const depth = useContext(DepthContext)
    return <DepthProvider>
        <div
            className={"subForm importEntryForm" + depthAndEntryClasses(depth, props.entryType)}>{/* TODO: likely not working as expected. +1?*/}
            {props.children}
        </div>
    </DepthProvider>
}

export function DisplayFormWrapper(props: React.PropsWithChildren<{ entryType: string, id?: string }>) { // TODO: USE THIS EVERYWHERE!
    const depth = useContext(DepthContext)
    return <DepthProvider>
        <div id={props.id} className={"subForm displayForm" + depthAndEntryClasses(depth, props.entryType)}>{/* TODO: likely not working as expected. +1?*/}
                {props.children}
        </div>
    </DepthProvider>
}

export function Subform(props: React.PropsWithChildren<{}>) {
    const depth = useContext(DepthContext)
    return <DepthProvider>
        <div className={"subForm depth" + depth}>{/* TODO: likely not working as expected. +1?*/}
            {props.children}
        </div>
    </DepthProvider>
}

export function CreatedLinkFor({linkId, typ, linkText}: { linkId: string, typ: string, linkText?: string }) {
    return <EntryLinkForId props={{displayId: linkText || linkId, linkId: linkId, entryType: typ, openInNewTab: false/* TODO: ok?*/}}/>
}

export function AssertDualListResult<T>(input: any, validateEntry: (inp: any) => void): asserts input is ListResult<T> {
    if (typeof input !== 'object') {
        console.error('Input is not an object! Input is ' + typeof input)
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['recent', validatorForAssertion(validateEntry)], // TODO: ensure ok
        ['standard', validatorForAssertion(validateEntry)],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            console.error('optional array key ' + key + ' was not valid')
            throw new Error('optional array key ' + key + ' was not valid');
        }
    }
    return
}

// export function AssertSubRecipeListResult(input: any): asserts input is ListResult<SubstrateRecipeData> {
//     AssertDualListResult<SubstrateRecipeData>(input, AssertSubstrateRecipe)
// }

export function validatorForAssertion(asserter: ((input: any) => void)) {
    return (inp: any) => {
        try {
            asserter(inp)
            return true
        } catch (e) {
            console.error("error in validatorForAssertion: ", e)
            return false
        }
    }
}

export function DoCreateRequest<T>(entryType: string, body: any, asserter: TypeAsserter<T>, cookies: string): Promise<T> {
    return fetch(createApiUrlFor(entryType), {
        method: "POST",
        headers: {...clientPostRequestHeaders, 'Cookie': cookies},
        body: JSON.stringify(body)
    })
        .then(HandleJsonResponse)
        .then((entry:any):T => {
            asserter(entry)
            return entry
        })
}

export function DoCreateRequestMultipart<T>(entryType: string, formData: FormData, asserter: TypeAsserter<T>, cookies: string): Promise<T> {
    return SendMultipartRequest(createApiUrlFor(entryType), formData, cookies)
        .then(HandleJsonResponse)
        .then((entry):T => {
            asserter(entry)
            return entry
        })
}

export function DoUpdateRequest<T>(entryType: string, urlId: string, body: any, asserter: TypeAsserter<T>, cookies: string): Promise<T> {
    return fetch(updateApiUrlFor(entryType, urlId), {
        method: "POST",
        headers: {...clientPostRequestHeaders, 'Cookie': cookies},
        body: JSON.stringify(body)
    }).then(HandleJsonResponse)
        .then((entry) => {
            asserter(entry)
            return entry
        })
}
export function DoUpdateMultipartRequest<T>(entryType: string, urlId: string, formData: FormData, asserter: TypeAsserter<T>, cookies: string): Promise<T> {
    return SendMultipartRequest(updateApiUrlFor(entryType,urlId), formData, cookies)
        .then(HandleJsonResponse)
        .then((entry) => {
            asserter(entry)
            return entry
        })
}

// TODO: DICTAPHONES SHOULD BE USED IN:
// TODO: creates: anything that needs a sterile environment (LIST)
// TODO: views: all of them!
// TODO: consider embedding dictaphones in notes areas for views and creates, and controlling the notes with a context of some sort?
// export function Dictaphone({createNoteHandler}: { createNoteHandler?: (note: string) => void }) {
//     // const cmds = ["simon says", "new note"]
//     const timeoutRef = useRef<NodeJS.Timeout | undefined>(undefined)
//     const [activeCommand, setActiveCommand] = useState<string | undefined>(undefined)
//     // TODO: const [startedBody, setStartedBody] = useState(false)
//     const listenArgs = {
//         continuous: true, // TODO: ok? was false
//         interimResults: true, // TODO: ok? was false
//         language: "en-US",
//     }
//
//     // const startBodyListener = ()=>{
//     //
//     // }
//     // const startCommandListener = ()=>{
//     //
//     // }
//     //
//     // //const fullCmdRegex = new RegExp("(?<=^command )simon says [a-zA-Z0-9 ]+(?= end dictation)")
//     // //const startDictationString = "command"
//     // const resetString = "clear dictation"
//     // const resetDictationRegex = new RegExp(resetString, "g")
//     // const endBodyString = "end dictation"
//     // const endDictationRegex = new RegExp("^[a-zA-Z0-9 ]+ "+endBodyString+"$)", "g")
//     // const bodyCommand = "* "+endBodyString
//     // const simonSaysRegex = regexForCmd("simon says")
//     // const cmdRegex = [simonSaysRegex]
//     // const removePrefix = (str: string, pre: string):string => {
//     //     str.slice(pre.length);
//     // }
//     // const bodyCallback = (command: string, resetTranscript:()=>void):void=>{
//     //     const body = command.substring(0,command.length-(2+endBodyString.length)) // TODO: ensure length right
//     //     switch(activeCommand){
//     //         case undefined:
//     //             // TODO: ERROR
//     //     }
//     // }
//     // const cmdCallback = (command: string, resetTranscript:()=>void):void => {
//     //     const commandAndBody = removePrefix(lessEnd, prefixes[0])
//     //     switch(command){
//     //         case cmds[0]: //simon says
//     //             setActiveCommand(cmds[0])
//     //             break;
//     //         default:
//     //     }
//     //     if (lessEnd.startsWith(prefixes[0])){
//     //         let body = removePrefix(lessEnd, prefixes[0])
//     //
//     //     }
//     //     resetTranscript()
//     // }
//     const commands = [
//         {
//             command: ["reset dictation", "clear transcript", "reset transcript"],
//             callback: () => {
//                 resetTranscript()
//                 setActiveCommand(undefined)
//             },
//             matchInterim: true,
//         },
//         {
//             command: ["repeat after me", "simon says"],
//             callback: () => {
//                 resetTranscript()
//                 setActiveCommand("repeat after me")
//             },
//             matchInterim: true,
//         },
//         {
//             command: [
//                 "new note",
//                 "create note",
//                 "create new note",
//                 "create a note",
//                 "create a new note",
//                 "make note",
//                 "make a note",
//                 "make new note",
//                 "make a new note",
//
//             ],
//             callback: () => {
//                 resetTranscript()
//                 setActiveCommand("create note")
//             },
//             matchInterim: true,
//         },
//     ]
//     const {transcript, listening, resetTranscript, browserSupportsSpeechRecognition} = useSpeechRecognition({
//         commands: commands,
//     });
//     // 3-Second Timeout Logic
//     useEffect(() => {
//         // Clear existing timeout each time a new transcript word is detected
//         if (timeoutRef.current) {
//             clearTimeout(timeoutRef.current);
//         }
//
//         // Set a new 3-second timer
//         const currentText = transcript
//         // TODO: handle 0-length transcripts?
//         const onTimeout = () => {
//             switch (activeCommand) {
//                 case "repeat after me":
//                     console.log("repeat after me: " + currentText)
//                     SayString(currentText)
//                     break;
//                 // TODO: CREATE PLATE? Bag, Slant, Transfer?
//                 case "create note":
//                     // TODO: repeat and ask to save??????
//                     console.log("created note: " + currentText)
//                     createNoteHandler && createNoteHandler(currentText)
//                     break;
//                 default:
//                     return
//                 // TODO: this!
//             }
//             setActiveCommand(undefined)
//             resetTranscript()
//         }
//         timeoutRef.current = setTimeout(onTimeout, 3000);
//
//         return () => clearTimeout(timeoutRef.current);
//     }, [transcript, activeCommand]);
//
//     if (!browserSupportsSpeechRecognition) {
//         return <span>{"Browser doesn't support speech recognition."}</span>;
//     }
//
//     return (
//         <div>
//             <p>{"Microphone: " + (listening ? 'on' : 'off')}</p>
//             <button onClick={e => {
//                 e.stopPropagation();
//                 SpeechRecognition.startListening(listenArgs)
//             }}>{"Start"}</button>
//             <button onClick={e => {
//                 e.stopPropagation();
//                 SpeechRecognition.stopListening()
//             }}>{"Stop"}</button>
//             <button onClick={e => {
//                 e.stopPropagation();
//                 resetTranscript()
//             }}>Reset
//             </button>
//             <p>{transcript}</p>
//         </div>
//     );
// };
//
// // TODO: USE ON TFID VIEW PAGES!
// // TODO: SHOULD ADD WHERE NEEDED
// // TODO: LIKELY NEEDS MAJOR OVERHAUL
// export function ViewPageDictaphone({doUpdate}: {
//     doUpdate: () => void
// }) {
//     const rfidRdr = useRfidReaderContext()
//     const dict = useDictationContext()
//     // TODO: let readerWriter = state.selected // TODO: or lastReaderUsed???
//     const listenArgs = {
//         continuous: true,
//         interimResults: true,
//         language: "en-US",
//     }
//     const handleViewById = (idToSearch: string) => {
//         getPathFor(idToSearch).then((path) => {
//             location.assign(webUrl("/view/" + path))
//         }).catch((err) => {
//             console.log("failed to get path for id: " + JSON.stringify(err))
//             SpeechRecognition.startListening(listenArgs)
//         })
//     }
//     const commands = [
//         {
//             command: ["create transfer", "new transfer"],
//             callback: () => {
//                 resetTranscript()
//                 SpeechRecognition.stopListening()
//                 dict.dispatch({type: ActionTypes.SET_CURRENT,payload:"create transfer"})
//             },
//             matchInterim: true,
//         },
//         {
//             command: ["view tag"], // TODO: ok?
//             callback: () => {
//                 SpeechRecognition.stopListening()
//                 resetTranscript()
//                 ReadTagFunc(rfidRdr.dispatch, undefined, rfidRdr.state.selected)
//                     .then(handleViewById)// redir to the new page
//                     .catch(e => {
//                         console.error("failed to read linking tag: " + JSON.stringify(e))
//                         SpeechRecognition.startListening(listenArgs)
//                     })
//             },
//             matchInterim: true,
//         },
//         {
//             command: ["submit updates"], // TODO: ok?
//             callback: () => {
//                 SpeechRecognition.stopListening()
//                 resetTranscript()
//                 doUpdate()
//                 SpeechRecognition.startListening(listenArgs)
//             },
//             matchInterim: true,
//         },
//         {
//             command: [
//                 "new note",
//                 "create note",
//                 "create new note",
//                 "create a note",
//                 "create a new note",
//                 "make note",
//                 "make a note",
//                 "make new note",
//                 "make a new note",
//                 "add a new note",
//                 "add new note",
//                 "add a note",
//                 "add note",
//             ],
//             callback: () => {
//                 SpeechRecognition.stopListening()
//                 resetTranscript()
//                 dict.dispatch({type: ActionTypes.SET_CURRENT,payload:"create note"})
//             },
//             matchInterim: true,
//         },
//     ]
//     const {transcript, listening, resetTranscript, browserSupportsSpeechRecognition} = useSpeechRecognition({
//         commands: commands,
//     });
//
//     if (!browserSupportsSpeechRecognition) {
//         return <span>{"Browser doesn't support speech recognition."}</span>;
//     }
//     useEffect(() => { // TODO: validate works right
//         if (dict.state.current === "main") {
//             SpeechRecognition.startListening(listenArgs)
//         }
//     }, [dict.state.current])
//
//     return (
//         <div>
//             <button onClick={e => {
//                 e.stopPropagation();
//                 SpeechRecognition.startListening(listenArgs)
//             }}>{"Enable Dictation"}</button>{/* TODO: dictation enablement in cookies? We want to be able to traverse pages without touching the screen*/}
//             <button onClick={e => {
//                 e.stopPropagation();
//                 SpeechRecognition.stopListening()
//             }}>{"Disable Dictation"}</button>
//         </div>
//     );
// };
//
// export function AddNoteDictaphone({parent,createNote}:{parent?:string,createNote:(s:string)=>void}){
//     // Always created in a state that is not listening by default
//     try {
//         const {state, dispatch} = useDictationContext()
//         const listenArgs = {
//             continuous: false,
//             interimResults: false, // TODO: UNSURE IF WE WANT THIS OR NOT
//             language: "en-US",
//         }
//         const commands = [
//             {
//                 command: ["* complete note"],
//                 callback: (note: string) => {
//                     SpeechRecognition.stopListening()
//                     createNote(note)
//                     resetTranscript()
//                     dispatch({type: ActionTypes.SET_CURRENT,payload:parent||"main"}) // Because if this is not right below the main parent, then it should revert to the closest parent
//                 },
//                 matchInterim: true,
//             },
//         ]
//         const {transcript, listening, resetTranscript, browserSupportsSpeechRecognition} = useSpeechRecognition({
//             commands: commands,
//         });
//         const parentPrefix = ((parent && parent !== "main")?parent+".":"")
//         useEffect(() => { // TODO: validate works right
//             if (state.current === parentPrefix+"create note") {
//                 SpeechRecognition.startListening(listenArgs)
//             }
//         }, [state.current])
//     } catch (e){
//         console.error("failed to create note dictation component: " + JSON.stringify(e))
//         return null
//     }
// }
//
// // TODO: USE THIS!
// export function CreateTransferDictaphone({submit,deleteLastNote,setDstId,setTransferReason}:{
//     submit:()=>void,
//     deleteLastNote:()=>void,
//     setDstId:(id:string)=>void,
//     setTransferReason:(id:string)=>void,
// }){
//     // Always created in a state that is not listening by default
//     try {
//         const rfidCtx = useRfidReaderContext()
//         const {state, dispatch} = useDictationContext()
//         const listenArgs = {
//             continuous: false,
//             interimResults: false, // TODO: UNSURE IF WE WANT THIS OR NOT
//             language: "en-US",
//         }
//         const commands = [
//             {
//                 command: ["scan destination"],
//                 callback: () => {
//                     SpeechRecognition.stopListening()
//                     resetTranscript()
//                     ReadTagFunc(rfidCtx.dispatch, undefined, rfidCtx.state.selected)
//                         .then((idRead)=>{
//                             setDstId(idRead) // TODO: validate working
//                             SpeechRecognition.startListening(listenArgs)
//                         })
//                         .catch(e => {
//                             console.error("failed to read linking tag: " + JSON.stringify(e))
//                             SpeechRecognition.startListening(listenArgs)
//                         })
//                 },
//                 matchInterim: true,
//             },
//             {
//                 command: ["* is the transfer reason"], // TODO: EW!
//                 callback: (arg:string) => {
//                     SpeechRecognition.stopListening()
//                     resetTranscript()
//                     setTransferReason(arg) // TODO: validate working
//                     SpeechRecognition.startListening(listenArgs)
//                 },
//                 matchInterim: true,
//             },
//             {
//                 command: ["list transfer reason options"], // TODO: EW!
//                 callback: () => {
//                     // TODO: THIS!
//                 },
//                 matchInterim: true,
//             },
//             // TODO: add notes (change to "create transfer.create note" in dictation context)
//             {
//                 command: ["delete last note"],
//                 callback: () => {
//                     SpeechRecognition.stopListening()
//                     resetTranscript()
//                     deleteLastNote()// TODO: THIS!
//                     SpeechRecognition.startListening(listenArgs)
//                 },
//                 matchInterim: true,
//             },
//             { // TODO: "with note * submit transfer" ?
//                 command: ["submit current transfer"],
//                 callback: () => {
//                     SpeechRecognition.stopListening()
//                     resetTranscript()
//                     submit()
//                     dispatch({type: ActionTypes.SET_CURRENT, payload:"main"}) // main is parent of transfer
//                 },
//                 matchInterim: true,
//             },
//         ]
//         const {transcript, listening, resetTranscript, browserSupportsSpeechRecognition} = useSpeechRecognition({
//             commands: commands,
//         });
//         useEffect(() => { // TODO: validate works right
//             if (state.current === "create transfer") {
//                 SpeechRecognition.startListening(listenArgs)
//             }
//         }, [state.current])
//     } catch (e){
//         console.error("failed to create transfer dictation component: " + JSON.stringify(e))
//         return null
//     }
// }

// export function SayString(toDictate: string) {
//     DictateString(toDictate)
// }
//
// export function DictateString(toDictate: string) { // TODO: USE!
//     if ('speechSynthesis' in window) {
//         window.speechSynthesis.speak(new SpeechSynthesisUtterance(toDictate))
//     } else {
//         throw "client speech synthesis not currently available"
//     }
// }
export interface PopupInfo {
    header: string
    text?: string
    isErr: boolean // TODO: use this!
}

export function PopupApp({info}: { info:PopupInfo }) {
    const [isOpen, setIsOpen] = useState(false);
    const [doneWithFirstLoad, setDoneWithFirstLoad] = useState<boolean>(false);
    const [data, setData] = useState<PopupInfo>(info);
    useEffect(() => {
        if(doneWithFirstLoad){
            setData(info)
            setIsOpen(true)
        }else{
            setDoneWithFirstLoad(true);
        }
    }, [info]);
    const close = (e:React.MouseEvent<HTMLButtonElement, MouseEvent>)=>{
        e.stopPropagation()
        e.preventDefault()
        setIsOpen(false)
    }
    if (isOpen) {
        return <div className={"popupModal"}>
            <div className={"popupModalContent"}>
                <h3>{data.header}</h3>
                <p className={data.isErr?"error":""}>{data.text || "no text set, you should never see this message"}</p>
                <button className={"basicButton buttonFullWidth"} onClick={close}>{"Close"}</button>
            </div>
        </div>
    }
    return null
}

export const DefaultPopupInfo: PopupInfo = {
    header: "initial header",
    text: "you should never see this text",
    isErr: false,
}
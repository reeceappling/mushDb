'use client'

import React, {JSX, useContext, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AllEntries} from "@/app/components/formSubcomponents/shared";
import ID, {IdPageLink} from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {TransferData} from "@/app/components/transferServer";
import {
    EntryLinkWrapper,
    EntryLinkIdWrapper
} from "@/app/components/formSubcomponents/entryLink";
import {ImageLocationFor} from "@/app/components/formSubcomponents/picWithNotes";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequestMultipart, DoUpdateRequest,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    MainCollectionInputOrRead,
    NewColumn,
    NewEntryFormWrapper,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalSimpleKey, RequiredKey, setFormData
} from "@/app/components/common";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {useQuery} from "@tanstack/react-query";
import {SelectorFor} from "@/app/components/selector";
import TestAndValidate from "@/app/components/testing/untested";
import {
    AclDisplay,
    IsValidAcl,
    MarshalAcl,
    TogglableAreaWithDepth,
    UnmarshalAcl
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {DepthContext, DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import {GetFilterSizes, GetTransferReasons} from "@/app/components/formSubcomponents/server";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import Image from "next/image";
// TODO: list not working
// TODO: ensure display is working and looks good

export function AssertTransfer(input: any): asserts input is TransferData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['from', 'string'],
        ['to', 'string'],
        ['fromType', 'string'],
        ['toType', 'string'],
        ['creationDate', 'number'],
        ['reason', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Transfer assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // // complex required array keys
    // let complexRequiredArrayKeys = new Map<string, (v: any) => boolean>([
    //     ['from', IsString],
    // ])
    // for (let [key, validator] of complexRequiredArrayKeys) {
    //     if (!RequiredArrayOfType(key, input, validator)) {
    //         throw new Error('Transfer assertion failure: optional array key ' + key + ' was not valid');
    //     }
    // }
    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['fromImage', 'string'],
        ['toImage', 'string'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Transfer assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Transfer assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Transfer assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export default function TransferDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput<TransferData>) {
        const [initial, setInitial] = useState(data)

        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL>(initial.acl)
        const updateInitial = (updated: TransferData) => {
            setInitial(updated)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
            setErr(undefined)
        }

        // TODO: RESET BUTTON???
        const fromToLink = (preText: string, itemType: string, itemId: string,) => {
            const b58id = itemId
            return <div className={"fromToLink"}>
                <EntryLinkIdWrapper props={{linkId: b58id, entryType: itemType, openInNewTab: false}}>
                    <div className={"xferEntryLink"}>{preText + ": " + itemType + " " + b58id}</div>
                </EntryLinkIdWrapper>
            </div>
        }
        const imageArea = (alt: string, loc?: string) => {
            return <div className={"fromToImage"}>
                {/*{loc ? <Image src={ImageLocationFor(loc)} alt={"fromImage"}/> : "No " + alt + " present"/* TODO: if not working, switch back*!/*/}
                {loc ? <img src={ImageLocationFor(loc)} alt={"fromImage"}/> : "No " + alt + " present"}
            </div>
        }
        const fromToArea = () => {
            return <div className={"fromToArea"}>
                {fromToLink("From", initial.fromType, initial.from)}
                {fromToLink("To", initial.toType, initial.to)}
                {imageArea("fromImage", initial.fromImage)}
                {imageArea("toImage", initial.toImage)}
            </div>
        }
        const cookies = useContext(CookiesContext)
        const transferSubmit = () => {
            const body: any = {
                notes: notes,
                acl: MarshalAcl(acl),
            }
            DoUpdateRequest("transfer",initial._id, body, AssertTransfer, allCookies(cookies))
                .then(v=>{
                    updateInitial(new TransferData(v))
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                })
        }
        const b58idMain = initial._id
        return <DisplayFormWrapper entryType={"transfer"} id={"transferDisplay"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <ID props={{id:initial._id, txt:"Transfer", entryType:"transfer", linkPage:false, allowOpenMainPage:false}}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <DateArea pre={"Created: "} when={initial.creationDate} readonly={true}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <div>{"Reason: " + initial.reason}</div>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            {fromToArea()}
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay initial={acl} readonly={readonly} updateParent={setAcl} />
            </TogglableAreaWithDepth>
            {readonly ? null : <div>
                <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    transferSubmit()
                }}>{"Update"}</button>
            </div>}
        </DisplayFormWrapper>
}

export function NewTransferArea({idFrom, typeFrom, validTypesTo, onCreated, disposeAfter}: { // TODO: use validTypesTo?
    idFrom: string,
    typeFrom: string,
    validTypesTo: string[],
    onCreated: (xfer: TransferData) => void,
    disposeAfter?: boolean, // nil is user choice (default false)
}) {
    const [isOpen, setIsOpen] = useState(false)

    const [idTo, setIdTo] = useState<string | undefined>()
    const [picFrom, setPicFrom] = useState<File | undefined>()
    const [picTo, setPicTo] = useState<File | undefined>()
    const [notes, setNotes] = useState<Note[]>([])
    const [reason, setReason] = useState<string | undefined>()
    const [dispose, setDispose] = useState<boolean>(disposeAfter || false)
    const [err, setErr] = useState<string | undefined>()

    const cookies = useContext(CookiesContext)
    const submitNewTransfer = () => {
        if (!idFrom || idFrom === "") {
            setErr("ID From cannot be blank!")
            return
        }
        if (!idTo || idTo === "") {
            setErr("ID To cannot be blank!")
            return
        }
        if (!reason || reason === "") {
            setErr("reason cannot be blank!")
            return
        }
        const formData = new FormData();
        const dataObj: any = {
            from: idFrom,
            to: idTo,
            reason: reason,
            // optional
            fromType: typeFrom,
            notes: notes,
            disposeParent: disposeAfter || dispose,
        }
        setFormData(formData, dataObj)
        picFrom && formData.set('picFrom', picFrom, 'picFrom')
        picTo && formData.set('picTo', picTo, 'picTo')
        DoCreateRequestMultipart("transfer", formData, AssertTransfer, allCookies(cookies))
            .then(v=>{
                onCreated ? onCreated(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    const toggleOpen = () => {
        setIsOpen(!isOpen)
    }
    if (!isOpen) {
        return <div className={"pushToBottom"}>
            <button className={"buttonFullWidth greenButton"} onClick={toggleOpen}>{"Create New Transfer"}</button>
        </div>
    }
    const newTransferNotifArea = () => {
        return <div>
            <div>
                <button className={"basicButton"} onClick={(e)=>{
                    e.stopPropagation();
                    toggleOpen();
                }}>{"Close Transfer Creator"}</button>
            </div>
        </div>
    }
    const xferImgArea = (txt: string, update: (f: File | undefined) => void, className: string) => {
        return <div className={className}>
            <div>{txt}</div>
            <div>
                <ImageSelector updateParent={update}/>
            </div>
        </div>
    }

    return <NewEntryFormWrapper entryType={"transfer"}>{/* TODO: overhaul styling? */}
        {newTransferNotifArea()}
        <ErrorDisplay err={err}/>
        <div className={"newTransferRow3"}>
            <div className={"reason-to"}>
                <div>{"Transfer Reason: "}</div>
                <TransferReasonSelector onSelect={setReason}/>
            </div>
            <div className={"id-to"}>
                <MainCollectionInputOrRead onIdSelected={setIdTo} label={"ID TO: "}/>
            </div>
        </div>
        <div className={"newTransferRow2"}>
            {xferImgArea("Image from:", setPicFrom, "image-from")}
            {xferImgArea("Image to:", setPicTo, "image-to")}
        </div>
        <div className={"new-xfer-notes gapTop"}>
            <NewEntryNotes setNotes={setNotes}/>
        </div>
        <div className={"newTransferRow5"}>{/*TODO: changed! handle styling!*/}
            <div className={"submitNewXfer"}>
                <button className={"buttonSmall greenButton"} onClick={() => { // TODO: ensure classes ok
                    submitNewTransfer()
                }}>{"Submit"}</button>
            </div>
            <div className={"inlineChildren"}>{/*TODO: new! handle styling!*/}
                {disposeAfter ? <div></div> : <>
                    <div>{"Dispose?"}</div>
                    <input type="checkbox" checked={dispose} onChange={e=>{setDispose(!dispose)}}/>
                </>}
            </div>
            <div className={"cancelNewXfer"}>
                <button className={"basicButtonSmall"} onClick={toggleOpen}>{"Cancel"}</button>
            </div>
        </div>
    </NewEntryFormWrapper>
}

export function NewTransferAreaNew({idFrom, typeFrom, validTypesTo, onCreated}: {
    idFrom: string,
    typeFrom: string,
    validTypesTo: string[],
    onCreated: (xfer: TransferData) => void,
}) {
    const [isOpen, setIsOpen] = useState(false)

    const [idTo, setIdTo] = useState<string | undefined>()
    const [picFrom, setPicFrom] = useState<File | undefined>()
    const [picTo, setPicTo] = useState<File | undefined>()
    const [notes, setNotes] = useState<Note[]>([])
    const [reason, setReason] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const cookies = useContext(CookiesContext)
    const submitNewTransfer = () => {
        if (!idFrom || idFrom === "") {
            setErr("ID From cannot be blank!")
            return
        }
        if (!idTo || idTo === "") {
            setErr("ID To cannot be blank!")
            return
        }
        if (!reason || reason === "") {
            setErr("reason cannot be blank!")
            return
        }
        const formData = new FormData();
        const dataObj: any = {
            from: idFrom,
            to: idTo,
            reason: reason,
            // optional
            fromType: typeFrom,
            notes: notes,
        }
        setFormData(formData, dataObj)
        picFrom && formData.set('picFrom', picFrom, 'picFrom')
        picTo && formData.set('picTo', picTo, 'picTo')
        // Send request
        DoCreateRequestMultipart("transfer", formData, AssertTransfer, allCookies(cookies))
            .then(v=>{
                onCreated ? onCreated(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    const toggleOpen = () => {
        setIsOpen(!isOpen)
    }
    if (!isOpen) {
        return <div>
            <button className={"buttonFullWidth greenButton"} onClick={toggleOpen}>{"Create New Transfer"}</button>
        </div>
    }
    const newTransferNotifArea = () => {
        return <div>
            <div>
                <button className={"basicButton"}>{"Close Transfer Creator"}</button>
            </div>
        </div>
    }
    const xferImgArea = (txt: string, update: (f: File | undefined) => void, className: string) => {
        return <div className={className}>
            <div>{txt}</div>
            <div>
                <ImageSelector updateParent={update}/>
            </div>
        </div>
    }

    return <>
        {newTransferNotifArea()}
        <></>
        <NewEntryFormWrapper entryType={"transfer"}>{/* TODO: overhaul styling? */}
            <ErrorDisplay err={err}/>

            <div className={"newTransferRow3"}>
                <div className={"reason-to"}>
                    <div>{"Transfer Reason: "}</div>
                    <TransferReasonSelector onSelect={setReason}/>
                </div>
                <div className={"id-to"}>
                    <MainCollectionInputOrRead onIdSelected={setIdTo} label={"ID TO: "}/>
                </div>
            </div>
            <div className={"newTransferRow2"}>
                {xferImgArea("Image from:", setPicFrom, "image-from")}
                {xferImgArea("Image to:", setPicTo, "image-to")}
            </div>
            <div className={"new-xfer-notes gapTop"}>
                <TestAndValidate todos={["notes should properly be populated"]}>
                    <NewEntryNotes setNotes={setNotes}/>
                </TestAndValidate>
            </div>
            <div className={"newTransferRow5"}>
                <div className={"submitNewXfer"}>
                    <button className={"greenButton buttonFullWidth"} onClick={(e) => {
                        e.stopPropagation();
                        submitNewTransfer()
                    }}>{"Submit"}</button>
                </div>
                <div className={"cancelNewXfer"}>
                    <button onClick={toggleOpen}>{"Cancel"}</button>
                </div>
            </div>
        </NewEntryFormWrapper>
    </>
}

export function AddToTransfers(set: (s: string[]) => void, current: string[]) {
    return (newXfer: TransferData) => {
        set([...current, newXfer._id])
    }
}

export function TransfersOutDisplay( // TODO: likely overhaul
    {
        thisId,
        thisEntryType,
        transfersOut,
        allowNewTransferCreation,
        headerTxt,
        validTypesTo,
        disposeAfter,
    }: {
        thisId: string,
        thisEntryType: string,
        transfersOut?: string[],
        allowNewTransferCreation: boolean,
        headerTxt?: string,
        validTypesTo?: string[],
        disposeAfter?: boolean, // undefined == let user select (default false), true is yes, false is no
    }) {
    const openInNewTab = false
    if (!allowNewTransferCreation) {
        return <TransfersOutViewOnlyDisplay transfersOut={transfersOut} headerTxt={headerTxt}/>
    }
    const depth = useContext(DepthContext)
    const validNewXferTypes = validTypesTo || ["bag", "jar", "lc", "plate", "slant", "stasisTube"]
    const [xfers, setXfers] = useState<string[]>(transfersOut || [])
    const [resultsHidden, setResultsHidden] = useState(false)
    const [newXfers, setNewXfers] = useState<string[]>([])
    useEffect(() => {
        const spreadXfers = [...(transfersOut || [])]
        setXfers(spreadXfers)
        setNewXfers([])
    }, [transfersOut])
    if (resultsHidden) {
        return <div className={"centerH"}>
            <button className={"basicButtonSmall"} onClick={() => {
                setResultsHidden(!resultsHidden)
            }}>{"Show Transfers"}</button>
        </div>
    }

    return <DepthProvider>
        <div className={"subForm transfersOutForm depth" + depth}>
            {headerTxt && <div className={"transferHeader"}><div className={"text-xl"}>{headerTxt}</div></div>}
            <div className={"transfersGrid"}>{/*</div>}*/}
                <div className={"centerH"}>
                    <button className={"basicButtonSmall"} onClick={() => {
                        setResultsHidden(true)
                    }}>{"Hide Transfers"}</button>
                </div>
                <div>{/* Placeholder for header*/}</div>
                <div>
                    <div>{"Existing:"}</div>
                    {!resultsHidden && <div>
                        {xfers.map((xfer, i) => {
                            return <div className={"existingTransferItem"} key={xfer + i}><EntryLinkIdWrapper props={{
                                linkId: xfer,
                                entryType: "transfer",
                                openInNewTab: openInNewTab,
                            }}>{xfer}</EntryLinkIdWrapper></div>
                        })}
                        {newXfers.map((xfer, i) => {
                            return <div className={"existingTransferItem"} key={xfer + i}><EntryLinkIdWrapper props={{
                                linkId: xfer,
                                entryType: "transfer",
                                openInNewTab: openInNewTab,
                            }}>{xfer}</EntryLinkIdWrapper></div>
                        })}
                    </div>}
                </div>

                <div id={"transferCreator"} className={"mt-2"}>{/* TODO: make button float to bottom???? */}
                    {allowNewTransferCreation &&
                        <NewTransferArea idFrom={thisId} typeFrom={thisEntryType} validTypesTo={validNewXferTypes}
                                         onCreated={(newXfer: TransferData) => {
                                             setNewXfers([...newXfers, newXfer._id])
                                         }} disposeAfter={disposeAfter}/>}
                </div>
            </div>
        </div>
    </DepthProvider>
}

export function TransfersOutViewOnlyDisplay(
    {
        transfersOut,
        headerTxt,
    }: {
        transfersOut?: string[],
        headerTxt?: string,
    }) {
    const depth = useContext(DepthContext)
    const openInNewTab = false // TODO: ???
    if (!transfersOut){
        return null
    }
    return <DepthProvider>
        {headerTxt && <div className={"transferHeader"}><div className={"text-xl"}>{headerTxt}</div></div>}
        <div className={"transfersOutViewOnlyForm depth" + depth}>
            {transfersOut.map((xfer, i) => {
                return <div className={"existingTransferItem"} key={xfer + i}><EntryLinkIdWrapper props={{
                    linkId: xfer,
                    entryType: "transfer",
                    openInNewTab: false,
                }}>{xfer}</EntryLinkIdWrapper></div>
            })}
        </div>
    </DepthProvider>
}

export function InnocDisplay(
    {innoc, openInNewTab}: {
        innoc?: string,
        openInNewTab?: boolean
    }
) {
    const out: JSX.Element | null = (innoc === undefined) ? null :
        <IdPageLink id={innoc} entryType={"transfer"} openInNewTab={true}/>
    return <div className={"innocDisplay"}>
        <div>{"Innoculation ID: "}</div>
        <div>{out || "none"}</div>
    </div>
}

export function TransferReasonSelector(
    {current, onSelect}: {
        current?: string,
        onSelect?: (ab?: string) => void
    }) {
    const {isPending, error, data} = useQuery({
        queryKey: ['transferReasonOptions'],
        queryFn: GetTransferReasons,
    })
    if (isPending || error !== null) {
        return <div>{isPending ? "TRANSFER REASON SELECTOR LOADING" : "TRANSFER REASON SELECTOR ERROR: " + error.message}</div>
    }
    return <SelectorFor disabled={onSelect === undefined} options={["", ...data.keys()]} initial={current || ""}
                        updateParent={(s) => {
                            if (s === "") {
                                onSelect && onSelect(undefined)
                            }
                            onSelect && onSelect(s as string)
                        }
                        }/>
}

export function TransferListPageTable({data, onClick, withLink}: ListPageItems<TransferData>) {
    let cols: ListTableColumn<TransferData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Date", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Src", (v)=>{
            return <EntryLinkIdWrapper props={{linkId:v.from,entryType:v.fromType,openInNewTab:true}}>
                <div>{v.from}</div>
            </EntryLinkIdWrapper>
        }),
        NewColumn("Dst", (v)=>{
            return <EntryLinkIdWrapper props={{linkId:v.to,entryType:v.toType,openInNewTab:true}}>
                <div>{v.to}</div>
            </EntryLinkIdWrapper>
        }),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Reason", v=>v.reason),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: TransferData)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new TransferData(v)}}/>
}

export function TransferSelectorTable({data, onClick, withLink}: ListPageItems<TransferData>) {
    const cols: ListTableColumn<TransferData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Date", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Src", (v)=>{
            return <EntryLinkIdWrapper props={{linkId:v.from,entryType:v.fromType,openInNewTab:true}}>
                <div>{v.from}</div>
            </EntryLinkIdWrapper>
        }),
        NewColumn("Dst", (v)=>{
            return <EntryLinkIdWrapper props={{linkId:v.to,entryType:v.toType,openInNewTab:true}}>
                <div>{v.to}</div>
            </EntryLinkIdWrapper>
        }),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Link", (v: TransferData)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })
    ]
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new TransferData(v)}}/>
}

// TODO: likely will not be used. Consider delete
export function TransferSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: TransferData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: TransferData[]):JSX.Element=>{
        return <TransferSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"transfer"} entryTypes={"transfers"} doSelect={doSelect} asserter={AssertTransfer}
                                   table={table}>
    </ExistingRecentSelector>
}

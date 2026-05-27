'use client'

import React, {JSX, useContext, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note} from "@/app/components/formSubcomponents/notes";
import {AllEntries} from "@/app/components/formSubcomponents/shared";
import ID, {IdPageLink} from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {TransferData} from "@/app/components/transferServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {ImageLocationFor} from "@/app/components/formSubcomponents/picWithNotes";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {
    DisplayInput,
    HandleJsonResponse, ListPageItems,
    MainCollectionInputOrRead,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    SendMultipartRequest
} from "@/app/components/common";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {useQuery} from "@tanstack/react-query";
import {SelectorFor} from "@/app/components/selector";
import TestAndValidate from "@/app/components/testing/untested";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {DisplayFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {
    FlexedArea,
    FlexedSinglesGroup, ListPageTable,
    ListTableColumn,
    NewColumn,
    NotesFormArea, NumberToDateStr
} from "@/app/components/agarBatchClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
import {DepthContext, DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import {ExistingRecentSelector} from "@/app/components/agarRecipeClient";
import {GetTransferReasons} from "@/app/components/formSubcomponents/server";
// TODO: list not working
// TODO: ensure display is working and looks good

export function AssertTransfer(input: any): asserts input is TransferData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['from', 'string'],
        ['to', 'string'],
        ['fromType', 'string'],
        ['toType', 'string'],
        ['creationDate', 'number'],
        ['reason', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
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
    let optionalSimpleKeys = new Map<string, string>([
        ['fromImage', 'string'],
        ['toImage', 'string'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Transfer assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Transfer assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Transfer assertion failure: optional array key ' + key + ' was not valid');
        }
    }

    return
}

export default function TransferDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput) {
    try {
        AssertTransfer(data)
        const [initial, setInitial] = useState(data)

        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const updateInitial = (updated: TransferData) => {
            setInitial(updated)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
        }

        // TODO: RESET BUTTON???
        const fromToLink = (preText: string, itemType: string, itemId: string,) => {
            const b58id = itemId
            return <div className={"fromToLink"}>
                <EntryLinkWrapper props={{linkId: b58id, entryType: itemType, openInNewTab: false}}>
                    <div className={"xferEntryLink"}>{preText + ": " + itemType + " " + b58id}</div>
                </EntryLinkWrapper>
            </div>
        }
        const imageArea = (alt: string, loc?: string) => {
            return <div className={"fromToImage"}>
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
        const transferSubmit = () => {
            fetch(BaseExternalUrl + "/db/update/transfer", {
                method: 'Post',
                body: JSON.stringify({
                    notes: notes,
                    acl: MarshalAcl(acl),
                }),
                headers: {
                    credentials: 'include',
                    'Content-type': "application/json"
                },
            })
                .then(HandleJsonResponse)
                .then((newEntry) => {
                    AssertTransfer(newEntry)
                    updateInitial(newEntry)
                })
                .catch((er) => {
                    setErr(JSON.stringify(er))
                });
        }
        const b58idMain = initial._id
        return <DisplayFormWrapper entryType={"transfer"} id={"transferDisplay"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <ID id={b58idMain} txt={"Transfer"} entryType={"transfer"} allowOpenMainPage={false} linkPage={false}/>
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
                <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
            </TogglableAreaWithDepth>
            {readonly ? null : <div>
                <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    transferSubmit()
                }}>{"Update"}</button>
            </div>}
            {/* TODO: unlikely to need: <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/> TODO: where to put?*/}
        </DisplayFormWrapper>
    } catch (err) {
        return <div>{"ERROR: Transfer data format incorrect: " + err}</div>
    }
}

export function NewTransferArea({idFrom, typeFrom, validTypesTo, onCreated, cookies}: {
    idFrom: string,
    typeFrom: string,
    validTypesTo: string[],
    onCreated: (xfer: TransferData) => void,
    cookies: string,
}) {
    const [isOpen, setIsOpen] = useState(false)

    const [idTo, setIdTo] = useState<string | undefined>()
    const [picFrom, setPicFrom] = useState<File | undefined>()
    const [picTo, setPicTo] = useState<File | undefined>()
    const [notes, setNotes] = useState<Note[]>([])
    const [reason, setReason] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
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
        let formData = new FormData();
        let dataObj: any = {
            from: idFrom,
            to: idTo,
            fromType: typeFrom,
            reason: reason,
            notes: notes,
        }
        formData.set('data', JSON.stringify(dataObj))
        picFrom && formData.set('picFrom', picFrom, 'picFrom')
        picTo && formData.set('picTo', picTo, 'picTo')
        // Send request
        SendMultipartRequest(BaseExternalUrl + "/db/create/transfer", cookies, formData)
            .then(HandleJsonResponse)
            .then((newEntry) => {
                AssertTransfer(newEntry)
                onCreated && onCreated(newEntry)
            })
            .catch((er) => {
                setErr(JSON.stringify(er))
            });
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
                {/* TODO: THIS THING??? */}
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

    return <NewEntryFormWrapper entryType={"transfer"}>{/* TODO: overhaul styling */}
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
        <div className={"newTransferRow5"}>
            <div className={"submitNewXfer"}>
                <button className={"greenButtonSmall"} onClick={() => {
                    submitNewTransfer()
                }}>{"Submit"}</button>
            </div>
            <div className={"cancelNewXfer"}>
                <button className={"basicButtonSmall"} onClick={toggleOpen}>{"Cancel"}</button>
            </div>
        </div>
    </NewEntryFormWrapper>
}

export function NewTransferAreaNew({idFrom, typeFrom, validTypesTo, onCreated, cookies}: {
    idFrom: string,
    typeFrom: string,
    validTypesTo: string[],
    onCreated: (xfer: TransferData) => void,
    cookies: string,
}) {
    const [isOpen, setIsOpen] = useState(false)

    const [idTo, setIdTo] = useState<string | undefined>()
    const [picFrom, setPicFrom] = useState<File | undefined>()
    const [picTo, setPicTo] = useState<File | undefined>()
    const [notes, setNotes] = useState<Note[]>([])
    const [reason, setReason] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
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
        let formData = new FormData();
        let dataObj: any = {
            from: idFrom,
            to: idTo,
            fromType: typeFrom,
            reason: reason,
            notes: notes,
        }
        formData.set('data', JSON.stringify(dataObj))
        picFrom && formData.set('picFrom', picFrom, 'picFrom')
        picTo && formData.set('picTo', picTo, 'picTo')
        // Send request
        SendMultipartRequest(BaseExternalUrl + "/db/create/transfer", cookies, formData)
            .then(HandleJsonResponse)
            .then((newEntry) => {
                AssertTransfer(newEntry)
                onCreated && onCreated(newEntry)
            })
            .catch((er) => {
                setErr(JSON.stringify(er))
            });
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
                {/* TODO: THIS THING??? */}
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
        <NewEntryFormWrapper entryType={"transfer"}>{/* TODO: overhaul styling */}
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

export function TransfersOutDisplay(
    {
        thisId,
        thisEntryType,
        transfersOut,
        allowNewTransferCreation,
        headerTxt,
        validTypesTo,
        cookies,
    }: {
        thisId: string,
        thisEntryType: string,
        transfersOut?: string[],
        allowNewTransferCreation: boolean,
        headerTxt?: string,
        validTypesTo?: string[],
        cookies: string,
    }) {
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
                    {!resultsHidden && <div>{/* TODO: classes?*/}
                        {xfers.map((xfer, i) => { // TODO: FLEXBOX?
                            return <div className={"existingTransferItem"} key={xfer + i}><EntryLinkWrapper props={{
                                linkId: xfer,
                                entryType: "transfer",
                                openInNewTab: false, // TODO: ok?
                            }}>{xfer}</EntryLinkWrapper></div>
                        })}
                        {newXfers.map((xfer, i) => { // TODO: FLEXBOX?
                            return <div className={"existingTransferItem"} key={xfer + i}><EntryLinkWrapper props={{
                                linkId: xfer,
                                entryType: "transfer",
                                openInNewTab: false, // TODO: ok?
                            }}>{xfer}</EntryLinkWrapper></div>
                        })}
                    </div>}
                </div>

                <div id={"transferCreator"} className={"mt-2"}>{/* TODO: make button float to bottom???? */}
                    {allowNewTransferCreation &&
                        <NewTransferArea idFrom={thisId} typeFrom={thisEntryType} validTypesTo={validNewXferTypes}
                                         onCreated={(newXfer: TransferData) => {
                                             setNewXfers([...newXfers, newXfer._id])
                                         }} cookies={cookies}/>}
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
    if (!transfersOut){
        return null
    }
    return <DepthProvider>
        {headerTxt && <div className={"transferHeader"}><div className={"text-xl"}>{headerTxt}</div></div>}
        <div className={"transfersOutViewOnlyForm depth" + depth}>
            {transfersOut.map((xfer, i) => { // TODO: FLEXBOX for inline?
                return <div className={"existingTransferItem"} key={xfer + i}><EntryLinkWrapper props={{
                    linkId: xfer,
                    entryType: "transfer",
                    openInNewTab: false, // TODO: ok?
                }}>{xfer}</EntryLinkWrapper></div>
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
    let out: JSX.Element | null = (innoc === undefined) ? null :
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
        // async () => {
        //     try {
        //         const reasonsMap = await GetTransferReasons();
        //         return reasonsMap;
        //     } catch (e) {
        //         throw e;
        //     }
        // },
        /*
        return fetch(BaseInternalUrl + "/options/transferReasons").then(HandleJsonResponse).then((resJson) => {
                return convertObjectToStringMap(resJson)
            }).catch((e) => {
                throw e
            })
        */
    })
    if (isPending || error !== null) {
        return <div>{isPending ? "TRANSFER REASON SELECTOR LOADING" : "TRANSFER REASON SELECTOR ERROR: " + error.message}</div>
    }
    return <SelectorFor disabled={onSelect === undefined} options={["", ...data.keys()]} initial={current || ""}
                        updateParent={(s) => {
                            if (s === "") {
                                onSelect && onSelect()
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
            return <EntryLinkWrapper props={{linkId:v.from,entryType:v.fromType,openInNewTab:true}}>
                <div>{v.from}</div>
            </EntryLinkWrapper>
        }),
        NewColumn("Dst", (v)=>{
            return <EntryLinkWrapper props={{linkId:v.to,entryType:v.toType,openInNewTab:true}}>
                <div>{v.to}</div>
            </EntryLinkWrapper>
        }),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Reason", v=>v.reason),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: TransferData)=>{
            return <EntryLinkWrapper props={{linkId:encodeURI(v._id),entryType:"transfer",openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    // TODO: expansion for everything else????
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}

export function TransferSelectorTable({data, onClick, withLink}: ListPageItems<TransferData>) {
    let cols: ListTableColumn<TransferData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Date", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Src", (v)=>{
            return <EntryLinkWrapper props={{linkId:v.from,entryType:v.fromType,openInNewTab:true}}>
                <div>{v.from}</div>
            </EntryLinkWrapper>
        }),
        NewColumn("Dst", (v)=>{
            return <EntryLinkWrapper props={{linkId:v.to,entryType:v.toType,openInNewTab:true}}>
                <div>{v.to}</div>
            </EntryLinkWrapper>
        }),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Link", (v: TransferData)=>{
            return <EntryLinkWrapper props={{linkId:encodeURI(v._id),entryType:"transfer",openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })
    ]
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}

// TODO: likely will not be used
export function TransferSelector( // TODO: USE ELSEWHERE
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
        {/* TODO: ok? allowCreate && <NewTransferForm handlers={{onCreate: doSelect,isTopLevel: false}}/>*/}
    </ExistingRecentSelector>
}

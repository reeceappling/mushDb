'use client'

import React, {JSX, useContext, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note} from "@/app/components/formSubcomponents/notes";
import {AllEntries} from "@/app/components/formSubcomponents/shared";
import ID, {IdPageLink} from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {TransferData} from "@/app/components/transferServer";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {ImageLocationFor} from "@/app/components/formSubcomponents/picWithNotes";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {
    DisplayInput,
    HandleJsonResponse,
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
import {AclDisplay, IsValidAcl, MarshalAcl} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {DisplayFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {FlexedArea, FlexedSinglesGroup, NotesFormArea} from "@/app/components/agarBatchClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
import {DepthContext, DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";

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
            <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
            {readonly ? null : <div><input type="submit" value="Update" onClick={transferSubmit} onSubmit={(e) => {
                e.preventDefault();
            }}/></div>}
            {/* TODO: unlikely to need: <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/> TODO: where to put?*/}
        </DisplayFormWrapper>
    } catch (err) {
        return <div>{"ERROR: Transfer data format incorrect: " + err}</div>
    }
}

// export function NewTransferAreaGivenTo({idFrom, typeFrom, validTypesFrom, onTransferCreated}: {
//     idFrom: string[],
//     typeFrom: string[],
//     validTypesFrom: string[],
//     onTransferCreated: (xfer: TransferData) => void
// }){
//     // TODO: do this!
//     // TODO: Use this!
// }

// TODO: ENSURE ALL NEW TRANSFERS ADOPT MOST STRINGENT PERMS AND ONLY WORKS IF THE USER CAN WRITE TO THE PARENT
// TODO: ENSURE idFrom and typeFrom are PROPERLY ARRAYS
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
    const submitNewTransfer = () => { // TODO: CHANGE TO NEW FORMAT!!!!!
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
        }
        if (notes !== undefined && notes.length > 0) {
            dataObj.notes = notes
        }
        formData.set('data', JSON.stringify(dataObj))
        picFrom && formData.set('picFrom', picFrom, 'picFrom')
        picTo && formData.set('picTo', picTo, 'picTo')
        // Send request
        console.log("sending multipart transfer request") // TODO: THIS!
        SendMultipartRequest(BaseExternalUrl + "/db/create/transfer", cookies, formData)
            .then(HandleJsonResponse)
            .then((newEntry) => {
                AssertTransfer(newEntry)
                console.log("returning new transfer entry to component for listing") // TODO: THIS!
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
            <button className={"basicButtonSmall"} onClick={toggleOpen}>{"Create New Transfer"}</button>
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

    return <NewEntryFormWrapper entryType={"transfer"}>{/* TODO: overhaul styling */}
        <ErrorDisplay err={err}/>
        {newTransferNotifArea()}
        {/* TODO: OVERHAUL THE TRANFER CREATOR TOP ROW */}
        <div className={"newTransferRow3"}>
            <div className={"reason-to"}>
                <div>{"Transfer Reason: "}</div><TransferReasonSelector onSelect={setReason}/>
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
                <button onClick={() => {
                    submitNewTransfer()
                }}>{"Submit"}</button>
            </div>
            <div className={"cancelNewXfer"}>
                <button onClick={toggleOpen}>{"Cancel"}</button>
            </div>
        </div>
    </NewEntryFormWrapper>
}

export function AddToTransfers(set: (s: string[]) => void, current: string[]) {
    return (newXfer: TransferData) => {
        set([...current, newXfer._id])
    }
}

// TODO: TRANSFERS OUT DISPLAY WITH MULTI-FROMs

export function TransfersOutDisplay( // TODO: OVERHAUL!
    {
        thisId, // TODO: BINARY, RIGHT?
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
    if (resultsHidden){
        return <div className={"centerH"}><button onClick={()=>{setResultsHidden(!resultsHidden)}}>{"Show Transfers"}</button></div>
    }

    return <DepthProvider>
        <div className={"subForm transfersOutForm depth" + depth}>
            {headerTxt && <div className={"centerH"}>{headerTxt}</div>}
            <div className={"centerH"}><button onClick={()=>{setResultsHidden(true)}}>{"Hide Transfers"}</button></div>
            {!resultsHidden && <div className={"transfersOutFlex"}>{/* TODO: ensure closeable! Also split up the old and new transfers */}
                {newXfers.concat(xfers)?.map((xfer) => { // TODO: FLEXBOX? Ensure concat is ok here
                    return <EntryLinkWrapper props={{
                        linkId: xfer,
                        entryType: "transfer",
                        openInNewTab: false, // TODO: ok?
                    }}>{xfer}</EntryLinkWrapper> // TODO: wrap div if needed
                })}
            </div>}

            <div className={"centerH"}>
                {allowNewTransferCreation &&
                    <NewTransferArea idFrom={thisId} typeFrom={thisEntryType} validTypesTo={validNewXferTypes}
                                     onCreated={(newXfer: TransferData) => {
                                         setNewXfers([...newXfers, newXfer._id])
                                     }} cookies={cookies}/>}
            </div>
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
        <IdPageLink id={innoc} entryType={"transfer"}/>
    return <div className={"innocDisplay"}>
        <div>{"Innoculation ID: "}</div>
        <div>{out||"none"}</div>
    </div>
}

export function TransferReasonSelector(
    {current, onSelect}: {
        current?: string,
        onSelect?: (ab?: string) => void
    }) {
    const {isPending, error, data} = useQuery({
        queryKey: ['transferReasonOptions'], // TODO: FIX THIS!!!!!
        queryFn: () => {
            // TODO: delete lines before fetch for the real server
            const map = new Map<string, string>();
            map.set("outgrew", "outgrew plate");
            map.set("contaminated", "parent was contaminated");
            map.set("sectoring", "transferring a specific sector");
            return map;
            // TODO: reenable
            fetch(BaseExternalUrl + "/options/transferReasons").then(HandleJsonResponse).then((resJson) => {
                return convertObjectToStringMap(resJson)
            }).catch((e) => {
                throw e
            })
        },
    })
    if (isPending || error !== null) {
        return <div>{isPending ? "TRANSFER REASON SELECTOR LOADING" : "TRANSFER REASON SELECTOR ERROR: " + error.message}</div>
    }
    // TODO: Does this selector need more depth???
    // TODO: maybe do selector a little differently because this one is a map?
    return <SelectorFor disabled={onSelect === undefined} options={["", ...data.keys()]} initial={current || ""}
                        updateParent={(s) => {
                            if (s === "") {
                                onSelect && onSelect()
                            }
                            onSelect && onSelect(s as string)
                        }
                        }/>
}

function convertObjectToStringMap(obj: { [key: string]: string }): Map<string, string> {
    const map = new Map<string, any>();
    for (const key in obj) {
        if (Object.prototype.hasOwnProperty.call(obj, key)) {
            map.set(key, obj[key]);
        }
    }
    return map;
}

// TODO: fix!
// export function TransferListDisplay({data, onClick}: SingleListProps<TransferData>) {
//     return <div>
//         {data.map((b,i)=>{
//             return <TransferInline data={b} onClick={()=>{onClick(b)}} key={i}/>
//         })}
//     </div>
// }
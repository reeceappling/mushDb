'use client'

import React, {useEffect, useState} from "react";
import NotesAreaOld, {IsValidNote, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import {AddCreatedQuadColFunction, AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    DisplayInput,
    HandleJsonResponse,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea, ListPageItems,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalSimpleKey,
} from "@/app/components/common";
import ReaderWriterSelector from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {DisposedDisplay, ErrorDisplay, OpenMainPage} from "@/app/components/formSubcomponents/commonClient";
import {InitialNotesState,} from "@/app/components/formSubcomponents/contaminations";
import {BaseExternalUrl} from "@/app/components/Constants";
import {WaterJarData} from "@/app/components/waterJarServer";
import {PcRunData, RecentPCRunSelector} from "@/app/components/pcRunServer";
import {OnViewCreatorsTriColArea} from "@/app/components/pcRunClient";
import {CreatedLinkFor} from "@/app/components/substrateRecipeClient";
import {NewMssForm, RecentSelectorV2} from "@/app/components/mssClient";
import {MssData} from "@/app/components/mssServer";
import {DisplayFormWrapper, NewEntryFormWrapper} from "./lcRecipeClient";
import { InlineEntry } from "./agarRecipeClient";
import {
    FlexedArea,
    FlexedSinglesGroup, ListPageTable,
    ListTableColumn,
    NewColumn,
    NotesFormArea, NumberToDateStr
} from "@/app/components/agarBatchClient";
import {CreatedUpdatedDisposedArea} from "@/app/components/plateClient";
import {SubstrateBatchData} from "@/app/components/substrateBatchServer";

export function AssertWaterJar(input: any): asserts input is WaterJarData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['pcRun', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('WJ assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['disposed', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('WJ assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('WJ assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export default function WaterJarDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }: DisplayInput) {
    const [initial, setInitial] = useState(data as WaterJarData)
    // TODO: updateInitial!!!!!
    const [disposed, setDisposed] = useState<number | undefined>(data.disposed)
    const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
    // Helper states
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const updateInitial = (updated: WaterJarData) => {
        setInitial(updated)
        setDisposed(updated.disposed)
        setNotes(InitialNotesState(updated.notes))
    }

    const submit = () => {
        if (notes.new.length === 0 && notes.existing === InitialNotesState(initial.notes).existing) { // TODO: ensure ok
            setErr("No changes found")
            return
        }
        fetch(BaseExternalUrl + "/db/update/waterJar/" + initial._id, { // This ID is in base58 // TODO: ID IS NOT PROPERLY POPULATING
            method: 'Post',
            body: JSON.stringify({
                notes: notes,
                disposed: disposed,
                writeTagTo: writeTagTo
            }),
            headers: {
                credentials: 'include',
                'Content-type': 'application/json'
            },
        })
            .then(HandleJsonResponse)
            .then((newEntry) => {
                AssertWaterJar(newEntry)
                updateInitial(newEntry)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    const ovcs: OnViewCreatorQuadCol[] = [
        {
            txt: "New MultiSpore Syringe",
            newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                return <NewMssForm waterJarIn={initial} handlers={{
                    onCreate: (item: MssData) => {
                        onCreate([{
                            typeText: "MultiSpore Syringe",
                            node: <CreatedLinkFor linkId={item._id} typ={"mss"}/>,
                        }])
                    }, isTopLevel: false,
                }}/>
            }
        }
        // TODO: NEW MSS
        // TODO: NEW STASIS TUBE (Should only be done this way if the stasis tube goes through the PC empty)
        //OvcForXfers(data._id, "bag", ["bag","fruitingChamber","jar","plate","slant","stasisTube"]), // TODO: ensure list correct
    ] // TODO: THIS!
    return (
        <DisplayFormWrapper entryType={"waterJar"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <ID txt={"Water Jar"} id={initial._id} entryType={"waterJar"} />
            <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/> {/*TODO: where to put?*/}
            <FlexedArea>
                <FlexedSinglesGroup>
                    <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated} readonly={readonly}
                                                disposed={initial.disposed} setDisposedOnParent={setDisposed}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            {readonly || <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>}
            {readonly || <button className={"bottomButton greenButton"} onClick={(e)=>{
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}
        </DisplayFormWrapper>
    )
}

// export function CreatedLastUpdatedArea( // TODO: MOVE
//     {created, updated}: { created: number, updated: number }
// ) {
//     return <div className={"createdUpdatedArea fullWidth"}>
//         <DateArea pre={"Created: "} when={created} readonly={true}/>
//         <DateArea pre={"Updated: "} when={updated} readonly={true}/>
//     </div>
// }

export function NewWaterJarForm(
    {handlers, pcRunIn}: { handlers: NewEntryInput<WaterJarData>, pcRunIn?: PcRunData }
) {
    const [pcRun, setPCRun] = useState(pcRunIn)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    // TODO: handle isTopLevel
    const createJar = (e: React.MouseEvent) => {
        e.preventDefault()
        if (pcRun === undefined) {
            setErr("An pc run must be selected")
            return
        }
        let body: any = {
            pcRun: pcRun._id,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        fetch(BaseExternalUrl + "/create/waterJar", { // TODO: ensure correct
            method: "POST",
            headers: {
                credentials: 'include',
                'Content-type': 'application/json'
            },
            body: JSON.stringify(body)
        })
            .then(HandleJsonResponse)
            .then((entry) => {
                AssertWaterJar(entry)
                handlers.onCreate && handlers.onCreate(entry)
            })
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
    }
    return <NewEntryFormWrapper entryType={"waterJar"}>
        <ErrorDisplay err={err}/>
        {pcRunIn !== undefined && <RecentPCRunSelector doSelect={setPCRun} allowCreation={true} creatorInPage={true}/>}
        {/* TODO: NOTES AREA */}
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"bottomButton"} onClick={createJar}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function WaterJarInline({data, expandByDefault, onClick, showMainPageButton, idIsLink}: InlineProps<WaterJarData>) {
    const [expanded, setExpanded] = useState(expandByDefault)
    const b58id = data._id
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={b58id} txt={"Water Jar"} entryType={"waterJar"} allowOpenMainPage={showMainPageButton} linkPage={idIsLink}/>
            <DateArea pre={"Created: "} when={data.creationDate} readonly={true}/>
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            {/* TODO: FIX ME! <PcRunArea binaryId={data.pcRun} readonly={true} headerLevel={1}/>*/}
            <NotesAreaInline notes={data.notes} offset={-1}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                                                     expanded={expanded}/>
    </InlineEntry>
}

// TODO: HEAVILY TEST!!!!
export function WaterJarRecentSelector({onSelect}: { onSelect: (selected?: WaterJarData) => void }) {
    return <RecentSelectorV2<WaterJarData> listUrlType={"waterJars"} assertion={AssertWaterJar}
                                           singleConstructor={(val, i) => {
                                               return <WaterJarInline data={val} expandByDefault={false}
                                                                      onClick={onSelect}/>
                                           }}/>
}

export function WaterJarListPageTable({data, onClick}: ListPageItems<WaterJarData>) {
    const cols: ListTableColumn<WaterJarData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("PcRun", (v)=>v.pcRun),
        NewColumn("Disposed", (v)=>{
            return v.disposed?NumberToDateStr(v.disposed):""
        }),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    // TODO: expansion for everything else????
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}

'use client'

import React, {JSX, useState} from "react";
import {IsValidNote, NewEntryNotes, Note} from "@/app/components/formSubcomponents/notes";
import {AddCreatedQuadColFunction, AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import {
    DisplayInput,
    HandleJsonResponse,
    ListPageItems,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalSimpleKey,
} from "@/app/components/common";
import ReaderWriterSelector, {
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {InitialNotesState,} from "@/app/components/formSubcomponents/contaminations";
import {BaseExternalUrl} from "@/app/components/Constants";
import {WaterJarData} from "@/app/components/waterJarServer";
import {PcRunData, PcRunSelectorCloseable} from "@/app/components/pcRunServer";
import {OnViewCreatorsTriColArea} from "@/app/components/pcRunClient";
import {CreatedLinkFor} from "@/app/components/substrateRecipeClient";
import {NewMssForm} from "@/app/components/mssClient";
import {MssData} from "@/app/components/mssServer";
import {DisplayFormWrapper, NewEntryFormWrapper} from "./lcRecipeClient";
import {ExistingRecentSelector} from "./agarRecipeClient";
import {
    FlexedArea,
    FlexedSinglesGroup, ListPageTable,
    ListTableColumn,
    NewColumn,
    NotesFormArea, NumberToDateStr
} from "@/app/components/agarBatchClient";
import {CreatedUpdatedDisposedArea} from "@/app/components/plateClient";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {SelectorWrapper} from "@/app/components/lcClient";
import {NewStasisTubeForm} from "@/app/components/stasisTubeClient";

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
        { // TODO: TEST HEAVILY
            txt: "New MultiSpore Syringe",
            newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                return <NewMssForm waterJarIn={initial} handlers={{
                    onCreate: (item: MssData) => {
                        onCreate([{
                            typeText: "MultiSpore Syringe",
                            node: <CreatedLinkFor linkId={item._id} typ={"mss"}/>,
                        }], false)
                    }, isTopLevel: false,
                }}/>
            },
        },
        WriteRfidOvcArea(initial._id),
        // Stasis tubes must be PC'd with water in them, so we don't have a creator here
    ]
    return (
        <DisplayFormWrapper entryType={"waterJar"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <ID txt={"Water Jar"} id={id} entryType={"waterJar"} />
            <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>
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
        <SelectorWrapper title={"PC Run"} nameFunc={(v:PcRunData):string=>{
            return v._id
        }} current={pcRun}>{/* TODO: validate working */}
            <PcRunSelectorCloseable doSelect={setPCRun} allowCreation={true} creatorInPage={true}/>
        </SelectorWrapper>
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton buttonFullWidth"} onClick={createJar}>{"Create"}</button>
    </NewEntryFormWrapper>
}

// export function WaterJarInline({data, expandByDefault, onClick, showMainPageButton, idIsLink}: InlineProps<WaterJarData>) {
//     const [expanded, setExpanded] = useState(expandByDefault)
//     const b58id = data._id
//     return <InlineEntry onClick={onClick}>
//         <InlineSubArea props={{}}>
//             <ID id={b58id} txt={"Water Jar"} entryType={"waterJar"} allowOpenMainPage={showMainPageButton} linkPage={idIsLink}/>
//             <DateArea pre={"Created: "} when={data.creationDate} readonly={true}/>
//         </InlineSubArea>
//         <InlineExpansionArea props={{expanded: expanded}}>
//             {/* TODO: FIX ME! <PcRunArea binaryId={data.pcRun} readonly={true} headerLevel={1}/>*/}
//             <NotesAreaInline notes={data.notes} offset={-1}/>
//             <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
//         </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
//                                                      expanded={expanded}/>
//     </InlineEntry>
// }

export function WaterJarListPageTable({data, onClick, withLink}: ListPageItems<WaterJarData>) {
    let cols: ListTableColumn<WaterJarData>[] = [
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
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: WaterJarData)=>{
            return <EntryLinkWrapper props={{linkId:encodeURI(v._id),entryType:"waterJar",openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    // TODO: expansion for everything else????
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}
export function WaterJarSelectorTable({data, onClick}: ListPageItems<WaterJarData>) {
    let cols: ListTableColumn<WaterJarData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Disposed", (v)=>{
            return v.disposed?NumberToDateStr(v.disposed):""
        }),
        NewColumn("Link", (v: WaterJarData)=>{
            return <EntryLinkWrapper props={{linkId:encodeURI(v._id),entryType:"waterJar",openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })
    ]
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}

export function WaterJarSelector( // TODO: USE ELSEWHERE
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: WaterJarData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: WaterJarData[]):JSX.Element=>{
        return <WaterJarSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"waterJar"} entryTypes={"waterJars"} doSelect={doSelect} asserter={AssertWaterJar}
                                   table={table}>
        {allowCreate && <NewWaterJarForm handlers={{onCreate: doSelect,isTopLevel: false}}/>}
    </ExistingRecentSelector>
}

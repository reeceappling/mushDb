'use client'

import React, {JSX, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AddCreatedQuadColFunction, AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import {
    createApiUrlFor,
    CreatedLinkFor, DisplayFormWrapper,
    DisplayInput, ErrHandler, ExistingRecentSelector, FlexedArea, FlexedSinglesGroup,
    HandleJsonResponse,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn, NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType,
    OptionalSimpleKey, SelectorWrapper, updateApiUrlFor,
} from "@/app/components/common";
import ReaderWriterSelector, {
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {InitialNotesState,} from "@/app/components/formSubcomponents/contaminations";
import {WaterJarData} from "@/app/components/waterJarServer";
import {PcRunData, PcRunSelectorCloseable} from "@/app/components/pcRunServer";
import {NewMssForm} from "@/app/components/mssClient";
import {MssData} from "@/app/components/mssServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";

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
        fetch(updateApiUrlFor("waterJar",initial._id), { // TODO: ID IS NOT PROPERLY POPULATING
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
            .catch(ErrHandler(setErr));
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
            <ID txt={"Water Jar"} id={initial._id} entryType={"waterJar"} />
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
        fetch(createApiUrlFor("waterJar"), {
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
            .catch(ErrHandler(setErr));
    }
    return <NewEntryFormWrapper entryType={"waterJar"}>
        <ErrorDisplay err={err}/>
        <SelectorWrapper title={"PC Run"} nameFunc={(v:PcRunData):string=>{
            return v._id
        }} current={pcRun}>
            <PcRunSelectorCloseable doSelect={setPCRun} allowCreation={true} creatorInPage={true}/>{/*TODO: handlers.isTopLevel vs just true*/}
        </SelectorWrapper>
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton buttonFullWidth"} onClick={createJar}>{"Create"}</button>
    </NewEntryFormWrapper>
}

// TODO: WATER JAR IMPORT?

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

export function WaterJarSelector(
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

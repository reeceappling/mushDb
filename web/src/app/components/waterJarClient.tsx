'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AddCreatedQuadColFunction, AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import {
    CreatedLinkFor,
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoImportRequest,
    DoUpdateRequest,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    ImportDisplayInput,
    ImportEntryFormWrapper,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalSimpleKey,
    RequiredKey,
} from "@/app/components/common";
import ReaderWriterSelector, {
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {WaterJarData} from "@/app/components/waterJarServer";
import {PcRunData, PcRunSelectorCloseable} from "@/app/components/pcRunServer";
import {NewMssForm} from "@/app/components/mssClient";
import {MssData} from "@/app/components/mssServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {MarshalAcl, UnmarshalAcl} from "@/app/components/accessControlClient";
import DateArea from "@/app/components/formSubcomponents/date";
import {ActionTypes, useModalContext} from "@/app/components/formSubcomponents/modalContext/modal";
import {JarRecipeData} from "@/app/components/jarRecipeServer";

export function AssertWaterJar(input: any): asserts input is WaterJarData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['pcRun', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('WJ assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['disposed', 'number'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('WJ assertion failure: optional key ' + key + ' was not valid');
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
            throw new Error('WJ assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export default function WaterJarDisplay(
    {
        readonly, data, headerLevel, isTopLevel
    }: DisplayInput<WaterJarData>) {
    const {dispatch} = useModalContext();
    const [initial, setInitial] = useState(data as WaterJarData)
    const [disposed, setDisposed] = useState<number | undefined>(data.disposed)
    const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
    // Helper states
    const [acl, setAcl] = useState(initial.acl)
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const updateInitial = (updated: WaterJarData) => {
        setInitial(updated)
        setDisposed(updated.disposed)
        setNotes(InitialNotesState(updated.notes))
        setErr(undefined)
    }
    const cookies = useContext(CookiesContext)
    const submit = () => {
        if (notes.new.length === 0 && notes.existing === InitialNotesState(initial.notes).existing) { // TODO: ensure ok
            setErr("No changes found")
            return
        }
        const body: any = {
            notes: notes,
            disposed: disposed,
            acl: MarshalAcl(acl)
        }
        DoUpdateRequest("waterJar", initial._id, body, AssertWaterJar, allCookies(cookies))
            .then(v => {
                updateInitial(new WaterJarData(v))
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Update Success",
                        text: "entry updated successfully",
                        isErr: false
                    }})
            })
            .catch(e => {
                setErr(JSON.stringify(e))
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Update Failed",
                        text: "failed to update: " + JSON.stringify(e),
                        isErr: true
                    }})
            })
    }
    const ovcs: ()=>OnViewCreatorQuadCol[] = ()=> {
        const disp = initial.disposed !== undefined
        return !disp ? [
            // TODO: new stasis tube? probably not
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

        ]:[]
    }

    return (
        <DisplayFormWrapper entryType={"waterJar"}>
            <ErrorDisplay err={err}/>
            <ID props={{
                id: initial._id,
                txt: "Water Jar",
                entryType: "waterJar",
                linkPage: false,
                allowOpenMainPage: false
            }}/>
            <OnViewCreatorsTriColArea OnViewCreators={ovcs()} readonly={readonly}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated}
                                                readonly={readonly}
                                                initialDisposed={initial.disposed} setDisposedOnParent={setDisposed}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            {readonly || <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}
        </DisplayFormWrapper>
    )
}

export function NewWaterJarForm(
    {handlers, pcRunIn}: { handlers: NewEntryInput<WaterJarData>, pcRunIn?: PcRunData }
) {
    const {dispatch} = useModalContext();
    const [pcRun, setPCRun] = useState(pcRunIn)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const cookies = useContext(CookiesContext)
    const createJar = (e: React.MouseEvent) => {
        e.preventDefault()
        if (pcRun === undefined) {
            setErr("An pc run must be selected")
            return
        }
        const body: any = {
            pcRun: pcRun._id,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        DoCreateRequest("waterJar", body, AssertWaterJar, allCookies(cookies))
            .then(v => {
                if(handlers.onCreate!==undefined){
                    handlers.onCreate(new WaterJarData(v))
                    handlers.isTopLevel && dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                            header: "Create Success",
                            text: "entry created successfully",
                            isErr: false
                        }})
                } else {
                    console.log("no onCreate provided")
                }
            })
            .catch(e => {
                const newErr = "entry failed to create: " + JSON.stringify(e)
                setErr(newErr)
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Create Failure",
                        text: newErr,
                        isErr: true
                    }})
            })
    }
    return <NewEntryFormWrapper entryType={"waterJar"}>
        <ErrorDisplay err={err}/>
        <PcRunSelectorCloseable doSelect={setPCRun} allowCreation={true} creatorInPage={true}/>
        {/*</SelectorWrapper>*/}
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton buttonFullWidth"} onClick={createJar}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function WaterJarImportDisplay({headerLevel}: ImportDisplayInput) {
    const {dispatch} = useModalContext();
    const [created, setCreated] = useState<number>(Date.now())
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const cookies = useContext(CookiesContext)
    const ImportWaterJar = () => {
        const body = {
            creationDate: created,
            notes: notes,
            writeTagTo: writeTagTo,
        }

        DoImportRequest(body, "waterJar", AssertWaterJar, setErr, allCookies(cookies))
        // TODO: on error dispatch the modal stuff!
    }
    return <ImportEntryFormWrapper entryType={"slant"}>
        <ErrorDisplay err={err}/>
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={ImportWaterJar}>{"Import Water Jar"}</button>
    </ImportEntryFormWrapper>
}

export function WaterJarListPageTable({data, onClick, withLink}: ListPageItems<WaterJarData>) {
    let cols: ListTableColumn<WaterJarData>[] = [
        NewColumn("ID", (v) => v._id, true),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }, true),
        NewColumn("PcRun", (v) => v.pcRun, true),
        NewColumn("Disposed", (v) => {
            return v.disposed ? NumberToDateStr(v.disposed) : ""
        }, true),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }), // TODO: fit?
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: WaterJarData) => {
            return <EntryLinkWrapper props={{entry: v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v => {
        return new WaterJarData(v)
    }}/>
}

export function WaterJarSelectorTable({data, onClick}: ListPageItems<WaterJarData>) {
    const cols: ListTableColumn<WaterJarData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Disposed", (v) => {
            return v.disposed ? NumberToDateStr(v.disposed) : ""
        }),
        NewColumn("Link", (v: WaterJarData) => {
            return <EntryLinkWrapper props={{entry: v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })
    ]
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v => {
        return new WaterJarData(v)
    }}/>
}

export function WaterJarSelector(
    {
        doSelect,
        allowCreate,
        hideDisposed
    }: {
        doSelect: (val: WaterJarData | undefined) => void,
        allowCreate?: boolean,
        hideDisposed?:boolean
    }) {
    const table = (items: WaterJarData[]): JSX.Element => {
        return <WaterJarSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"waterJar"} entryTypes={"waterJars"} doSelect={doSelect}
                                   asserter={AssertWaterJar}
                                   table={table} hideDisposed={hideDisposed}>
        {allowCreate && <NewWaterJarForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingRecentSelector>
}

'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import DateArea from "@/app/components/formSubcomponents/date";
import {LcData} from "@/app/components/lcServer";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import {
    clientPostRequestHeaders,
    ConfirmedCleanArea,
    ConfirmedCleanSelector, createApiUrlFor, DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoUpdateRequest, ErrHandler, ExistingRecentSelector, FlexedArea, FlexedSinglesGroup,
    HandleJsonResponse, importApiUrlFor, ImportEntryFormWrapper,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn, NewEntryFormWrapper, NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey, updateApiUrlFor, viewUrlFor,
} from "@/app/components/common";
import ReaderWriterSelector, {
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {redirect} from "next/navigation";
import {
    ErrorDisplay,
    GensFormDisplay,
    ParentDisplay,
} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import ID from "@/app/components/formSubcomponents/id";
import {ExistingSpeciesSelector, SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {LcSyringe} from "@/app/components/lcSyringeServer";
import {AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import {TransfersOutDisplay} from "@/app/components/transferClient";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import TestAndValidate from "@/app/components/testing/untested";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {AssertLcRecipe} from "@/app/components/lcRecipeClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";

export function AssertLcSyringe(input: any): asserts input is LcSyringe {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['species', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Lc syringe assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['parent', 'string'],
        ['sale', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['confirmedClean', 'boolean'],
        ['knownFruitable', 'boolean'],
        ['disposed', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Lc syringe assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Jar assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Lc syringe assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export function LcSyringeImportDisplay() {
    const cookies = useContext(CookiesContext)
    const [created, setCreated] = useState<number>(Date.now())
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>(undefined)
    const [confirmedClean, setConfirmedClean] = useState<boolean | undefined>(undefined)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(undefined)
    const [generation, setGeneration] = useState<number | undefined>(undefined)
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const ImportLcSyringe = () => {
        if (species === undefined) {
            setErr("Species must be set!")
            return
        }
        let dataObj: any = {
            creationDate: created,
            species: species._id,
            subspecies: subspecies?._id,
            confirmedClean: confirmedClean,
            knownFruitable: knownFruitable,
            generation: generation,
            writeTagTo: writeTagTo,
        }

        fetch(importApiUrlFor("lcSyringe"), {
            method: 'Post',
            body: JSON.stringify(dataObj),
            headers: clientPostRequestHeaders,
        })
            .then(HandleJsonResponse)
            .then((newLcSyringe) => {
                AssertLcSyringe(newLcSyringe)
                redirect(viewUrlFor("lcSyringe", newLcSyringe._id))
            })
            .catch(ErrHandler(setErr));
    }
    return <ImportEntryFormWrapper entryType={"lcSyringe"}>
        {err != undefined && <div>{"Error: " + err}</div>}
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <ExistingSpeciesSelector doSelect={setSpecies}/>
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}/>
        <ConfirmedCleanSelector updateParent={setConfirmedClean}/>
        <KnownFruitableArea doSelect={setKnownFruitable}/>
        <GenerationInput updateParent={setGeneration}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton bottomButton"} onClick={ImportLcSyringe}>{"Submit"}</button>
    </ImportEntryFormWrapper>
}

export default function LcSyringeDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput) {
    try {
        AssertLcSyringe(data)
    } catch (err) {
        return <div>{"ERROR: Liquid Culture Syringe data format incorrect: " + JSON.stringify(err)}</div>
    }
    const [transfersOut, setTransfersOut] = useState<string[]>(data.transfersOut || [])
    const [confirmedClean, setConfirmedClean] = useState<boolean | undefined>(data.confirmedClean)
    const [knownFruitable, setKnownFruitable] = useState(data.knownFruitable)
    const [disposed, setDisposed] = useState(data.disposed)
    const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [acl, setAcl] = useState<ACL | undefined>(data.acl)
    const [err, setErr] = useState<string | undefined>()
    // TODO: THIS WHOLE FUNC???
    const [initial, setInitial] = useState(data)
    const updateInitial = (updated: LcSyringe) => {
        setInitial(updated)
        setTransfersOut(updated.transfersOut || [])
        setConfirmedClean(updated.confirmedClean)
        setKnownFruitable(updated.knownFruitable)
        setDisposed(updated.disposed)
        setNotes(InitialNotesState(updated.notes))
        setAcl(updated.acl)
    }

    const cookies = useContext(CookiesContext)
    const lcSyringeSubmit = () => {
        let body: any = { // TODO: ensure ok
            confirmedClean: confirmedClean,
            knownFruitable: knownFruitable,
            disposed: disposed,
            notes: notes,
            writeTagTo: writeTagTo,
            acl: MarshalAcl(acl),
        }
        DoUpdateRequest("lcSyringe",initial._id, body, AssertLcSyringe, allCookies(cookies))
            .then(updateInitial)
            .catch(ErrHandler(setErr))

        // fetch(updateApiUrlFor("lcSyringe",initial._id), {
        //     method: 'Post',
        //     body: JSON.stringify(body),
        //     headers: clientPostRequestHeaders,
        // })
        //     .then(HandleJsonResponse)
        //     .then((updatedEntry) => {
        //         AssertLcSyringe(updatedEntry)
        //         updateInitial(updatedEntry)
        //     })
        //     .catch(ErrHandler(setErr));
    }
    const ovcs: OnViewCreatorQuadCol[] = [
        WriteRfidOvcArea(initial._id),
    ]
    return <DisplayFormWrapper entryType={"lcSyringe"}>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <ID id={data._id} txt={"Liquid Culture Syringe"} entryType={"lcSyringe"}/>
        <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>
        <FlexedArea>
            <FlexedSinglesGroup>
                <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated}
                                            disposed={disposed} setDisposedOnParent={setDisposed} readonly={readonly}/>
            </FlexedSinglesGroup>
            <FlexedSinglesGroup>
                <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                <ParentDisplay parent={initial.parent} parentType={"lc"} headerLevel={headerLevel}/>
            </FlexedSinglesGroup>
            <FlexedSinglesGroup>
                <GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore}
                                 headerLevel={headerLevel}/>
            </FlexedSinglesGroup>
            <FlexedSinglesGroup>
                <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable} readonly={readonly}
                                    headerLevel={headerLevel}/>
                <ConfirmedCleanArea onSelect={setConfirmedClean} readonly={readonly} initial={confirmedClean}
                                    headerLevel={headerLevel}/>
            </FlexedSinglesGroup>
        </FlexedArea>
        <TransfersOutDisplay thisId={initial._id} thisEntryType={"plate"} transfersOut={transfersOut}
                             allowNewTransferCreation={!readonly}/>
        <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
        <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
            <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
        </TogglableAreaWithDepth>
        {readonly ? null : <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>}
        {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
            e.stopPropagation();
            lcSyringeSubmit()
        }}>{"Update"}</button>}
    </DisplayFormWrapper>

}

export function NewLcSyringeForm({parentLc, onCreate, txt}: {
    parentLc?: LcData,
    onCreate?: (newItem: LcSyringe) => void,
    txt: string
}) {
    // TODO: THIS WHOLE FUNC?
    const cookies = useContext(CookiesContext)
    const [itemsCreated, setItemsCreated] = useState<string[]>([])
    const [parent, setParent] = useState<LcData | undefined>(parentLc) // TODO: this ok to not call set??
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const createdItemsDiv = () => {
        if (itemsCreated.length === 0) {
            return null
        }
        return <div>
            <div>
                <div>{"Lc syringes Created:"}</div>
            </div>
            {itemsCreated.map((createdLc) => {
                const b58id = createdLc
                return <EntryLink props={{displayId: b58id, linkId: b58id, entryType: "lcSyringe", openInNewTab: false /* TODO: ok?*/}}/>
            })}
        </div>
    }
    const errHandler = ErrHandler(setErr)
    const createEntry = (e: React.MouseEvent) => {
        e.preventDefault()
        if (!parent) {
            setErr("A parent must be selected")
            return
        }
        let body: any = {
            writeTagTo: writeTagTo,
            parent: parent,
            notes: notes,
        }
        DoCreateRequest("lcSyringe", body, AssertLcSyringe, allCookies(cookies))
            .then(item=>{
                onCreate && onCreate(item)
                setItemsCreated([...itemsCreated, item._id]) // TODO: ok?
            })
            .catch(errHandler)
        // fetch(createApiUrlFor("lcSyringe"), { // TODO: del if not needed
        //     method: "POST",
        //     headers: clientPostRequestHeaders,
        //     body: JSON.stringify(body)
        // })
        //     .then(HandleJsonResponse)
        //     .then((newEntry) => {
        //         AssertLcSyringe(newEntry)
        //         onCreate && onCreate(newEntry)
        //         setItemsCreated([...itemsCreated, newEntry._id]) // TODO: ok?
        //     })
        //     .catch(ErrHandler(setErr));
    }

    return <NewEntryFormWrapper entryType={"lcSyringe"}>
        <TestAndValidate todos={["fix and test this area"]}>
            <div>{txt}</div>
            {createdItemsDiv()}
            <ErrorDisplay err={err}/>
            {!parent && <TestAndValidate todos={["SELECT LC RECIPE"]}>
                <div>{"LC SElECTOR HERE"}</div>
            </TestAndValidate>}
            <NewEntryNotes setNotes={setNotes}/>
            <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
            <button className={"greenButton"} onClick={createEntry}>{"Create"}</button>
        </TestAndValidate>
    </NewEntryFormWrapper>
}

export function LcSyringeListPageTable({data, onClick, withLink}: ListPageItems<LcSyringe>) {
    let cols: ListTableColumn<LcSyringe>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Spec", (v)=>v.species||""),
        NewColumn("Subspec", v=>v.subspecies||"" ),
        NewColumn("Clean",v=>v.confirmedClean?(v.confirmedClean?"clean":"dirty"):"?"),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: LcSyringe)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}
export function LcSyringeSelectorTable({data, onClick}: ListPageItems<LcSyringe>) {
    return <LcSyringeListPageTable data={data} onClick={onClick} withLink={true} />
}

export function LcSyringeSelector(
    {
        doSelect,
        // TODO: allowCreate
    }: {
        doSelect: (val: LcSyringe | undefined) => void,
        // TODO: allowCreate?: boolean
    }) {
    const table = (items: LcSyringe[]):JSX.Element=>{
        return <LcSyringeSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"lcSyringe"} entryTypes={"lcSyringes"} doSelect={doSelect} asserter={AssertLcSyringe}
                                   table={table}>
        {/* TODO: ok? allowCreate && <NewLcSyringeForm handlers={{onCreate: doSelect,isTopLevel: false}}/>*/}
    </ExistingRecentSelector>
}

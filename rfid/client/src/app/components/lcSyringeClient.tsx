'use client'

// TODO: THIS WHOLE FILE

import React, {useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import DateArea from "@/app/components/formSubcomponents/date";
import {LcData} from "@/app/components/lcServer";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import GenerationArea from "@/app/components/formSubcomponents/generationInput";
import {
    ConfirmedCleanArea,
    ConfirmedCleanSelector,
    DisplayInput,
    HandleJsonResponse,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea, ListPageItems,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
} from "@/app/components/common";
import ReaderWriterSelector from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {redirect} from "next/navigation";
import {
    DisposedDisplay,
    ErrorDisplay,
    GensFormDisplay,
    GensInlineDisplay,
    ParentDisplay,
    SpeciesArea,
    SubspeciesArea,
} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import ID from "@/app/components/formSubcomponents/id";
import {ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {LcSyringe} from "@/app/components/lcSyringeServer";
import {AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import {AddToTransfers, TransfersOutDisplay} from "@/app/components/transferClient";
import EntryLink from "@/app/components/formSubcomponents/entryLink";
import TestAndValidate from "@/app/components/testing/untested";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {dataFor, InlineEntry} from "@/app/components/agarRecipeClient";
import {OnViewCreatorsQuadColArea} from "@/app/components/pcRunClient";
import {DisplayFormWrapper, ImportEntryFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {
    FlexedArea,
    FlexedSinglesGroup, ListPageTable,
    ListTableColumn,
    NewColumn,
    NotesFormArea, NumberToDateStr
} from "@/app/components/agarBatchClient";
import {CreatedUpdatedDisposedArea} from "@/app/components/plateClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
import {SpeciesSubspeciesArea} from "@/app/components/lcClient";

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

export function LcSyringeImportDisplay({onImport, cookies}: { onImport: (lcs: LcSyringe) => void, cookies: string }) {
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

        // TODO: change all fetches into "SendStandardPOst()"???
        fetch(BaseExternalUrl + "/db/import/lcSyringe", {
            method: 'Post',
            body: JSON.stringify(dataObj),
            headers: {
                'Cookie': cookies, // TODO: ensure this is everywhere
                credentials: 'include',
                'Content-type': "application/json"
                //Authorization: tokenFetch,
            },
        })
            .then(HandleJsonResponse)
            .then((newLcSyringe) => {
                AssertLcSyringe(newLcSyringe)
                onImport(newLcSyringe)
                redirect(BaseExternalUrl + "/view/lcSyringe/" + newLcSyringe._id)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    return <ImportEntryFormWrapper entryType={"lcSyringe"}>
        {err != undefined && <div>{"Error: " + err}</div>}
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <ExistingSpeciesSelector doSelect={setSpecies}/>
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}/>
        <ConfirmedCleanSelector updateParent={setConfirmedClean}/>
        <KnownFruitableArea doSelect={setKnownFruitable}/>
        <GenerationArea readonly={false} updateParent={setGeneration}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={ImportLcSyringe}>{"Submit"}</button>
    </ImportEntryFormWrapper>
}

export default function LcSyringeDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
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
    // TODO: THIS WHOLE FUNC
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


    const lcSyringeSubmit = () => {
        let bodyObj: any = {
            confirmedClean: confirmedClean,
            knownFruitable: knownFruitable,
            disposed: disposed,
            notes: notes,
            writeTagTo: writeTagTo,
            acl: MarshalAcl(acl),
        }

        fetch(BaseExternalUrl + "/db/update/lcSyringe/" + initial._id, {
            method: 'Post',
            body: JSON.stringify(bodyObj),
            headers: {
                credentials: 'include',
                'Content-type': "application/json"
            },
        })
            .then(HandleJsonResponse)
            .then((updatedEntry) => {
                AssertLcSyringe(updatedEntry)
                updateInitial(updatedEntry)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    const ovcs: OnViewCreatorQuadCol[] = [
        // TODO: can lcSyr do anything else?

    ]
    return <DisplayFormWrapper entryType={"lcSyringe"}>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <ID id={data._id} txt={"Liquid Culture Syringe"} entryType={"lcSyringe"}/>
        <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
        <FlexedArea>
            <FlexedSinglesGroup>
                <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated}
                                            disposed={disposed} setDisposedOnParent={setDisposed} readonly={readonly}/>
            </FlexedSinglesGroup>
            <FlexedSinglesGroup>
                <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                {/*<SpeciesSubspeciesFormArea species={initial.species} subspecies={initial.subspecies}/>*/}
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
                             allowNewTransferCreation={!readonly}
                             cookies={cookies}/>
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

export function NewLcSyringeForm({parentLc, onCreate, cookies, txt}: {
    parentLc?: LcData,
    onCreate?: (newItem: LcSyringe) => void,
    cookies: string
    txt: string
}) {
    // TODO: THIS WHOLE FUNC
    const [itemsCreated, setItemsCreated] = useState<string[]>([])
    const [parent, setParent] = useState<LcData | undefined>(parentLc) // TODO: this ok to not call set??
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    // //const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
    const createdItemsDiv = () => {
        if (itemsCreated.length === 0) {
            return null
        }
        // TODO: STYLING SO NEW IS GREEN, newest is greener????
        return <div>
            <div>
                <div>{"Lc syringes Created:"}</div>
            </div>
            {itemsCreated.map((createdLc) => {
                const b58id = createdLc
                return <EntryLink props={{displayedId: b58id, linkId: b58id, entryType: "lcSyringe"}}>
                    <div>{b58id}</div>
                </EntryLink>
            })}
        </div>
    }
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
        fetch(BaseExternalUrl + "/create/lcSyringe", { // TODO: ensure correct
            method: "POST",
            headers: {
                credentials: 'include',
                'Cookie': cookies, // TODO: likely dont need? ensure authed?
                'Content-type': 'application/json'
            },
            body: JSON.stringify(body)
        })
            .then(HandleJsonResponse)
            .then((newEntry) => {
                try {
                    AssertLcSyringe(newEntry)
                    onCreate && onCreate(newEntry)
                    setItemsCreated([...itemsCreated, newEntry._id]) // TODO: ok?
                } catch (e) {
                    // TODO: ?????
                }
            })
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
    }

    return <NewEntryFormWrapper entryType={"lcSyringe"}>
        <TestAndValidate todos={["fix and test this area"]}>
            <div>{txt}</div>
            {/* TODO: ensure ok */}
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

export function LcSyringeInline({
                                    data,
                                    expandByDefault,
                                    onClick,
                                    showMainPageButton,
                                    idIsLink
                                }: InlineProps<LcSyringe>) {
    const [expanded, setExpanded] = useState(expandByDefault)
    const b58id = data._id
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={b58id} txt={"Liquid Culture Syringe"} entryType={"lcSyringe"} allowOpenMainPage={showMainPageButton}
                linkPage={idIsLink}/>
            {/* TODO: parent? */}
            <DateArea pre={"Created: "} when={data.creationDate} readonly={true}/>
            <SpeciesArea readonly={true} initial={data.species}/>
            <SubspeciesArea readonly={true} initialSub={data.species} currentSpecies={data.species}/>
            <KnownFruitableArea initial={data.knownFruitable} readonly={true}/>
            <ConfirmedCleanArea initial={data.confirmedClean} readonly={true}/>
            <DisposedDisplay readonly={true} disposed={data.disposed}/>
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded,}}>
            <GensInlineDisplay gensSinceSpore={data.genSpore} gensSinceFruitOrSpore={data.genFruitOrSpore} offset={-1}/>
            {/*TODO: <ProjectsArea allowCreate={false} projects={data.perms?.projectPerms.ids} readonly={true} headerLevel={headerLevel} offset={-1} allowRemove={false}/>*/}
            <NotesAreaInline notes={data.notes} offset={-1}/>
            {/* TODO: SALE? */}
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                                                     expanded={expanded}/>
    </InlineEntry>
}

// TODO: fix???
// export function LcSyringeListDisplay({data, onClick}: SingleListProps<LcSyringe>) {
//     return <div>
//         {data.map((b,i)=>{
//             return <LcSyringeInline data={b} onClick={()=>{onClick(b)}} key={i}/>
//         })}
//     </div>
// }

export function LcSyringeListPageTable({data, onClick}: ListPageItems<LcSyringe>) {
    const cols: ListTableColumn<LcSyringe>[] = [
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
    // TODO: expansion for everything else????
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}

'use client'

import React, {JSX, useContext, useState} from "react";
import {
    DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoUpdateRequest,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    ImportDisplayInput,
    ImportEntryFormWrapper,
    IsString,
    ListPageItems,
    ListPageTable,
    ListTableColumn, MultipartImportRequest,
    NewColumn,
    NewEntryFormWrapper,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalSimpleKey, RequiredKey,
    setFormData
} from "@/app/components/common";
import {
    DisposedDisplay,
    ErrorDisplay,
    ParentDisplay,
} from "@/app/components/formSubcomponents/commonClient";
import {
    IsValidNote,
    NewEntryNotes,
    Note,
    NoteEntriesGroup, NotesFormArea,
} from "@/app/components/formSubcomponents/notes";
import {SporePrintData} from "@/app/components/sporePrintServer";
import TestAndValidate from "@/app/components/testing/untested";
import {FruitData} from "@/app/components/fruitServer";
import {
    AclDisplay,
    IsValidAcl,
    MarshalAcl,
    TogglableAreaWithDepth,
    UnmarshalAcl
} from "@/app/components/accessControlClient";
import {SporeSwabData} from "@/app/components/sporeSwabServer";
import ID from "@/app/components/formSubcomponents/id";
import {ACL} from "@/app/components/accessControlServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import DateArea from "@/app/components/formSubcomponents/date";
import {ExistingSpeciesSelector, SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {SaleArea} from "@/app/components/saleClient";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import ReaderWriterSelector, {WriteRfidOvcArea} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";

// TODO: list page not working
// TODO: ensure display page doing what we want

export function AssertSporeSwab(input: any): asserts input is SporeSwabData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['species', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Plate assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['parent', 'string'],
        ['parentType', 'string'],
        ['subspecies', 'string'],
        ['sale', 'string'],
        ['disposed', 'number'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Swab assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Spore Swab assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', IsString],
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Swab assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export function SporeSwabImportDisplay({headerLevel}: ImportDisplayInput) { // TODO: USE ONLY FOR EXISTING SPORE PRINTS!
    const [swabDate, setSwabDate] = useState<number>(Date.now())
    const [notes, setNotes] = useState<Note[]>([])
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>()
    const [image, setImage] = useState<File | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const cookies = useContext(CookiesContext)
    const importEntry = (e: React.MouseEvent) => {
        e.preventDefault()
        if (!species) {
            setErr("A species must be selected")
            return
        }
        const formData = new FormData()
        const dataObj: any = {
            creationDate: swabDate,
            species: species._id,
            // optional
            subspecies: subspecies?._id,
            notes: notes,
        }
        setFormData(formData, dataObj)
        formData.set("data", JSON.stringify(dataObj))
        // Img
        if (image) {
            formData.set("img", image, "img")
        }

        MultipartImportRequest(formData, "sporeSwab", AssertSporeSwab, setErr, allCookies(cookies))
    }
    //no parent because we couldn't possibly know it
    return <ImportEntryFormWrapper entryType={"sporeSwab"}>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <DateArea pre={"Swab Date: "} readonly={false} when={Date.now()} updateParent={setSwabDate}/>
        <ExistingSpeciesSelector doSelect={setSpecies} headerLevel={headerLevel}/>
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies} headerLevel={headerLevel}/>
        <ImageSelector updateParent={setImage}/>
        <NoteEntriesGroup preexisting={false} readonly={false} updateParent={ns => {
            setNotes(ns.map(n => {
                return n.data
            }))
        }} headerLevel={headerLevel}/>
        <button className={"greenButton"} onClick={importEntry}>{"Create"}</button>
    </ImportEntryFormWrapper>

}

export default function SporeSwabDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput<SporeSwabData>) {
        const [initial, setInitial] = useState(data)

        const [sale, setSale] = useState(initial.sale)
        const [disposed, setDisposed] = useState(initial.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [acl, setAcl] = useState<ACL>(initial.acl)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: SporeSwabData) => {
            setInitial(updated)
            setSale(updated.sale)
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
            setErr(undefined)
        }
        const cookies = useContext(CookiesContext)
        const submit = () => {
            const body: any = {
                sale: sale,
                disposed: disposed,
                notes: notes,
                acl: MarshalAcl(acl),
            }
            DoUpdateRequest("sporeSwab",data._id, body, AssertSporeSwab, allCookies(cookies))
                .then(v=>{
                    updateInitial(new SporeSwabData(v))
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                })
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            // TODO: probably get rid of? OvcForXfers(data._id, "sporeSwab", ["plate", "slant", "stasisTube", "jar", "bag", "fruitingChamber"], allCookies(cookies)),
            WriteRfidOvcArea(initial._id),
        ]
        return <DisplayFormWrapper entryType={"sporeSwab"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <ID props={{id:data._id, txt:"Spore Swab", entryType:"sporeSwab", linkPage:false, allowOpenMainPage:false}}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <TestAndValidate todos={["reformat these into groups"]}>
                        <ParentDisplay parent={initial.parent} parentType={initial.parentType} />
                        <SaleArea readonly={false} canCreateSale={true} sale={sale} setSale={setSale}
                                  headerLevel={headerLevel}/>
                        <DateArea pre={"Print Date: "} readonly={true} when={initial.creationDate}/>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                        <DisposedDisplay readonly={false} disposed={disposed} setDisposedOnParent={setDisposed}/>
                        <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                       </TestAndValidate>
                </FlexedSinglesGroup>
            </FlexedArea>

            <NotesFormArea initial={initial.notes} readonly={readonly} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay initial={acl} readonly={readonly} updateParent={setAcl} />
            </TogglableAreaWithDepth>
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}
            <OnViewCreatorsTriColArea OnViewCreators={ovcs}
                                      readonly={readonly}/> {/*swab to agar and that's about it */}
        </DisplayFormWrapper>
}

// Should only be accessible from a fruit's page
export function NewSporeSwabForm(
    {printIn, fruitIn, headerLevel, offset, onCreate}: {
        printIn?: SporePrintData
        fruitIn?: FruitData
        headerLevel?: number
        offset?: number
        onCreate: (sp: SporeSwabData) => void
    }) {
    // TODO: EITHER PRINT OR FRUIT!!!!!

    // TODO: THIS!
    const [parent, setParent] = useState<string | undefined>(printIn?._id || fruitIn?._id || undefined)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)

    const [err, setErr] = useState<string | undefined>(undefined)

    const cookies = useContext(CookiesContext)
    const createEntry = (e: React.MouseEvent) => {
        e.preventDefault()
        if (!parent || parent === "") {
            setErr("Parent must be selected")
            return
        }
        const body: any = {
            parent: parent,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        DoCreateRequest("sporeSwab", body, AssertSporeSwab, allCookies(cookies))
            .then(v=>{
                onCreate ? onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }

    return <NewEntryFormWrapper entryType={"sporeSwab"}>
        <ErrorDisplay err={err} headerLevel={headerLevel} offset={offset}/>
        {(printIn || fruitIn) && <div>{"TODO: PARENT SELECTOR"/* TODO: THIS! PARENT SELECTOR IF NOT PROVIDED*/}</div>}
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={createEntry}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function SporeSwabListPageTable({data, onClick, withLink}: ListPageItems<SporeSwabData>) {
    let cols: ListTableColumn<SporeSwabData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Spec", (v)=>v.species||""),
        NewColumn("Subspec", v=>v.subspecies||"" ),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SporeSwabData)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new SporeSwabData(v)}}/>
}
export function SporeSwabSelectorTable({data, onClick}: ListPageItems<SporeSwabData>) {
    return <SporeSwabListPageTable data={data} onClick={onClick} withLink={true} />
}

export function SporeSwabSelector(
    {
        doSelect,
    }: {
        doSelect: (val: SporeSwabData | undefined) => void,
    }) {
    const table = (items: SporeSwabData[]):JSX.Element=>{
        return <SporeSwabSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"sporeSwab"} entryTypes={"sporeSwabs"} doSelect={doSelect} asserter={AssertSporeSwab}
                                   table={table}>
    </ExistingRecentSelector>
}
